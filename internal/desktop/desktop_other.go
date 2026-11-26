//go:build !linux

package desktop

import (
	"os/exec"
	"runtime"
	"strings"
)

// Desktop outside Linux only shows a notification when a new surah starts;
// media keys need per-OS APIs (Now Playing, SMTC) that qari does not have.
type Desktop struct {
	Cmds   chan Cmd
	notify bool
	title  string
}

func Start(notify bool) *Desktop { return &Desktop{Cmds: make(chan Cmd), notify: notify} }

func (d *Desktop) Set(s State) {
	if d == nil || s.Title == d.title {
		return
	}
	d.title = s.Title
	if d.notify && s.Title != "" {
		go notify(s.Title, s.Artist)
	}
}

func (d *Desktop) Close() {}

func notify(title, body string) {
	switch runtime.GOOS {
	case "darwin":
		q := func(s string) string { return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(s) + `"` }
		exec.Command("osascript", "-e", "display notification "+q(body)+" with title "+q(title)).Run()
	case "windows":
		// The texts go in through environment variables, never into the script.
		cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", winToast)
		cmd.Env = append(cmd.Environ(), "QARI_TITLE="+title, "QARI_BODY="+body)
		hide(cmd)
		cmd.Run()
	}
}

const winToast = `[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] > $null
$x = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText02)
$t = $x.GetElementsByTagName('text')
$t.Item(0).AppendChild($x.CreateTextNode($env:QARI_TITLE)) > $null
$t.Item(1).AppendChild($x.CreateTextNode($env:QARI_BODY)) > $null
$app = '{1AC14E77-02E7-4E5D-B744-2EB1AE5198B7}\WindowsPowerShell\v1.0\powershell.exe'
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier($app).Show([Windows.UI.Notifications.ToastNotification]::new($x))`
