package player

import (
	"net"
	"path/filepath"
	"time"

	"github.com/Microsoft/go-winio"
)

// mpv on Windows serves IPC on a named pipe; the temp dir name makes it unique.
func ipcPath(dir string) string { return `\\.\pipe\` + filepath.Base(dir) }

func dialIPC(path string) (net.Conn, error) {
	timeout := 200 * time.Millisecond
	return winio.DialPipe(path, &timeout)
}
