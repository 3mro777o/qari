//go:build !windows

package player

import (
	"net"
	"path/filepath"
)

func ipcPath(dir string) string { return filepath.Join(dir, "mpv.sock") }

func dialIPC(path string) (net.Conn, error) { return net.Dial("unix", path) }
