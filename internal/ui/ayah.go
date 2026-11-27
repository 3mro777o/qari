package ui

// Ayah repeat for memorizing (hifz): mpv seeks within the surah file using
// the per-ayah timings mp3quran publishes for some recordings.

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/3mro777o/qari/internal/api"
	"github.com/3mro777o/qari/internal/store"
	tea "github.com/charmbracelet/bubbletea"
)

type (
	// timingMsg brings fetched timings; then runs once they are known.
	timingMsg struct {
		tr    store.Track
		reads []api.TimingRead
		read  int
		t     []api.AyahTiming
		err   error
		then  func(*Model, []api.AyahTiming) tea.Cmd
	}
	gapMsg int
)

var eachSteps = []int{1, 2, 3, 5, 7, 10}

// repeat is an ayah repeat in progress. at moves it along with the playback
// position, so it can be tested without mpv.
type repeat struct {
	store.Hifz
	t                []api.AyahTiming
	ayah, n, round   int // current ayah, its play (1..Each), the range round
	dur              float64
	seeking, gapping bool // a seek was sent and has not landed; paused between repeats
	waited           int  // positions ignored while seeking
}

// repAct is what the player must do: seek >= 0 jumps there, gap pauses
// first, end (a message) stops the repeat.
type repAct struct {
	seek float64
	gap  bool
	end  string
}

// span is where ayah a plays, in seconds; the end stays clear of the file's end.
func (r *repeat) span(a int) (float64, float64) {
	i := slices.IndexFunc(r.t, func(t api.AyahTiming) bool { return t.Ayah == a })
	if i < 0 {
		return 0, 0
	}
	s, e := float64(r.t[i].Start)/1000, float64(r.t[i].End)/1000
	if r.dur > 0 {
		e = min(e, r.dur-0.3)
	}
	return s, e
}

func (r *repeat) at(pos float64) repAct {
	none := repAct{seek: -1}
	s, e := r.span(r.ayah)
	if r.seeking {
		// Seeks land at the ayah's start; positions from before the seek (a
		// pause reports one too) are stale. Give up waiting after a while.
		if r.waited++; (pos < s-0.5 || pos >= min(s+1.5, e-0.1)) && r.waited < 100 {
			return none
		}
		r.seeking, r.waited = false, 0
	}
	if pos < s-1 || pos > e+1 { // the listener moved: follow inside the range, stop outside it
		a := ayahAt(r.t, pos)
		if a < r.From || a > r.To {
			return repAct{seek: -1, end: "Ayah repeat off"}
		}
		r.ayah, r.n = a, 1
		return none
	}
	if pos < e-0.05 {
		return none
	}
	switch {
	case r.n < r.Each:
		r.n++
	case r.ayah < r.To:
		r.ayah, r.n = r.ayah+1, 1
		ns, _ := r.span(r.ayah)
		if ns-e < 0.3 && e-ns < 0.3 {
			return none // the next ayah follows on
		}
		r.seeking = true
		return repAct{seek: ns}
	case r.Times == 0 || r.round < r.Times:
		r.round++
		r.ayah, r.n = r.From, 1
	default:
		return repAct{seek: -1, end: "Ayah repeat finished"}
	}
	s, _ = r.span(r.ayah)
	r.seeking = true
	return repAct{seek: s, gap: r.Gap > 0}
}

// status is "ayah 3/7 · 2 of 3 · range 1 of 2".
func (r *repeat) status() string {
	out := fmt.Sprintf("ayah %d/%d", r.ayah, r.To)
	if r.Each > 1 {
		out += fmt.Sprintf(" · %d of %d", r.n, r.Each)
	}
	switch {
	case r.Times == 0:
		out += fmt.Sprintf(" · range %d, looping", r.round)
	case r.Times > 1:
		out += fmt.Sprintf(" · range %d of %d", r.round, r.Times)
	}
	return out
}

// ayahAt is the ayah playing at pos (0 = the opening), or -1 outside them all.
func ayahAt(t []api.AyahTiming, pos float64) int {
	for _, a := range t {
		if pos >= float64(a.Start)/1000 && pos < float64(a.End)/1000 {
			return a.Ayah
		}
	}
	return -1
}

func lastAyah(t []api.AyahTiming) int {
	if len(t) == 0 {
		return 0
	}
	return t[len(t)-1].Ayah
}

// hifzFields are the popup's rows: label, value, lowest and highest value.
func hifzFields(h *store.Hifz, n int) []struct {
	label  string
	v      *int
	lo, hi int
} {
	return []struct {
		label  string
		v      *int
		lo, hi int
	}{{"From ayah", &h.From, 1, n}, {"To ayah", &h.To, 1, n}, {"Each ayah", &h.Each, 1, 99}, {"Whole range", &h.Times, 0, 99}, {"Gap", &h.Gap, 0, 60}}
}

// hifzErr checks the popup's values for a surah of n ayahs.
func hifzErr(h store.Hifz, n int) string {
	if h.From > h.To {
		return "From must not be after To"
	}
	for _, f := range hifzFields(&h, n) {
		if *f.v < f.lo || *f.v > f.hi {
			return fmt.Sprintf("%s must be %d-%d", f.label, f.lo, f.hi)
		}
	}
	return ""
}

// hifzValue is how the popup shows field i.
func hifzValue(h store.Hifz, i int) string {
	switch v := *hifzFields(&h, 0)[i].v; {
	case i == 3 && v == 0:
		return "until stopped"
	case i == 4 && v == 0:
		return "none"
	case i == 4:
		return fmt.Sprintf("%d s", v)
	case i >= 2:
		return plural(v, "time")
	default:
		return fmt.Sprint(v)
	}
}

// withTimings runs then with the ayah timings of tr, fetching them first if
// needed (the api client caches them on disk, so they work offline later).
func (m *Model) withTimings(tr store.Track, then func(*Model, []api.AyahTiming) tea.Cmd) tea.Cmd {
	if m.Player == nil {
		return m.flash("Install mpv to play audio")
	}
	mo, _ := findMoshaf(m.cat.Reciter(tr.Reciter), tr.Moshaf)
	if m.reads != nil {
		read := api.ReadFor(m.reads, mo)
		if read == 0 {
			return m.flash(m.noTimings(tr))
		}
		if t, ok := m.timings[[2]int{read, tr.Surah}]; ok {
			return then(m, t)
		}
	}
	if m.Client == nil {
		return m.flash("No ayah timings for this recording")
	}
	c, reads := m.Client, m.reads
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		msg := timingMsg{tr: tr, reads: reads, then: then}
		if msg.reads == nil {
			if msg.reads, msg.err = c.TimingReads(ctx); msg.err != nil {
				return msg
			}
		}
		if msg.read = api.ReadFor(msg.reads, mo); msg.read > 0 {
			msg.t, msg.err = c.AyahTimings(ctx, msg.read, tr.Surah)
		}
		return msg
	}
}

func (m *Model) onTimings(msg timingMsg) tea.Cmd {
	if msg.reads != nil {
		m.reads = msg.reads
	}
	switch {
	case msg.err != nil:
		return m.flash("Could not load ayah timings: " + msg.err.Error())
	case msg.read == 0:
		return m.flash(m.noTimings(msg.tr))
	case len(msg.t) < 2:
		return m.flash("No ayah timings for this surah")
	}
	if m.timings == nil {
		m.timings = map[[2]int][]api.AyahTiming{}
	}
	m.timings[[2]int{msg.read, msg.tr.Surah}] = msg.t
	return msg.then(m, msg.t)
}

// noTimings names another recording of the same qari that has timings.
func (m *Model) noTimings(tr store.Track) string {
	if rec := m.cat.Reciter(tr.Reciter); rec != nil {
		for _, mo := range rec.Moshaf {
			if mo.ID != tr.Moshaf && mo.Surahs()[tr.Surah] && api.ReadFor(m.reads, mo) > 0 {
				return "No ayah timings for this recording · v → " + label(m.cat, mo, true) + " has them"
			}
		}
	}
	return "No ayah timings for this recording"
}

// openAyahs opens the repeat popup for the surah under the cursor, or the
// one playing.
func (m *Model) openAyahs(string) tea.Cmd {
	tr := m.track
	if rec, mo, n := m.selected(); (m.page == pQari || m.page == pSurah) && rec != nil && n > 0 {
		tr = store.Track{Reciter: rec.ID, Moshaf: mo.ID, Surah: n}
	} else if m.kind != playSurah {
		return m.flash("Choose a surah first")
	}
	return m.withTimings(tr, func(m *Model, t []api.AyahTiming) tea.Cmd {
		if m.ov != ovNone {
			return nil
		}
		n, h := lastAyah(t), m.Store.Settings.Hifz
		if h.Each == 0 {
			h = store.Hifz{Each: 3, Times: 1}
		}
		if h.To = min(h.To, n); h.To < 1 {
			h.To = n
		}
		h.From = min(max(h.From, 1), h.To)
		m.hifz, m.hifzTrack, m.hifzT = h, tr, t
		m.openPick(ovAyahs, 0)
		return nil
	})
}

// ayahKey handles the repeat popup: arrows change a value, digits type it.
func (m *Model) ayahKey(k string) (tea.Cmd, bool) {
	if m.ov != ovAyahs {
		return nil, false
	}
	typed := m.hifzTyped
	m.hifzTyped = false
	f := hifzFields(&m.hifz, lastAyah(m.hifzT))[m.pick.cursor]
	switch {
	case k == "left" || k == "h" || k == "-":
		*f.v = max(*f.v-1, f.lo)
	case k == "right" || k == "l" || k == "+" || k == "=":
		*f.v = min(*f.v+1, f.hi)
	case len(k) == 1 && k[0] >= '0' && k[0] <= '9':
		if !typed {
			*f.v = 0
		}
		*f.v = min(*f.v*10+int(k[0]-'0'), 999)
		m.hifzTyped = true
	case k == "backspace":
		*f.v /= 10
		m.hifzTyped = true
	case k == "enter":
		return m.startHifz(), true
	default:
		return nil, false
	}
	return nil, true
}

func (m *Model) startHifz() tea.Cmd {
	if err := hifzErr(m.hifz, lastAyah(m.hifzT)); err != "" {
		return m.flash(err)
	}
	m.ov = ovNone
	m.Store.Settings.Hifz = m.hifz
	m.save()
	return m.startRepeat(m.hifzTrack, m.hifzT, m.hifz)
}

// startRepeat plays tr's ayahs h.From-h.To as h says.
func (m *Model) startRepeat(tr store.Track, t []api.AyahTiming, h store.Hifz) tea.Cmd {
	r := &repeat{Hifz: h, t: t, ayah: h.From, n: 1, round: 1, seeking: true}
	s, _ := r.span(h.From)
	var cmd tea.Cmd
	if m.kind == playSurah && m.track == tr {
		m.endRepeat()
		m.Player.SeekTo(s)
		if m.paused {
			m.Player.SetPause(false)
		}
		m.pos = s
	} else {
		m.recipe = nil
		cmd = m.playAt(tr, s)
	}
	r.dur = m.dur
	m.rep = r
	m.Player.Precise(true)
	what := fmt.Sprintf("ayahs %d-%d", h.From, h.To)
	if h.From == h.To {
		what = fmt.Sprint("ayah ", h.From)
	}
	return tea.Batch(cmd, m.flash("Repeating "+what+" · esc or i stops"))
}

func (m *Model) endRepeat() {
	if m.rep != nil && m.Player != nil {
		m.Player.Precise(false)
		if m.rep.gapping { // stopped during a pause between repeats: play on
			m.Player.SetPause(false)
		}
	}
	m.rep, m.abFrom = nil, 0
}

// repeatTick moves the repeat along with a new playback position.
func (m *Model) repeatTick(pos float64) tea.Cmd {
	m.rep.dur = m.dur
	a := m.rep.at(pos)
	switch {
	case a.end != "":
		m.endRepeat()
		return m.flash(a.end)
	case a.seek < 0:
		return nil
	}
	m.pos = a.seek
	if !a.gap {
		m.Player.SeekTo(a.seek)
		return nil
	}
	m.Player.SetPause(true)
	m.Player.SeekTo(a.seek)
	m.rep.gapping = true
	m.gapSeq++
	id := m.gapSeq
	return tea.Tick(time.Duration(m.rep.Gap)*time.Second, func(time.Time) tea.Msg { return gapMsg(id) })
}

func (m *Model) gapEnd(id int) {
	if m.rep != nil && m.rep.gapping && id == m.gapSeq {
		m.rep.gapping = false
		m.Player.SetPause(false)
	}
}

// markAyah is the quick A-B repeat: the first press marks the ayah playing,
// the second marks the end and starts repeating, the third stops.
func (m *Model) markAyah(string) tea.Cmd {
	switch {
	case m.rep != nil:
		m.endRepeat()
		return m.flash("Ayah repeat off")
	case m.kind != playSurah:
		return m.flash("Play a surah first: i marks the ayahs to repeat")
	}
	tr := m.track
	return m.withTimings(tr, func(m *Model, t []api.AyahTiming) tea.Cmd {
		if m.kind != playSurah || m.track != tr {
			return nil
		}
		a := 1 // the last ayah started by now; the opening counts as ayah 1
		for _, x := range t {
			if x.Ayah > 0 && m.pos >= float64(x.Start)/1000 {
				a = x.Ayah
			}
		}
		if m.abFrom == 0 {
			m.abFrom = a
			return m.flash(fmt.Sprintf("Repeat from ayah %d · press i again at the last ayah", a))
		}
		return m.startRepeat(tr, t, store.Hifz{From: min(m.abFrom, a), To: max(m.abFrom, a), Each: max(m.abEach, 1)})
	})
}

// cycleEach changes how many times each ayah plays, in the running repeat
// and for the next A-B repeat.
func (m *Model) cycleEach(string) tea.Cmd {
	cur := max(m.abEach, 1)
	if m.rep != nil {
		cur = m.rep.Each
	}
	next := eachSteps[0]
	for _, v := range eachSteps {
		if v > cur {
			next = v
			break
		}
	}
	m.abEach = next
	if m.rep != nil {
		m.rep.Each, m.rep.n = next, min(m.rep.n, next)
	}
	return m.flash("Each ayah plays " + plural(next, "time"))
}
