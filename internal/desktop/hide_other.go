//go:build !windows

package desktop

import "os/exec"

func hide(*exec.Cmd) {}
