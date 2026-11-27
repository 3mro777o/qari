package ui

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/3mro777o/qari/internal/api"
	tea "github.com/charmbracelet/bubbletea"
)

var buffers = []int{5, 10, 25, 50, 100, 150} // MiB

// setting is one row of the Settings tab. change gets +1 (enter, right) or
// -1 (left); choices that open a picker ignore the sign.
type setting struct {
	name, hint string
	value      func(m *Model) string
	change     func(m *Model, d int) tea.Cmd
}

// settings is set in init: its functions refer back to it through count.
var settings []setting

func init() {
	settings = []setting{
		{"Buffer", "mpv memory per stream",
			func(m *Model) string { return fmt.Sprintf("%d MB", m.Store.Settings.Buffer()) },
			func(m *Model, d int) tea.Cmd {
				i := max(slices.Index(buffers, m.Store.Settings.Buffer()), 0)
				mb := buffers[(i+d+len(buffers))%len(buffers)]
				m.Store.Settings.BufferMB = mb
				m.save()
				if m.Player != nil {
					m.Player.SetBuffer(mb)
				}
				return m.flash(fmt.Sprintf("Buffer %d MB", mb))
			}},
		{"Volume", "+ -", func(m *Model) string { return fmt.Sprintf("%.0f%%", m.vol) },
			func(m *Model, d int) tea.Cmd { return m.volume(float64(5 * d)) }},
		{"Speed", "[ ]", func(m *Model) string { return fmt.Sprintf("%gx", speeds[m.speed]) },
			func(m *Model, d int) tea.Cmd { return m.setSpeed(m.speed + d) }},
		{"After a surah", "r", func(m *Model) string { return repeatLong[mode(repeatModes, m.Store.Settings.Repeat)] },
			func(m *Model, _ int) tea.Cmd { return m.cycleRepeat("") }},
		{"Qaris shown", "a", func(m *Model) string {
			return map[bool]string{false: "Well-known", true: "All"}[m.Store.Settings.ShowAll]
		}, func(m *Model, _ int) tea.Cmd { return m.showAll("") }},
		{"Sort qaris", "s", func(m *Model) string { return sortNames[mode(sortModes, m.Store.Settings.Sort)] },
			func(m *Model, _ int) tea.Cmd { return m.cycleSort("") }},
		{"Riwayah", "w", func(m *Model) string {
			if i := slices.IndexFunc(m.riwayat, func(o riwayaOption) bool { return o.id == m.Store.Settings.Rewaya }); i >= 0 {
				return m.riwayat[i].name
			}
			return "All"
		}, func(m *Model, _ int) tea.Cmd {
			m.openPick(ovRiwayah, slices.IndexFunc(m.riwayat, func(o riwayaOption) bool { return o.id == m.Store.Settings.Rewaya }))
			return nil
		}},
		{"Color theme", "c", func(m *Model) string { return m.th.Name },
			func(m *Model, _ int) tea.Cmd {
				m.themeWas = m.th
				m.openPick(ovTheme, themeIndex(m.th.ID))
				return nil
			}},
		{"Name language", "names of qaris and surahs", func(m *Model) string { return m.Lang },
			func(m *Model, _ int) tea.Cmd { return m.openLangs() }},
		{"Arabic names", "beside English names", func(m *Model) string { return onOff(!m.NoArabic) },
			func(m *Model, _ int) tea.Cmd {
				m.NoArabic = !m.NoArabic
				m.Store.Settings.Arabic = !m.NoArabic
				m.save()
				return m.flash("Arabic names " + onOff(!m.NoArabic))
			}},
		{"Notifications", "when a surah starts", func(m *Model) string { return onOff(!m.NoNotify) },
			func(m *Model, _ int) tea.Cmd {
				m.NoNotify = !m.NoNotify
				m.Store.Settings.NoNotify = m.NoNotify
				m.Desktop.SetNotify(!m.NoNotify)
				m.save()
				return m.flash("Notifications " + onOff(!m.NoNotify))
			}},
	}
}

func onOff(on bool) string { return map[bool]string{true: "on", false: "off"}[on] }

// settingsKey handles the keys that change a setting; ok is false for the rest.
func (m *Model) settingsKey(k string) (tea.Cmd, bool) {
	if m.page != pSettings {
		return nil, false
	}
	switch k {
	case "enter", "l", "right":
		return settings[m.list().cursor].change(m, 1), true
	case "left":
		return settings[m.list().cursor].change(m, -1), true
	}
	return nil, false
}

type langsMsg struct {
	langs []api.Language
	err   error
}

// openLangs fetches the name languages the API offers, then opens the picker.
func (m *Model) openLangs() tea.Cmd {
	if m.langs != nil {
		m.openPick(ovLang, max(slices.IndexFunc(m.langs, func(l api.Language) bool { return l.Locale == m.Lang }), 0))
		return nil
	}
	c := m.Client
	return tea.Batch(m.flash("Loading languages…"), func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		ls, err := c.Languages(ctx)
		return langsMsg{ls, err}
	})
}

func (m *Model) onLangs(msg langsMsg) tea.Cmd {
	if msg.err != nil || len(msg.langs) == 0 {
		return m.flash("Could not load languages")
	}
	m.langs = msg.langs
	return m.openLangs()
}

// setLang reloads the catalog with names in another language.
func (m *Model) setLang(l api.Language) tea.Cmd {
	if l.Locale == m.Lang {
		return nil
	}
	m.Lang, m.Store.Settings.Lang = l.Locale, l.Locale
	m.save()
	m.loading = true
	return tea.Batch(m.spin.Tick, m.load())
}

func (p *painter) settingsPage() page {
	s := p.s
	return page{
		head: []string{lr(" "+p.s.AccentB.Render("Settings"), p.kv("enter", "change")+" "+p.kv("← →", "adjust")+" ", p.w)},
		n:    len(settings),
		row: func(i, w int, sel bool) string {
			st := settings[i]
			return p.row(sel, "  "+fit(p.name(sel, st.name), 16)+s.AccentB.Render(trunc(st.value(p.Model), max(w-44, 8))), s.Muted.Render(trunc(st.hint, 26)), w)
		},
	}
}
