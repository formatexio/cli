package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
)

// openFile opens a file with the OS default application.
func openFile(path string) {
	var err error
	switch runtime.GOOS {
	case "darwin":
		err = exec.Command("open", path).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", path).Start()
	default:
		err = exec.Command("xdg-open", path).Start()
	}
	if err != nil {
		fmt.Printf("note: could not open %s automatically\n", path)
	}
}
