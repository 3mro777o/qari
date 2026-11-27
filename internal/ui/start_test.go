package ui

import (
	"errors"
	"strings"
	"testing"

	"github.com/3mro777o/qari/internal/store"
)

func TestResolve(t *testing.T) {
	m := testModel(t, 100, 30)
	tr := func(r, mo, s int) store.Track { return store.Track{Reciter: r, Moshaf: mo, Surah: s} }
	for _, c := range []struct {
		cmd  string
		args []string
		want store.Entry
	}{
		{"play", []string{"alafasy", "kahf"}, store.Entry{Track: tr(1, 101, 18)}},
		{"play", []string{"mishary", "alafasy", "18"}, store.Entry{Track: tr(1, 101, 18)}}, // no quotes needed
		{"play", []string{"العفاسي", "الكهف"}, store.Entry{Track: tr(1, 101, 18)}},
		{"play", []string{"2", "1"}, store.Entry{Track: tr(2, 202, 1)}},   // qari id; Murattal preferred
		{"play", []string{"alafasy"}, store.Entry{Track: tr(1, 101, 18)}}, // resumes the saved surah
		{"play", []string{"partial"}, store.Entry{Track: tr(3, 301, 1)}},  // nothing heard: first surah
		{"radio", []string{"cairo"}, store.Entry{Radio: 1}},
	} {
		s, err := Resolve(m.Options, c.cmd, c.args)
		if err != nil || s.Recipe != nil || s.Entry != c.want {
			t.Errorf("%s %q = %+v, %v; want %+v", c.cmd, c.args, s, err, c.want)
		}
	}
	if s, err := Resolve(m.Options, "random", []string{"short"}); err != nil || s.Recipe.Name != "Short surahs" {
		t.Errorf("random short = %+v, %v", s, err)
	}
	if s, err := Resolve(m.Options, "random", nil); err != nil || s.Recipe.Name != "Anything" {
		t.Errorf("random = %+v, %v", s, err)
	}

	for _, c := range []struct {
		cmd   string
		args  []string
		cands int // -1: not a NotFound
		msg   string
	}{
		{"play", []string{"m"}, 2, "several qaris"},
		{"play", []string{"zzz"}, 0, `no qari matches "zzz"`},
		{"play", []string{"alafasy", "zzz"}, 0, `no surah matches "zzz"`},
		{"play", []string{"alafasy", "200"}, -1, "does not exist"},
		{"play", []string{"partial", "2"}, -1, "no recording of"},
		{"radio", []string{"radio"}, 2, "several stations"},
		{"random", []string{"favorite"}, 2, "several recipes"},
	} {
		_, err := Resolve(m.Options, c.cmd, c.args)
		var nf *NotFound
		if err == nil || !strings.Contains(err.Error(), c.msg) || errors.As(err, &nf) != (c.cands >= 0) || nf != nil && len(nf.Candidates) != c.cands {
			t.Errorf("%s %q: %v, want %d candidates and %q", c.cmd, c.args, err, c.cands, c.msg)
		}
	}
}

func TestStartOpensQariHeadless(t *testing.T) {
	m := testModel(t, 100, 30)
	m.Start, m.Headless = &Start{Entry: store.Entry{Track: store.Track{Reciter: 1, Moshaf: 101, Surah: 18}}}, true
	if m.Init() == nil {
		t.Fatal("no start command")
	}
	nm, cmd := m.Update(startMsg{})
	m = nm.(Model)
	if m.page != pQari || m.qari.ID != 1 {
		t.Errorf("start did not open the qari page: page %d", m.page)
	}
	// Without mpv nothing plays, so --no-ui quits with the reason on its line.
	if cmd == nil || !m.headlessDone() || !strings.Contains(m.View(), "Install mpv") {
		t.Errorf("headless start without mpv: done %v, view %q", m.headlessDone(), m.View())
	}
	m.kind, m.track, m.pos, m.dur, m.toast = playSurah, store.Track{Reciter: 1, Moshaf: 101, Surah: 18}, 65, 2200, ""
	if v := m.View(); v != "Al-Kahf (18) · Mishary Alafasy · 1:05/36:40" {
		t.Errorf("status line %q", v)
	}
}

func TestNames(t *testing.T) {
	m := testModel(t, 100, 30)
	q := Names(m.Options, "qari", "")
	if len(q) != 3 || q[0].Value != "Mishary Alafasy" || q[0].Desc != "Hafs · Murattal" {
		t.Errorf("qaris %+v", q)
	}
	if s := Names(m.Options, "surah", "partial"); len(s) != 3 || s[1].Value != "18" || s[1].Desc != "Al-Kahf" {
		t.Errorf("surahs of partial %+v", s)
	}
	if n := len(Names(m.Options, "surah", "")); n != 114 {
		t.Errorf("all surahs: %d", n)
	}
	if r := Names(m.Options, "recipe", ""); len(r) != len(builtins) {
		t.Errorf("recipes %d", len(r))
	}
}
