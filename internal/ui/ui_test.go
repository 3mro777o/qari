package ui

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode"

	"github.com/3mro777o/qari/internal/api"
	"github.com/3mro777o/qari/internal/norm"
	"github.com/3mro777o/qari/internal/player"
	"github.com/3mro777o/qari/internal/popular"
	"github.com/3mro777o/qari/internal/store"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func fullSurahList() string {
	parts := make([]string, 114)
	for i := 1; i <= 114; i++ {
		parts[i-1] = strconv.Itoa(i)
	}
	return strings.Join(parts, ",")
}

// buildFixtureCatalog constructs a rich, realistic catalog covering multi-moshaf qaris,
// partial recordings, multiple riwayat, Arabic surah names, long names, and radio stations.
func buildFixtureCatalog() (*api.Catalog, map[int]popular.Entry) {
	all114 := fullSurahList()
	reciters := []*api.ReciterEntry{
		{
			Reciter: api.Reciter{
				ID:   1,
				Name: "Mishary Rashid Alafasy",
				Date: "2025-10-01",
				Moshaf: []api.Moshaf{
					{ID: 101, Name: "Rewayat Hafs A'n Assem - Murattal", RewayaID: 1, Server: "https://server8.mp3quran.net/afs/", SurahTotal: 114, MoshafType: 11, SurahList: all114},
					{ID: 102, Name: "Rewayat Al-Duri A'n Al-Kisa'i - Murattal", RewayaID: 4, Server: "https://server8.mp3quran.net/afs_duri/", SurahTotal: 114, MoshafType: 41, SurahList: all114},
				},
			},
			Arabic:   "مشاري راشد العفاسي",
			Search:   norm.Fold("Mishary Rashid Alafasy مشاري راشد العفاسي"),
			Complete: true,
		},
		{
			Reciter: api.Reciter{
				ID:   2,
				Name: "Mahmoud Khalil Al-Hussary",
				Date: "2024-05-12",
				Moshaf: []api.Moshaf{
					{ID: 201, Name: "Rewayat Hafs A'n Assem - Mujawwad", RewayaID: 1, Server: "https://server13.mp3quran.net/husr_mjwd/", SurahTotal: 114, MoshafType: 12, SurahList: all114},
					{ID: 202, Name: "Rewayat Hafs A'n Assem - Murattal", RewayaID: 1, Server: "https://server13.mp3quran.net/husr/", SurahTotal: 114, MoshafType: 11, SurahList: all114},
					{ID: 203, Name: "Rewayat Warsh A'n Nafi' - Murattal", RewayaID: 2, Server: "https://server13.mp3quran.net/husr_warsh/", SurahTotal: 114, MoshafType: 21, SurahList: all114},
					{ID: 204, Name: "Rewayat Hafs A'n Assem - Teaching", RewayaID: 1, Server: "https://server13.mp3quran.net/husr_teach/", SurahTotal: 114, MoshafType: 13, SurahList: all114},
				},
			},
			Arabic:   "محمود خليل الحصري",
			Search:   norm.Fold("Mahmoud Khalil Al-Hussary محمود خليل الحصري"),
			Complete: true,
		},
		{
			Reciter: api.Reciter{
				ID:   3,
				Name: "Abdulrahman Al-Sudais With An Extraordinarily Long Traditional Name For Truncation Testing",
				Date: "2026-01-15",
				Moshaf: []api.Moshaf{
					{ID: 301, Name: "Rewayat Hafs A'n Assem - Murattal", RewayaID: 1, Server: "https://server11.mp3quran.net/sds/", SurahTotal: 114, MoshafType: 11, SurahList: all114},
				},
			},
			Arabic:   "عبدالرحمن السديس",
			Search:   norm.Fold("Abdulrahman Al-Sudais عبدالرحمن السديس"),
			Complete: true,
		},
		{
			Reciter: api.Reciter{
				ID:   4,
				Name: "Yassin Al-Jazaery",
				Date: "2023-03-01",
				Moshaf: []api.Moshaf{
					{ID: 401, Name: "Rewayat Warsh A'n Nafi' - Murattal", RewayaID: 2, Server: "https://server11.mp3quran.net/yassin/", SurahTotal: 114, MoshafType: 21, SurahList: all114},
				},
			},
			Arabic:   "ياسين الجزائري",
			Search:   norm.Fold("Yassin Al-Jazaery ياسين الجزائري"),
			Complete: true,
		},
		{
			Reciter: api.Reciter{
				ID:   5,
				Name: "Partial Obscure Reciter",
				Date: "2026-02-20",
				Moshaf: []api.Moshaf{
					{ID: 501, Name: "Rewayat Qalun A'n Nafi' - Murattal", RewayaID: 3, Server: "https://server10.mp3quran.net/obscure/", SurahTotal: 3, MoshafType: 31, SurahList: "1,18,36"},
				},
			},
			Arabic:   "قارئ غير مشهور",
			Search:   norm.Fold("Partial Obscure Reciter قارئ غير مشهور"),
			Complete: false,
		},
	}

	surahSamples := map[int][2]string{
		1:   {"Al-Fatihah", "الفاتحة"},
		2:   {"Al-Baqarah", "البقرة"},
		18:  {"Al-Kahf", "الكهف"},
		36:  {"Ya-Sin", "يس"},
		55:  {"Ar-Rahman", "الرحمن"},
		67:  {"Al-Mulk", "الملك"},
		112: {"Al-Ikhlas", "الإخلاص"},
		113: {"Al-Falaq", "الفلق"},
		114: {"An-Nas", "الناس"},
	}

	surahs := make([]api.SurahEntry, 114)
	for i := 1; i <= 114; i++ {
		en := fmt.Sprintf("Surah %d", i)
		ar := "سورة"
		if pair, ok := surahSamples[i]; ok {
			en, ar = pair[0], pair[1]
		}
		makkia := 1
		if i == 2 {
			makkia = 0
		}
		surahs[i-1] = api.SurahEntry{
			Surah:  api.Surah{ID: i, Name: en, Makkia: makkia, StartPage: i * 5, EndPage: i*5 + 4},
			Arabic: ar,
			Search: norm.Fold(en + " " + ar),
		}
	}

	cat := &api.Catalog{
		Lang:     "eng",
		Reciters: reciters,
		Surahs:   surahs,
		Riwayat: map[int]string{
			1: "Rewayat Hafs A'n Assem",
			2: "Rewayat Warsh A'n Nafi'",
			3: "Rewayat Qalun A'n Nafi'",
			4: "Rewayat Al-Duri A'n Al-Kisa'i",
		},
		Radios: []api.Radio{
			{ID: 1, Name: "Quran Radio Cairo - Live Broadcast", URL: "https://stream.radiojar.com/8s5u5tpdtwzuv"},
			{ID: 2, Name: "Mishary Alafasy Continuous Radio", URL: "https://backup.qurango.net/radio/mishary_alafasy"},
			{ID: 3, Name: "Amazing Recitations Radio Station With A Very Long Broadcast Title", URL: "https://backup.qurango.net/radio/amazing"},
		},
	}

	// Populate internal byID map via Reciter helper? api.Catalog.byID is unexported,
	// so load through a tiny cached Client or test both fixture & cached API catalog!
	// Wait: let's build a real *api.Catalog through api.LoadCatalog? We can use a fake HTTP RoundTripper!
	pop := map[int]popular.Entry{
		1: {Listens: 98500, Known: true},
		2: {Listens: 74200, Known: true},
		3: {Listens: 61000, Known: true},
		4: {Listens: 12000, Known: true},
		5: {Listens: 0, Known: false},
	}
	return cat, pop
}

// buildRealCatalogFromFixture creates a genuine *api.Catalog with populated byID
// by serving the fixture JSON over an in-memory http.Client to api.LoadCatalog.
func buildCatalogViaAPI(t *testing.T) (*api.Catalog, map[int]popular.Entry) {
	t.Helper()
	cat, pop := buildFixtureCatalog()

	recJSON := func(arabic bool) string {
		var parts []string
		for _, r := range cat.Reciters {
			name := r.Name
			if arabic {
				name = r.Arabic
			}
			var ms []string
			for _, m := range r.Moshaf {
				ms = append(ms, fmt.Sprintf(`{"id":%d,"name":%q,"rewaya_id":%d,"server":%q,"surah_total":%d,"moshaf_type":%d,"surah_list":%q}`,
					m.ID, m.Name, m.RewayaID, m.Server, m.SurahTotal, m.MoshafType, m.SurahList))
			}
			parts = append(parts, fmt.Sprintf(`{"id":%d,"name":%q,"date":%q,"moshaf":[%s]}`,
				r.ID, name, r.Date, strings.Join(ms, ",")))
		}
		return `{"reciters":[` + strings.Join(parts, ",") + `]}`
	}
	suwarJSON := func(arabic bool) string {
		var parts []string
		for _, s := range cat.Surahs {
			name := s.Name
			if arabic {
				name = s.Arabic
			}
			parts = append(parts, fmt.Sprintf(`{"id":%d,"name":%q,"makkia":%d,"start_page":%d,"end_page":%d}`,
				s.ID, name, s.Makkia, s.StartPage, s.EndPage))
		}
		return `{"suwar":[` + strings.Join(parts, ",") + `]}`
	}
	riwayatJSON := `{"riwayat":[{"id":1,"name":"Rewayat Hafs A'n Assem"},{"id":2,"name":"Rewayat Warsh A'n Nafi'"},{"id":3,"name":"Rewayat Qalun A'n Nafi'"},{"id":4,"name":"Rewayat Al-Duri A'n Al-Kisa'i"}]}`
	radiosJSON := `{"radios":[{"id":1,"name":"Quran Radio Cairo - Live Broadcast","url":"https://stream.radiojar.com/8s5u5tpdtwzuv"},{"id":2,"name":"Mishary Alafasy Continuous Radio","url":"https://backup.qurango.net/radio/mishary_alafasy"},{"id":3,"name":"Amazing Recitations Radio Station With A Very Long Broadcast Title","url":"https://backup.qurango.net/radio/amazing"}]}`

	cacheDir := t.TempDir()
	files := map[string]string{
		"reciters_language-eng.json": recJSON(false),
		"reciters_language-ar.json":  recJSON(true),
		"suwar_language-eng.json":    suwarJSON(false),
		"suwar_language-ar.json":     suwarJSON(true),
		"riwayat_language-eng.json":  riwayatJSON,
		"radios_language-eng.json":   radiosJSON,
	}
	for name, body := range files {
		if err := writeTestFile(filepath.Join(cacheDir, name), body); err != nil {
			t.Fatal(err)
		}
	}
	c := api.New(nil, cacheDir)
	loaded, err := api.LoadCatalog(t.Context(), c, "eng")
	if err != nil {
		t.Fatalf("LoadCatalog from test cache: %v", err)
	}
	return loaded, pop
}

func writeTestFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}

func newTestStore(t *testing.T) *store.Store {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	st.Favorites[1] = true
	st.Update(store.Track{Reciter: 1, Moshaf: 101, Surah: 18}, 740, 2290)
	st.Update(store.Track{Reciter: 2, Moshaf: 202, Surah: 36}, 600, 1200)
	st.MarkFinished(store.Track{Reciter: 1, Moshaf: 101, Surah: 1})
	return st
}

// assertFrame verifies all hard technical constraints on a single rendered frame:
// 1. Exact height lines
// 2. Every line exact width display cells
// 3. No forbidden glyphs (☾ ◑ ◕ ⏸, emoji, Arabic ornaments)
// 4. No ANSI background fill codes
// 5. No displayed qari Arabic names
// 6. Balanced U+2067..U+2069 isolates around any displayed Arabic text
func assertFrame(t *testing.T, label, view string, w, h int, cat *api.Catalog, expectNoArabic bool) {
	t.Helper()
	lines := strings.Split(view, "\n")
	if len(lines) != h {
		t.Fatalf("%s: got %d lines, want exact %d", label, len(lines), h)
	}

	forbidden := map[rune]bool{
		'☾': true, '◑': true, '◕': true, '⏸': true,
		'\u06DD': true, '\u06DE': true, '\u06E9': true,
	}

	for row, line := range lines {
		dw := ansi.StringWidth(line)
		if dw != w {
			t.Fatalf("%s row %d: display width = %d, want exact %d\nline=%q", label, row, dw, w, ansi.Strip(line))
		}
		if strings.Contains(line, "\x1b[48;") {
			t.Fatalf("%s row %d: painted 24-bit/256 background area (forbidden for translucent terminals)", label, row)
		}
		plain := ansi.Strip(line)
		depth := 0
		for _, r := range plain {
			if forbidden[r] {
				t.Fatalf("%s row %d: forbidden glyph %U (%q)", label, row, r, r)
			}
			// Reject emoji blocks
			if r >= 0x1F300 && r <= 0x1FAFF {
				t.Fatalf("%s row %d: emoji glyph %U forbidden", label, row, r)
			}
			if r == '\u2067' {
				depth++
			} else if r == '\u2069' {
				depth--
				if depth < 0 {
					t.Fatalf("%s row %d: unbalanced PDI U+2069", label, row)
				}
			} else if unicode.In(r, unicode.Arabic) {
				if expectNoArabic {
					t.Fatalf("%s row %d: Arabic rune %U displayed when --no-arabic is set (line=%q)", label, row, r, plain)
				}
				if depth <= 0 {
					t.Fatalf("%s row %d: Arabic rune %U rendered outside U+2067..U+2069 isolate (line=%q)", label, row, r, plain)
				}
			}
		}
		if depth != 0 {
			t.Fatalf("%s row %d: unclosed RLI U+2067 isolate (depth=%d)", label, row, depth)
		}
		// Ensure NO qari Arabic names are ever displayed.
		if cat != nil {
			for _, rec := range cat.Reciters {
				if rec.Arabic != "" && strings.Contains(plain, rec.Arabic) {
					t.Fatalf("%s row %d: qari Arabic name %q must never be displayed", label, row, rec.Arabic)
				}
			}
		}
	}
}

func TestExactFrameDimensionsAllScreensThemesSizes(t *testing.T) {
	cat, pop := buildCatalogViaAPI(t)
	sizes := [][2]int{
		{60, 16},
		{64, 18},
		{80, 24},
		{100, 30},
		{130, 34},
		{200, 60},
	}

	makeBase := func(themeID string, noArabic bool) Model {
		st := newTestStore(t)
		m := New(Options{
			Store:    st,
			Popular:  pop,
			Catalog:  cat,
			Player:   player.New(nil),
			Theme:    themeID,
			NoArabic: noArabic,
			Version:  "test",
		})
		return m
	}

	scenarios := []struct {
		name  string
		setup func(m *Model)
	}{
		{"loading", func(m *Model) { m.loading = true }},
		{"network_error", func(m *Model) {
			m.loadErr = errors.New("reciters: response is not JSON (blocked by a network filter?)")
		}},
		{"home_idle", func(m *Model) { m.tab = tabHome; m.view = viewTabs }},
		{"home_fresh", func(m *Model) {
			m.store.Progress = map[string]*store.Progress{}
			m.store.Favorites = map[int]bool{}
			m.store.Last = nil
			m.tab = tabHome
		}},
		{"home_playing_toast_sleep", func(m *Model) {
			m.tab = tabHome
			m.playKind = playSurah
			m.playTrack = store.Track{Reciter: 1, Moshaf: 101, Surah: 18}
			m.playPos, m.playDur = 740, 2290
			m.sleepMode = sleep30
			m.sleepUntil = time.Now().Add(22 * time.Minute)
			m.toast = "Theme switched"
		}},
		{"qaris_curated", func(m *Model) { m.tab = tabQaris; m.view = viewTabs }},
		{"qaris_all_searching", func(m *Model) {
			m.tab = tabQaris
			m.store.Settings.ShowAll = true
			m.searching = true
			m.searchInput.SetValue("afasy")
			m.qarisList.query = "afasy"
		}},
		{"qaris_empty_search", func(m *Model) {
			m.tab = tabQaris
			m.qarisList.query = "zzzznotfound"
		}},
		{"qari_detail_playing", func(m *Model) {
			m.openQari(cat.Reciter(2), 202, 18)
			m.playKind = playSurah
			m.playTrack = store.Track{Reciter: 2, Moshaf: 202, Surah: 18}
			m.playPos, m.playDur = 300, 1800
		}},
		{"qari_detail_paused", func(m *Model) {
			m.openQari(cat.Reciter(1), 101, 36)
			m.playKind = playSurah
			m.playTrack = store.Track{Reciter: 1, Moshaf: 101, Surah: 36}
			m.paused = true
			m.playPos, m.playDur = 120, 900
		}},
		{"surahs_tab", func(m *Model) { m.tab = tabSurahs; m.view = viewTabs }},
		{"surah_detail", func(m *Model) { m.openSurah(18) }},
		{"radio_live", func(m *Model) {
			m.tab = tabRadio
			m.playKind = playRadio
			m.playRadio = cat.Radios[0]
		}},
		{"no_mpv", func(m *Model) {
			m.player = nil
			m.tab = tabQaris
		}},
		{"overlay_help", func(m *Model) {
			m.tab = tabHome
			m.overlay = overlayHelp
		}},
		{"overlay_help_scrolled", func(m *Model) {
			m.tab = tabHome
			m.overlay = overlayHelp
			m.helpOffset = 8
		}},
		{"overlay_riwayah", func(m *Model) {
			m.tab = tabQaris
			m.overlay = overlayRiwayah
		}},
		{"overlay_recording", func(m *Model) {
			m.openQari(cat.Reciter(2), 202, 1)
			m.overlay = overlayRecording
		}},
	}

	// Test every theme × every screen × key sizes, and test all 6 sizes across core themes.
	for ti, th := range Themes() {
		testSizes := sizes
		if ti > 0 && th.ID != "ansi" && th.ID != "parchment" && th.ID != "contrast" {
			testSizes = [][2]int{{60, 16}, {100, 30}, {200, 60}}
		}
		for _, sz := range testSizes {
			for _, sc := range scenarios {
				m := makeBase(th.ID, false)
				m.width, m.height = sz[0], sz[1]
				sc.setup(&m)
				view := m.View()
				label := fmt.Sprintf("theme=%s size=%dx%d screen=%s", th.ID, sz[0], sz[1], sc.name)
				assertFrame(t, label, view, sz[0], sz[1], cat, false)
			}
		}
	}

	// Verify --no-arabic removes all Arabic text from every screen.
	for _, sc := range scenarios {
		m := makeBase("mihrab", true)
		m.width, m.height = 120, 32
		sc.setup(&m)
		assertFrame(t, "no-arabic/"+sc.name, m.View(), 120, 32, cat, true)
	}
}

func TestThemesRequirements(t *testing.T) {
	themes := Themes()
	if len(themes) < 6 {
		t.Fatalf("got %d themes, need at least 6", len(themes))
	}
	var hasANSI, hasLight, hasContrast bool
	for _, th := range compiledThemes {
		if th.IsANSI {
			hasANSI = true
			// Verify every color token in the ANSI theme is empty or 0..15.
			colors := []string{
				string(th.Accent), string(th.Accent2), string(th.Success),
				string(th.Text), string(th.Subtext), string(th.Muted),
				string(th.Border), string(th.Fav), string(th.Live), string(th.Warn),
			}
			for _, c := range colors {
				if c == "" {
					continue
				}
				n, err := strconv.Atoi(c)
				if err != nil || n < 0 || n > 15 {
					t.Fatalf("ANSI theme %q uses non-16-color token %q", th.ID, c)
				}
			}
		}
		if th.IsLight {
			hasLight = true
		}
		if th.HighCon {
			hasContrast = true
		}
	}
	if !hasANSI || !hasLight || !hasContrast {
		t.Fatalf("missing required theme category: ansi=%v light=%v contrast=%v", hasANSI, hasLight, hasContrast)
	}
}

func TestBestMoshafAndFilteringSorting(t *testing.T) {
	cat, pop := buildCatalogViaAPI(t)
	applyPopularity(cat, pop)

	// Al-Hussary (ID 2) has Hafs Mujawwad (201), Hafs Murattal (202), Warsh Murattal (203), Hafs Teaching (204).
	hussary := cat.Reciter(2)

	// No preferred riwayah -> must pick Hafs Murattal (202), NOT Mujawwad (201) or Teaching (204).
	m, ok := bestMoshaf(cat, hussary, 0, 18, false)
	if !ok || m.ID != 202 {
		t.Fatalf("bestMoshaf default = %d (ok=%v), want 202 (Hafs Murattal)", m.ID, ok)
	}

	// Preferred riwayah = Warsh (2) -> must pick Warsh Murattal (203).
	m, ok = bestMoshaf(cat, hussary, 2, 18, false)
	if !ok || m.ID != 203 {
		t.Fatalf("bestMoshaf Warsh = %d (ok=%v), want 203", m.ID, ok)
	}

	// Well-known complete filter excludes reciter 5 (obscure & partial), unless favorited or showAll=true.
	favs := map[int]bool{}
	curated := filterQaris(cat, favs, false, 0, sortPopular, "")
	for _, r := range curated {
		if r.ID == 5 {
			t.Fatal("obscure partial reciter should be hidden by default")
		}
	}
	// Favoriting reciter 5 forces them to appear even when showAll=false!
	favs[5] = true
	withFav := filterQaris(cat, favs, false, 0, sortPopular, "")
	foundFav := false
	for _, r := range withFav {
		if r.ID == 5 {
			foundFav = true
		}
	}
	if !foundFav {
		t.Fatal("favorited qari must always show even when showAll=false")
	}

	// Sorting checks
	byPop := filterQaris(cat, map[int]bool{}, true, 0, sortPopular, "")
	if byPop[0].ID != 1 || byPop[1].ID != 2 {
		t.Fatalf("popular sort order wrong: got IDs %d, %d", byPop[0].ID, byPop[1].ID)
	}
	byName := filterQaris(cat, map[int]bool{}, true, 0, sortName, "")
	if byName[0].ID != 3 { // Abdulrahman...
		t.Fatalf("name sort first = %q, want Abdulrahman..", byName[0].Name)
	}
	byUpdated := filterQaris(cat, map[int]bool{}, true, 0, sortUpdated, "")
	if byUpdated[0].ID != 5 { // 2026-02-20
		t.Fatalf("updated sort first ID = %d (date %s), want 5", byUpdated[0].ID, byUpdated[0].Date)
	}

	// Surah search by number ("18"), English ("kahf"), and Arabic ("الكهف")
	for _, q := range []string{"18", "kahf", "الكهف", "الكَهف"} {
		res := filterSurahs(cat, q)
		if len(res) == 0 || res[0].ID != 18 {
			t.Fatalf("surah search %q: got %+v, want Al-Kahf (18) first", q, res)
		}
	}
}

func TestInteractionsKeysArabicLayoutAndPlayback(t *testing.T) {
	cat, pop := buildCatalogViaAPI(t)
	st := newTestStore(t)
	m := New(Options{
		Store:   st,
		Popular: pop,
		Catalog: cat,
		Player:  player.New(nil),
		Version: "test",
	})
	m.width, m.height = 120, 32

	sendKey := func(r rune) {
		nm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = nm.(Model)
	}

	// Switch to Qaris tab with '2', verify default showAll=false label says "show all".
	sendKey('2')
	if m.tab != tabQaris {
		t.Fatalf("tab = %v, want tabQaris", m.tab)
	}
	viewCurated := ansi.Strip(m.View())
	if !strings.Contains(viewCurated, "show all") {
		t.Fatalf("expected 'show all' action label when ShowAll=false")
	}

	// Press Arabic 'ش' (maps to 'a') -> toggles ShowAll=true and flips label to "well-known only"!
	sendKey('ش')
	if !m.store.Settings.ShowAll {
		t.Fatal("Arabic 'ش' should toggle ShowAll via keymap.LatinKey")
	}
	viewAll := ansi.Strip(m.View())
	if !strings.Contains(viewAll, "well-known only") {
		t.Fatalf("expected 'well-known only' action label when ShowAll=true")
	}

	// Press Arabic 'ظ' (maps to '/') -> opens search, then typing Arabic 'العفاسي' MUST NOT be translated to Latin!
	sendKey('ظ')
	if !m.searching {
		t.Fatal("Arabic 'ظ' should enter search mode")
	}
	for _, r := range "العفاسي" {
		sendKey(r)
	}
	if m.qarisList.query != "العفاسي" {
		t.Fatalf("search query = %q, want Arabic 'العفاسي' untranslated", m.qarisList.query)
	}
	if len(m.filteredQaris()) != 1 || m.filteredQaris()[0].ID != 1 {
		t.Fatalf("Arabic search for العفاسي should match Alafasy")
	}

	// Escape clears search and exits search focus.
	nm, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = nm.(Model)
	if m.searching || m.qarisList.query != "" {
		t.Fatal("esc should clear search and exit search mode")
	}

	// Theme switching with Arabic 'ؤ' (maps to 'c') persists theme choice!
	origTheme := m.theme.ID
	sendKey('ؤ')
	if m.theme.ID == origTheme || m.store.Settings.Theme != m.theme.ID {
		t.Fatalf("theme switch failed: orig=%s now=%s saved=%s", origTheme, m.theme.ID, m.store.Settings.Theme)
	}

	// Playback event updates progress; Ended marks finished and obeys sleep timer end-of-surah!
	tr := store.Track{Reciter: 1, Moshaf: 101, Surah: 36}
	m.playKind = playSurah
	m.playTrack = tr
	m.sleepMode = sleepEndOfSurah

	nm, _ = m.Update(playerEventMsg{ev: player.Event{Pos: 120, Duration: 900}})
	m = nm.(Model)
	if p := m.store.Get(tr); p == nil || p.Position != 120 {
		t.Fatalf("progress not recorded: %+v", p)
	}

	nm, _ = m.Update(playerEventMsg{ev: player.Event{Pos: 900, Duration: 900, Ended: true}})
	m = nm.(Model)
	if p := m.store.Get(tr); p == nil || !p.Finished {
		t.Fatal("Ended event should mark surah finished")
	}
	if m.sleepMode != sleepOff || m.playKind != playNone {
		t.Fatalf("sleepEndOfSurah should stop playback and clear timer: sleep=%v playKind=%v", m.sleepMode, m.playKind)
	}

	// Timed sleep timer expires cleanly and pauses playback.
	m.playKind = playSurah
	m.playTrack = tr
	m.paused = false
	m.sleepMode = sleep15
	m.sleepUntil = time.Now().Add(-1 * time.Second)
	nm, _ = m.Update(sleepTickMsg{now: time.Now()})
	m = nm.(Model)
	if m.sleepMode != sleepOff {
		t.Fatal("expired sleep timer should reset to sleepOff")
	}
}

type failTransport struct{}

func (failTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("network unreachable")
}

func TestOfflineCachedCatalogState(t *testing.T) {
	// Populate cache then backdate mtimes by 48h so api.Client falls back to stale cache and sets Offline()=true.
	cat, pop := buildCatalogViaAPI(t)
	cacheDir := t.TempDir()
	// Copy fixture cache via minimal valid files and expire them.
	rec := `{"reciters":[{"id":1,"name":"Mishary Rashid Alafasy","date":"2025-10-01","moshaf":[{"id":101,"name":"Rewayat Hafs A'n Assem - Murattal","rewaya_id":1,"server":"https://server8.mp3quran.net/afs/","surah_total":114,"moshaf_type":11,"surah_list":"` + fullSurahList() + `"}]}]}`
	suw := `{"suwar":[{"id":1,"name":"Al-Fatihah","makkia":1,"start_page":1,"end_page":1}]}`
	riw := `{"riwayat":[{"id":1,"name":"Rewayat Hafs A'n Assem"}]}`
	rad := `{"radios":[{"id":1,"name":"Quran Radio","url":"https://example.com/r"}]}`
	old := time.Now().Add(-48 * time.Hour)
	for name, body := range map[string]string{
		"reciters_language-eng.json": rec,
		"reciters_language-ar.json":  rec,
		"suwar_language-eng.json":    suw,
		"suwar_language-ar.json":     suw,
		"riwayat_language-eng.json":  riw,
		"radios_language-eng.json":   rad,
	} {
		p := filepath.Join(cacheDir, name)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(p, old, old); err != nil {
			t.Fatal(err)
		}
	}
	client := api.New(&http.Client{Transport: failTransport{}}, cacheDir)
	loaded, err := api.LoadCatalog(t.Context(), client, "eng")
	if err != nil {
		t.Fatalf("offline fallback failed: %v", err)
	}
	if !client.Offline() {
		t.Fatal("expected client.Offline() == true")
	}

	m := New(Options{
		Client:  client,
		Store:   newTestStore(t),
		Popular: pop,
		Catalog: loaded,
		Player:  player.New(nil),
		Version: "test",
	})
	m.width, m.height = 100, 28
	view := m.View()
	assertFrame(t, "offline_badge", view, 100, 28, cat, false)
	if !strings.Contains(ansi.Strip(view), "OFFLINE") {
		t.Fatal("expected OFFLINE indicator in header when serving stale cache")
	}
}

func TestRealCachedCatalogIfPresent(t *testing.T) {
	cacheDir := filepath.Join("..", "..", ".design", "cache", "qari")
	if _, err := os.Stat(filepath.Join(cacheDir, "reciters_language-eng.json")); err != nil {
		t.Skip("no local live API cache in .design/cache/qari")
	}
	client := api.New(&http.Client{Transport: failTransport{}}, cacheDir)
	cat, err := api.LoadCatalog(t.Context(), client, "eng")
	if err != nil {
		t.Fatalf("load local cached catalog: %v", err)
	}
	pop := popular.Load()
	sizes := [][2]int{{60, 16}, {64, 18}, {130, 34}, {200, 60}}
	for _, th := range Themes() {
		for _, sz := range sizes {
			m := New(Options{
				Client:  client,
				Store:   newTestStore(t),
				Popular: pop,
				Catalog: cat,
				Player:  player.New(nil),
				Theme:   th.ID,
				Version: "v1.0.0",
			})
			m.width, m.height = sz[0], sz[1]
			for _, tab := range []tabKind{tabHome, tabQaris, tabSurahs, tabRadio} {
				m.switchTab(tab)
				assertFrame(t, fmt.Sprintf("live/%s/%dx%d/tab%d", th.ID, sz[0], sz[1], tab), m.View(), sz[0], sz[1], cat, false)
			}
			m.openQari(cat.Reciters[0], 0, 18)
			assertFrame(t, fmt.Sprintf("live/%s/%dx%d/qariDetail", th.ID, sz[0], sz[1]), m.View(), sz[0], sz[1], cat, false)
			m.openSurah(18)
			assertFrame(t, fmt.Sprintf("live/%s/%dx%d/surahDetail", th.ID, sz[0], sz[1]), m.View(), sz[0], sz[1], cat, false)
			m.overlay = overlayHelp
			assertFrame(t, fmt.Sprintf("live/%s/%dx%d/help", th.ID, sz[0], sz[1]), m.View(), sz[0], sz[1], cat, false)
		}
	}
}
