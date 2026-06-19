package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
	"time"
)

// Client is a thin HTTP client for the FormaTeX API.
type Client struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

func New(apiKey, baseURL string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: 120 * time.Second},
	}
}

// CompileRequest mirrors the /compile JSON body.
type CompileRequest struct {
	LaTeX   string `json:"latex"`
	Engine  string `json:"engine,omitempty"`
	Runs    int    `json:"runs,omitempty"`
	Timeout int    `json:"timeout,omitempty"`
}

// CompileResult holds a successful compilation response.
type CompileResult struct {
	PDF           []byte
	Engine        string
	DurationMs    int
	SizeBytes     int
	Log           string
	AIExplanation string
}

// LogError is a structured error parsed from the TeX log.
type LogError struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Message string `json:"message"`
	Context string `json:"context"`
}

// CompileFailure is returned as an error when compilation fails (HTTP 422).
type CompileFailure struct {
	Log           string
	Errors        []LogError
	AIExplanation string
}

func (e *CompileFailure) Error() string {
	if len(e.Errors) > 0 {
		return fmt.Sprintf("compilation failed: %s (line %d)", e.Errors[0].Message, e.Errors[0].Line)
	}
	return "compilation failed"
}

// ── HTTP helpers ──────────────────────────────────────────────────────────────

func (c *Client) postJSON(path string, body interface{}) ([]byte, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)
	return c.do(req)
}

func (c *Client) postMultipart(path string, fields map[string]string, fileName, mimeType string, fileData []byte) ([]byte, error) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	for k, v := range fields {
		if err := mw.WriteField(k, v); err != nil {
			return nil, err
		}
	}
	h := textproto.MIMEHeader{}
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, fileName))
	h.Set("Content-Type", mimeType)
	fw, err := mw.CreatePart(h)
	if err != nil {
		return nil, err
	}
	if _, err := fw.Write(fileData); err != nil {
		return nil, err
	}
	mw.Close()

	req, err := http.NewRequest(http.MethodPost, c.baseURL+path, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)
	return c.do(req)
}

func (c *Client) get(path string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", c.apiKey)
	return c.do(req)
}

func (c *Client) do(req *http.Request) ([]byte, error) {
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, parseAPIError(resp.StatusCode, data)
	}
	return data, nil
}

func parseAPIError(status int, body []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return fmt.Errorf("HTTP %d: %s", status, string(body))
	}
	if status == 422 {
		f := &CompileFailure{
			Log:           strField(raw, "log"),
			AIExplanation: strField(raw, "ai_explanation"),
		}
		if errs, ok := raw["errors"].([]interface{}); ok {
			for _, e := range errs {
				if m, ok := e.(map[string]interface{}); ok {
					le := LogError{
						Message: strField(m, "message"),
						File:    strField(m, "file"),
						Context: strField(m, "context"),
					}
					if line, ok := m["line"].(float64); ok {
						le.Line = int(line)
					}
					f.Errors = append(f.Errors, le)
				}
			}
		}
		return f
	}
	msg := strField(raw, "error")
	if msg == "" {
		msg = fmt.Sprintf("HTTP %d", status)
	}
	return fmt.Errorf("%s", msg)
}

func strField(m map[string]interface{}, key string) string {
	v, _ := m[key].(string)
	return v
}

// ── API methods ───────────────────────────────────────────────────────────────

// Compile compiles a LaTeX document and returns PDF bytes.
func (c *Client) Compile(req CompileRequest) (*CompileResult, error) {
	data, err := c.postJSON("/api/v1/compile", req)
	if err != nil {
		return nil, err
	}
	return parseCompileResult(data, req.Engine)
}

// CompileZip compiles a ZIP archive containing a LaTeX project.
func (c *Client) CompileZip(zipData []byte, engine, mainFile string, runs, timeout int) (*CompileResult, error) {
	fields := map[string]string{}
	if engine != "" {
		fields["engine"] = engine
	}
	if mainFile != "" {
		fields["main"] = mainFile
	}
	if runs > 0 {
		fields["runs"] = fmt.Sprintf("%d", runs)
	}
	if timeout > 0 {
		fields["timeout"] = fmt.Sprintf("%d", timeout)
	}
	data, err := c.postMultipart("/api/v1/compile/zip", fields, "archive.zip", "application/zip", zipData)
	if err != nil {
		return nil, err
	}
	return parseCompileResult(data, engine)
}

func parseCompileResult(data []byte, defaultEngine string) (*CompileResult, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("unexpected response: %w", err)
	}
	pdfB64 := strField(raw, "pdf")
	if pdfB64 == "" {
		return nil, fmt.Errorf("no PDF in response")
	}
	pdfBytes, err := base64.StdEncoding.DecodeString(pdfB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode PDF: %w", err)
	}
	res := &CompileResult{
		PDF:           pdfBytes,
		Engine:        strField(raw, "engine"),
		Log:           strField(raw, "log"),
		AIExplanation: strField(raw, "ai_explanation"),
	}
	if res.Engine == "" {
		res.Engine = defaultEngine
	}
	if d, ok := raw["duration"].(float64); ok {
		res.DurationMs = int(d)
	}
	if s, ok := raw["sizeBytes"].(float64); ok {
		res.SizeBytes = int(s)
	}
	return res, nil
}

// Convert converts LaTeX to another format (docx, html, odt, epub, markdown, txt).
// Returns raw output bytes.
func (c *Client) Convert(latex, format string) ([]byte, error) {
	payload, err := json.Marshal(map[string]string{"latex": latex, "format": format})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/v1/convert", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)
	return c.do(req)
}

// Format formats a LaTeX document using latexindent. Returns formatted source and duration.
func (c *Client) Format(latex string) (string, int, error) {
	data, err := c.postJSON("/api/v1/format", map[string]string{"latex": latex})
	if err != nil {
		return "", 0, err
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return "", 0, err
	}
	dur := 0
	if d, ok := raw["durationMs"].(float64); ok {
		dur = int(d)
	}
	return strField(raw, "formatted"), dur, nil
}

// WhoAmI calls GET /api/v1/me to verify the API key is valid.
func (c *Client) WhoAmI() (map[string]interface{}, error) {
	data, err := c.get("/api/v1/me")
	if err != nil {
		return nil, err
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}
