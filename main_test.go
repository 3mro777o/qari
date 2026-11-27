package main

import (
	"errors"
	"flag"
	"strings"
	"testing"
	"time"

	"github.com/3mro777o/qari/internal/store"
	"github.com/3mro777o/qari/internal/ui"
)

func TestParseArgs(t *testing.T) {
	for _, c := range []struct {
		args  []string
		pos   string
		theme string
		noUI  bool
	}{
		{[]string{"--theme", "lapis", "play", "alafasi", "18"}, "play alafasi 18", "lapis", false},
		{[]string{"play", "alafasi", "--no-ui", "18", "--theme=lapis"}, "play alafasi 18", "lapis", true},
		{[]string{"play", "--", "-x", "--no-ui"}, "play -x --no-ui", "", false},
		{nil, "", "", false},
	} {
		fs := flag.NewFlagSet("qari", flag.ContinueOnError)
		theme, noUI := fs.String("theme", "", ""), fs.Bool("no-ui", false, "")
		if pos := strings.Join(parseArgs(fs, c.args), " "); pos != c.pos || *theme != c.theme || *noUI != c.noUI {
			t.Errorf("%q: %q theme %q no-ui %v", c.args, pos, *theme, *noUI)
		}
	}
}

func TestCompletions(t *testing.T) {
	names := func(what, qari string) []ui.Cand {
		switch what {
		case "qari":
			return []ui.Cand{{Value: "Mishary Alafasi", Desc: "Hafs", Search: "mishary alafasi"}, {Value: "Maher Al Meaqli", Search: "maher al meaqli"}}
		case "surah":
			return []ui.Cand{{Value: "1", Desc: "Al-Fatihah", Search: "al fatihah"}, {Value: "18", Desc: "Al-Kahf ", Search: "al kahf"}, {Value: "112", Desc: "Al-Ikhlas " + qari}}
		case "station":
			return []ui.Cand{{Value: "Radio Ahmad Al-Ajmy"}}
		case "recipe":
			return []ui.Cand{{Value: "Short surahs"}}
		}
		return nil
	}
	values := func(words ...string) string {
		var out []string
		for _, c := range completions(words, names) {
			out = append(out, c.Value)
		}
		return strings.Join(out, ",")
	}
	for _, c := range []struct {
		words []string
		want  string
	}{
		{nil, "play,radio,random,completion,help"},
		{[]string{"r"}, "radio,random"},
		{[]string{"--no-ui", "pl"}, "play"},
		{[]string{"--theme", ""}, "mihrab,andalus,lapis,madinah,contrast"},
		{[]string{"play", "--theme", "la"}, "lapis"},
		{[]string{"play", "--no-u"}, "--no-ui"},
		{[]string{"play", ""}, "Mishary Alafasi,Maher Al Meaqli"},
		{[]string{"play", "m"}, "Mishary Alafasi,Maher Al Meaqli"},
		{[]string{"play", `"Mish`}, "Mishary Alafasi"},
		{[]string{"play", "alafasi"}, "Mishary Alafasi"}, // not a prefix: words inside the name
		{[]string{"play", `Mishary\ Alafasi`, "1"}, "1,18,112"},
		{[]string{"play", "'Mishary Alafasi'", "--no-ui", "kahf"}, "Al-Kahf"},
		{[]string{"--lang", "ar", "radio", ""}, "Radio Ahmad Al-Ajmy"},
		{[]string{"random", ""}, "Short surahs"},
		{[]string{"completion", ""}, "bash,zsh,fish,powershell"},
		{[]string{"radio", "x", ""}, ""},
		{[]string{"play", "a", "b", ""}, ""},
	} {
		if got := values(c.words...); got != c.want {
			t.Errorf("%q: %q, want %q", c.words, got, c.want)
		}
	}
	// The surah list is narrowed by the qari typed before it.
	if c := completions([]string{"play", `"Maher Al Meaqli"`, "11"}, names); len(c) != 1 || c[0].Desc != "Al-Ikhlas Maher Al Meaqli" {
		t.Errorf("surahs for a quoted qari: %+v", c)
	}
	for _, s := range []string{"bash", "zsh", "fish", "powershell"} {
		if !strings.Contains(scripts[s], "__complete") {
			t.Errorf("%s script does not call __complete", s)
		}
	}
}

func TestNewer(t *testing.T) {
	for _, c := range []struct {
		latest, current string
		want            bool
	}{
		{"v0.3.0", "0.2.1", true},
		{"v0.10.0", "0.9.9", true},
		{"v1.0.0", "v0.99.99", true},
		{"v0.2.1", "0.2.1", false},
		{"v0.2.0", "0.2.1", false},
		{"v0.3.0", "dev", false},
		{"v0.3.0", "0.2.2-SNAPSHOT-abc", false},
		{"", "0.2.1", false},
	} {
		if got := newer(c.latest, c.current); got != c.want {
			t.Errorf("newer(%q, %q) = %v", c.latest, c.current, got)
		}
	}
}

func TestUpdateNoticeThrottle(t *testing.T) {
	now := time.Date(2026, 11, 27, 12, 0, 0, 0, time.UTC)
	calls := 0
	latest := func(v string, err error) func() (string, error) {
		return func() (string, error) { calls++; return v, err }
	}
	var u store.UpdateCheck
	u, s := updateNotice(u, "0.2.1", now, latest("v0.3.0", nil))
	if calls != 1 || !strings.HasPrefix(s, "qari 0.3.0 is out") {
		t.Fatalf("first run: calls %d, %q", calls, s)
	}
	// Same day: no request, no second notice.
	u, s = updateNotice(u, "0.2.1", now.Add(time.Hour), latest("v0.3.0", nil))
	if calls != 1 || s != "" {
		t.Errorf("same day: calls %d, %q", calls, s)
	}
	// Next day: checks again; a failed check is silent but the known release
	// is announced again.
	u, s = updateNotice(u, "0.2.1", now.Add(25*time.Hour), latest("", errors.New("offline")))
	if calls != 2 || !strings.Contains(s, "0.3.0") || u.Latest != "v0.3.0" {
		t.Errorf("next day: calls %d, %q, %+v", calls, s, u)
	}
	// A newer version within the day of the last notice is announced at once.
	u.Checked = now.Add(2 * time.Hour)
	u, s = updateNotice(u, "0.2.1", now.Add(26*time.Hour+time.Minute), latest("v0.4.0", nil))
	if calls != 3 || !strings.Contains(s, "0.4.0") {
		t.Errorf("new version: calls %d, %q", calls, s)
	}
	// Up to date, and dev builds never check.
	if _, s := updateNotice(store.UpdateCheck{}, "0.4.0", now, latest("v0.4.0", nil)); s != "" {
		t.Errorf("up to date: %q", s)
	}
	n := calls
	if _, s := updateNotice(store.UpdateCheck{}, "dev", now, latest("v0.4.0", nil)); s != "" || calls != n {
		t.Errorf("dev build checked")
	}
}
