package desktop

import (
	"os/exec"
	"syscall"
)

// hide keeps PowerShell from flashing a console window.
func hide(cmd *exec.Cmd) { cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
