package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode"

	"github.com/3mro777o/qari/internal/api"
	"github.com/3mro777o/qari/internal/download"
	"github.com/3mro777o/qari/internal/popular"
	"github.com/3mro777o/qari/internal/store"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

// testCatalog loads a small catalog through the real loader from cache files.
func testCatalog(t *testing.T) *api.Catalog {
	all := make([]string, 114)
	var suwar, suwarAr []string
	for i := range all {
		all[i] = fmt.Sprint(i + 1)
		suwar = append(suwar, fmt.Sprintf(`{"id":%d,"name":"Surah %d","makkia":1,"start_page":%d,"end_page":%d}`, i+1, i+1, i+2, i+3))
		suwarAr = append(suwarAr, fmt.Sprintf(`{"id":%d,"name":"سورة رقم %d"}`, i+1, i+1))
	}
	suwar[17] = `{"id":18,"name":"Al-Kahf","makkia":1,"start_page":293,"end_page":304}`
	suwarAr[17] = `{"id":18,"name":"الكهف"}`
	full := strings.Join(all, ",")
	rec := func(ar bool) string {
		names := map[bool][3]string{false: {"Mishary Alafasy", "Mahmoud Khalil Al-Hussary With A Very Long Name For Truncation", "Partial Reciter"},
			true: {"مشاري العفاسي", "محمود خليل الحصري", "قارئ"}}[ar]
		return fmt.Sprintf(`{"reciters":[
{"id":1,"name":%q,"date":"2025-01-01","moshaf":[{"id":101,"name":"Rewayat Hafs A'n Assem - Murattal","rewaya_id":1,"server":"https://x/afs/","surah_total":114,"moshaf_type":11,"surah_list":%q},
 {"id":102,"name":"Rewayat Warsh A'n Nafi' - Murattal","rewaya_id":2,"server":"https://x/afs2/","surah_total":114,"moshaf_type":21,"surah_list":%q}]},
{"id":2,"name":%q,"date":"2024-01-01","moshaf":[{"id":201,"name":"Rewayat Hafs A'n Assem - Mujawwad","rewaya_id":1,"server":"https://x/h/","surah_total":114,"moshaf_type":12,"surah_list":%q},
 {"id":202,"name":"Rewayat Hafs A'n Assem - Murattal","rewaya_id":1,"server":"https://x/h2/","surah_total":114,"moshaf_type":11,"surah_list":%q}]},
{"id":3,"name":%q,"date":"2026-01-01","moshaf":[{"id":301,"name":"Rewayat Qalon A'n Nafi' - Murattal","rewaya_id":3,"server":"https://x/p/","surah_total":3,"moshaf_type":31,"surah_list":"1,18,36"}]}]}`,
			names[0], full, full, names[1], full, full, names[2])
	}
	dir := t.TempDir()
	for name, body := range map[string]string{
		"reciters_language-eng.json": rec(false), "reciters_language-ar.json": rec(true),
		"suwar_language-eng.json": `{"suwar":[` + strings.Join(suwar, ",") + `]}`, "suwar_language-ar.json": `{"suwar":[` + strings.Join(suwarAr, ",") + `]}`,
		"riwayat_language-eng.json": `{"riwayat":[{"id":1,"name":"Rewayat Hafs A'n Assem"},{"id":2,"name":"Rewayat Warsh A'n Nafi'"},{"id":3,"name":"Rewayat Qalon A'n Nafi'"}]}`,
		"radios_language-eng.json":  `{"radios":[{"id":1,"name":"Cairo Quran Radio","url":"https://r/1"},{"id":2,"name":"A Radio Station With A Very Long Broadcast Title Indeed","url":"https://r/2"}]}`,
	} {
		os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644)
	}
	cat, err := api.LoadCatalog(t.Context(), api.New(nil, dir), "eng")
	if err != nil {
		t.Fatal(err)
	}
	return cat
}

func testModel(t *testing.T, w, h int) Model {
	st, _ := store.Open(filepath.Join(t.TempDir(), "state.json"))
	st.Favorites[1] = true
	st.Update(store.Track{Reciter: 1, Moshaf: 101, Surah: 18}, 700, 2200)
	st.MarkFinished(store.Track{Reciter: 2, Moshaf: 202, Surah: 1})
	m := New(Options{Store: st, Catalog: testCatalog(t), Downloads: download.New(t.TempDir(), nil),
		Popular: map[int]popular.Entry{1: {Listens: 9000, Known: true}, 2: {Listens: 5000, Known: true}}})
	return send(m, tea.WindowSizeMsg{Width: w, Height: h})
}

func send(m Model, msgs ...tea.Msg) Model {
	for _, msg := range msgs {
		nm, _ := m.Update(msg)
		m = nm.(Model)
	}
	return m
}

func keys(s ...string) []tea.Msg {
	var out []tea.Msg
	for _, k := range s {
		switch k {
		case "enter":
			out = append(out, tea.KeyMsg{Type: tea.KeyEnter})
		case "esc":
			out = append(out, tea.KeyMsg{Type: tea.KeyEsc})
		case "space":
			out = append(out, tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})
		default:
			out = append(out, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)})
		}
	}
	return out
}

// checkFrame enforces the rendering rules: exact size, no background
// colors, no double-width glyphs, Arabic only inside balanced isolates, and
// never a qari's Arabic name.
func checkFrame(t *testing.T, label string, m Model) {
	t.Helper()
	lines := strings.Split(m.View(), "\n")
	if len(lines) != m.h {
		t.Fatalf("%s: %d lines, want %d", label, len(lines), m.h)
	}
	for i, l := range lines {
		plain := ansi.Strip(l)
		if w := ansi.StringWidth(l); w != m.w {
			t.Fatalf("%s row %d: width %d, want %d: %q", label, i, w, m.w, plain)
		}
		if strings.Contains(l, "\x1b[48") || strings.Contains(l, ";48;") {
			t.Fatalf("%s row %d: background color", label, i)
		}
		depth := 0
		for _, r := range plain {
			switch {
			case strings.ContainsRune("◑◕☾⏸۝", r) || r >= 0x1F300:
				t.Fatalf("%s row %d: unsafe glyph %q", label, i, r)
			case r == '⁧':
				depth++
			case r == '⁩':
				depth--
			case unicode.In(r, unicode.Arabic) && depth <= 0:
				t.Fatalf("%s row %d: Arabic outside an isolate: %q", label, i, plain)
			}
		}
		if depth != 0 || strings.Contains(plain, "مشاري") {
			t.Fatalf("%s row %d: unbalanced isolate or qari Arabic name: %q", label, i, plain)
		}
	}
}

func TestFramesExactSize(t *testing.T) {
	states := map[string][]string{
		"home": nil, "qaris": {"2"}, "surahs": {"3"}, "radio": {"4"}, "downloads": {"5"},
		"qari": {"2", "enter"}, "surah": {"3", "j", "enter"}, "search": {"3", "/", "ا", "ل"},
		"help": {"?"}, "theme": {"c", "j"}, "riwayah": {"w"}, "recording": {"2", "v"}, "empty": {"2", "/", "z", "z", "z", "enter"},
	}
	for _, size := range [][2]int{{60, 16}, {80, 24}, {130, 34}, {200, 60}} {
		for name, ks := range states {
			m := send(testModel(t, size[0], size[1]), keys(ks...)...)
			checkFrame(t, fmt.Sprintf("%s %dx%d", name, size[0], size[1]), m)
			m.kind, m.track, m.pos, m.dur = playSurah, store.Track{Reciter: 1, Moshaf: 101, Surah: 18}, 60, 2200
			checkFrame(t, fmt.Sprintf("%s playing %dx%d", name, size[0], size[1]), m)
		}
	}
}

func TestKeysWorkOnArabicLayout(t *testing.T) {
	m := send(testModel(t, 100, 30), keys("٢")...) // the 2 key
	if m.page != pQaris {
		t.Fatalf("page %d, want qaris", m.page)
	}
	m = send(m, keys("ظ", "ح", "ص", "enter")...) // "/" opens search; the text stays Arabic
	if m.lists[pQaris].query != "حص" || m.count() != 1 {
		t.Fatalf("search %q found %d", m.lists[pQaris].query, m.count())
	}
	m = send(m, keys("enter")...)
	if m.page != pQari || m.qari.ID != 2 || m.moshaf.ID != 202 {
		t.Fatalf("opened page %d qari %v recording %d, want Hussary's Murattal", m.page, m.qari, m.moshaf.ID)
	}
	if m = send(m, keys("esc")...); m.page != pQaris {
		t.Fatal("esc did not go back")
	}
}

func TestThemePickerPreviewRevertApply(t *testing.T) {
	m := send(testModel(t, 100, 30), keys("c", "j")...)
	if m.ov != ovTheme || m.th.ID != themes[1].ID {
		t.Fatalf("picker did not preview: %s", m.th.ID)
	}
	if m = send(m, keys("esc")...); m.th.ID != "mihrab" || m.ov != ovNone {
		t.Fatalf("esc kept %s", m.th.ID)
	}
	m = send(m, keys("c", "j", "j", "enter")...)
	if m.th.ID != themes[2].ID || m.Store.Settings.Theme != themes[2].ID {
		t.Fatalf("enter applied %s, saved %q", m.th.ID, m.Store.Settings.Theme)
	}
}

func TestMouseTabsAndRows(t *testing.T) {
	m := testModel(t, 100, 30)
	_, zones := m.paint()
	click := func(m Model, kind, i int) Model {
		for _, z := range zones {
			if z.kind == kind && z.i == i {
				return send(m, tea.MouseMsg{X: z.x0, Y: z.y0, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
			}
		}
		t.Fatalf("no zone %d/%d", kind, i)
		return m
	}
	if m = click(m, zTab, pSurahs); m.page != pSurahs {
		t.Fatal("tab click")
	}
	_, zones = m.paint()
	if m = click(m, zRow, 2); m.lists[pSurahs].cursor != 2 || m.page != pSurahs {
		t.Fatal("first click should select")
	}
	if m = click(m, zRow, 2); m.page != pSurah || m.surah != 3 {
		t.Fatal("second click should open")
	}
	m = send(m, keys("c")...)
	_, zones = m.paint()
	if m = click(m, zPick, 2); m.th.ID != themes[2].ID {
		t.Fatal("clicking a theme should preview it")
	}
	if m = send(m, tea.MouseMsg{X: 0, Y: 0, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft}); m.ov != ovNone || m.th.ID != "mihrab" {
		t.Fatal("clicking outside the picker should cancel it")
	}
}

func TestBestRecording(t *testing.T) {
	m := testModel(t, 100, 30)
	cat := m.cat
	if mo, _ := best(cat, cat.Reciter(2), 0, 0, false); mo.ID != 202 {
		t.Fatalf("want Hafs Murattal, got %d", mo.ID)
	}
	if mo, _ := best(cat, cat.Reciter(1), 2, 0, false); mo.ID != 102 {
		t.Fatalf("want the chosen riwayah, got %d", mo.ID)
	}
	if n := len(m.qaris("")); n != 2 {
		t.Fatalf("well-known qaris: %d, want 2", n)
	}
	if s := m.surahs(nil, "18"); len(s) != 1 || s[0].ID != 18 {
		t.Fatal("a number should find that surah only")
	}
	if s := m.surahs(nil, "11"); len(s) != 6 || s[0].ID != 11 {
		t.Fatalf("11 should find 11 first, then 110-114: %d", len(s))
	}
}
