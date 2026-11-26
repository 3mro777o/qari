package ui

import tea "github.com/charmbracelet/bubbletea"

// onMouse: click a tab or row to select it, click the selected row again to
// open or play it, wheel to scroll, click or drag the progress bar to seek.
func (m *Model) onMouse(msg tea.MouseMsg) tea.Cmd {
	if m.cat == nil || m.loading {
		return nil
	}
	switch {
	case msg.Button == tea.MouseButtonWheelUp || msg.Button == tea.MouseButtonWheelDown:
		d := 1
		if msg.Button == tea.MouseButtonWheelUp {
			d = -1
		}
		switch m.ov {
		case ovNone:
			m.move(3 * d)
		case ovHelp:
			m.helpOff = max(m.helpOff+d, 0)
		default:
			m.pickMove(d)
		}
		return nil
	case m.drag && msg.Action == tea.MouseActionMotion:
		m.pos = m.seekAt(msg.X)
		return nil
	case m.drag && msg.Action == tea.MouseActionRelease:
		m.drag = false
		m.Player.SeekTo(m.pos)
		return nil
	case msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft:
		return nil
	}
	_, zones := m.paint()
	var z zone
	for i := len(zones) - 1; i >= 0; i-- {
		if c := zones[i]; msg.X >= c.x0 && msg.X <= c.x1 && msg.Y >= c.y0 && msg.Y <= c.y1 {
			z = c
			break
		}
	}
	switch z.kind {
	case 0:
		if m.ov == ovHelp {
			m.ov = ovNone
		} else if m.ov != ovNone {
			m.closePick()
		}
	case zTab:
		m.goTab(z.i)
	case zRow:
		if m.list().cursor == z.i {
			return m.activate()
		}
		m.list().cursor = z.i
		m.clamp()
	case zPick:
		if m.pick.cursor == z.i {
			return m.choose()
		}
		m.pickMove(z.i - m.pick.cursor)
	case zPlay:
		return m.togglePause()
	case zSeek:
		if m.kind == playSurah && m.dur > 0 {
			m.seekX0, m.seekX1, m.drag = z.x0, z.x1, true
			m.pos = m.seekAt(msg.X)
		}
	}
	return nil
}

// seekAt is the position in the surah under column x of the progress bar.
func (m *Model) seekAt(x int) float64 {
	f := float64(x-m.seekX0) / float64(max(m.seekX1-m.seekX0, 1))
	return min(max(f, 0), 1) * m.dur
}
