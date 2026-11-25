package ui

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/3mro777o/qari/internal/api"
	"github.com/3mro777o/qari/internal/norm"
	"github.com/3mro777o/qari/internal/popular"
	"github.com/3mro777o/qari/internal/store"
)

const (
	sortPopular = "popular"
	sortName    = "name"
	sortUpdated = "updated"
)

const (
	repeatNext = "next"
	repeatOne  = "one"
	repeatOff  = "off"
)

// applyPopularity merges popularity listen counts and known flags onto catalog reciters.
func applyPopularity(cat *api.Catalog, pop map[int]popular.Entry) {
	if cat == nil || len(pop) == 0 {
		return
	}
	for _, r := range cat.Reciters {
		if e, ok := pop[r.ID]; ok {
			r.Listens = e.Listens
			r.Known = e.Known
		}
	}
}

func isWellKnownComplete(r *api.ReciterEntry) bool {
	return r != nil && r.Complete && (r.Known || r.Listens > 0)
}

func normalizeSort(s string) string {
	switch s {
	case sortName, sortUpdated, sortPopular:
		return s
	default:
		return sortPopular
	}
}

func nextSort(s string) string {
	switch normalizeSort(s) {
	case sortPopular:
		return sortName
	case sortName:
		return sortUpdated
	default:
		return sortPopular
	}
}

func sortLabel(s string) string {
	switch normalizeSort(s) {
	case sortName:
		return "A-Z"
	case sortUpdated:
		return "Updated"
	default:
		return "Popular"
	}
}

func normalizeRepeat(s string) string {
	switch s {
	case repeatNext, repeatOne, repeatOff:
		return s
	default:
		return repeatNext
	}
}

func nextRepeat(s string) string {
	switch normalizeRepeat(s) {
	case repeatNext:
		return repeatOne
	case repeatOne:
		return repeatOff
	default:
		return repeatNext
	}
}

func repeatLabel(s string) string {
	switch normalizeRepeat(s) {
	case repeatOne:
		return "repeat"
	case repeatOff:
		return "stop"
	default:
		return "auto"
	}
}

func repeatLongLabel(s string) string {
	switch normalizeRepeat(s) {
	case repeatOne:
		return "Repeat surah"
	case repeatOff:
		return "Stop after surah"
	default:
		return "Autoplay next surah"
	}
}

// moshafStyleName returns the human-readable recitation style from MoshafType.
// MoshafType = riwayah*10 + style (1=Murattal, 2=Mujawwad, 3=Teaching).
func moshafStyleName(m api.Moshaf) string {
	switch m.MoshafType % 10 {
	case 1:
		return "Murattal"
	case 2:
		return "Mujawwad"
	case 3:
		return "Teaching"
	default:
		// Fallback to parsing after " - " in Name.
		if _, after, ok := strings.Cut(m.Name, " - "); ok {
			return strings.TrimSpace(after)
		}
		return ""
	}
}

func moshafStyleRank(m api.Moshaf) int {
	switch m.MoshafType % 10 {
	case 1:
		return 3 // Murattal preferred
	case 2:
		return 2 // Mujawwad
	case 3:
		return 1 // Teaching
	default:
		return 0
	}
}

// cleanRiwayahName trims redundant leading "Rewayat "/"Riwayat " prefix for cleaner badges.
func cleanRiwayahName(name string) string {
	name = strings.TrimSpace(name)
	for _, p := range []string{"Rewayat ", "Riwayat ", "rewayat ", "riwayat "} {
		if strings.HasPrefix(name, p) {
			return strings.TrimSpace(name[len(p):])
		}
	}
	return name
}

// moshafRiwayahName returns the cleaned riwayah name for m.
func moshafRiwayahName(cat *api.Catalog, m api.Moshaf) string {
	if cat != nil {
		if name := cat.Riwayat[m.RewayaID]; name != "" {
			return cleanRiwayahName(name)
		}
	}
	base := m.Name
	if before, _, ok := strings.Cut(base, " - "); ok {
		base = before
	}
	return cleanRiwayahName(base)
}

// shortRiwayahName extracts a compact label (e.g. "Hafs", "Warsh", "Qalun") for dense columns.
func shortRiwayahName(cat *api.Catalog, m api.Moshaf) string {
	full := moshafRiwayahName(cat, m)
	if full == "" {
		return ""
	}
	for _, cut := range []string{" A'n ", " An ", " 'An ", " an ", " - "} {
		if before, _, ok := strings.Cut(full, cut); ok && strings.TrimSpace(before) != "" {
			return strings.TrimSpace(before)
		}
	}
	return full
}

// moshafDisplayLabel builds a clear label like "Hafs A'n Assem · Murattal".
func moshafDisplayLabel(cat *api.Catalog, m api.Moshaf) string {
	r := moshafRiwayahName(cat, m)
	st := moshafStyleName(m)
	if r == "" {
		return m.Name
	}
	if st == "" || strings.Contains(strings.ToLower(r), strings.ToLower(st)) {
		return r
	}
	return r + " · " + st
}

// moshafCompactLabel builds a tight label like "Hafs · Murattal".
func moshafCompactLabel(cat *api.Catalog, m api.Moshaf) string {
	r := shortRiwayahName(cat, m)
	st := moshafStyleName(m)
	if r == "" {
		return m.Name
	}
	if st == "" || strings.Contains(strings.ToLower(r), strings.ToLower(st)) {
		return r
	}
	return r + " · " + st
}

func isHafs(cat *api.Catalog, m api.Moshaf) bool {
	if m.RewayaID == 1 {
		return true
	}
	if strings.Contains(strings.ToLower(m.Name), "hafs") {
		return true
	}
	if cat != nil && strings.Contains(strings.ToLower(cat.Riwayat[m.RewayaID]), "hafs") {
		return true
	}
	return false
}

// moshafSortedSurahs returns the sorted 1-based surah numbers contained in m.
func moshafSortedSurahs(m api.Moshaf) []int {
	set := m.Surahs()
	out := make([]int, 0, len(set))
	for n := range set {
		if n >= 1 && n <= 114 {
			out = append(out, n)
		}
	}
	slices.Sort(out)
	return out
}

// bestMoshaf selects the preferred recording for r:
// chosen riwayah > Hafs > complete (114 surahs) > Murattal style > higher surah count > lowest ID.
// If surahNum > 0, only moshafs containing surahNum are considered.
// If strictRewaya is true and preferredRewaya > 0, only moshafs with that RewayaID are eligible.
func bestMoshaf(cat *api.Catalog, r *api.ReciterEntry, preferredRewaya int, surahNum int, strictRewaya bool) (api.Moshaf, bool) {
	if r == nil || len(r.Moshaf) == 0 {
		return api.Moshaf{}, false
	}
	var best api.Moshaf
	found := false
	for _, m := range r.Moshaf {
		if strictRewaya && preferredRewaya > 0 && m.RewayaID != preferredRewaya {
			continue
		}
		if surahNum > 0 && !m.Surahs()[surahNum] {
			continue
		}
		if !found || betterMoshaf(cat, m, best, preferredRewaya) {
			best = m
			found = true
		}
	}
	return best, found
}

func betterMoshaf(cat *api.Catalog, a, b api.Moshaf, preferredRewaya int) bool {
	if preferredRewaya > 0 {
		aPref := a.RewayaID == preferredRewaya
		bPref := b.RewayaID == preferredRewaya
		if aPref != bPref {
			return aPref
		}
	}
	aHafs := isHafs(cat, a)
	bHafs := isHafs(cat, b)
	if aHafs != bHafs {
		return aHafs
	}
	aComp := a.SurahTotal >= 114
	bComp := b.SurahTotal >= 114
	if aComp != bComp {
		return aComp
	}
	aStyle := moshafStyleRank(a)
	bStyle := moshafStyleRank(b)
	if aStyle != bStyle {
		return aStyle > bStyle
	}
	if a.SurahTotal != b.SurahTotal {
		return a.SurahTotal > b.SurahTotal
	}
	return a.ID < b.ID
}

func findMoshafByID(r *api.ReciterEntry, id int) (api.Moshaf, bool) {
	if r == nil {
		return api.Moshaf{}, false
	}
	for _, m := range r.Moshaf {
		if m.ID == id {
			return m, true
		}
	}
	return api.Moshaf{}, false
}

func reciterHasRewaya(r *api.ReciterEntry, rewayaID int) bool {
	if rewayaID == 0 {
		return true
	}
	for _, m := range r.Moshaf {
		if m.RewayaID == rewayaID {
			return true
		}
	}
	return false
}

// filterQaris returns qaris matching scope, riwayah, and search query, deterministically sorted.
func filterQaris(cat *api.Catalog, favs map[int]bool, showAll bool, rewayaID int, sortMode, query string) []*api.ReciterEntry {
	if cat == nil {
		return nil
	}
	out := make([]*api.ReciterEntry, 0, len(cat.Reciters))
	for _, r := range cat.Reciters {
		if !showAll && !favs[r.ID] && !isWellKnownComplete(r) {
			continue
		}
		if !reciterHasRewaya(r, rewayaID) {
			continue
		}
		if !norm.Match(r.Search, query) {
			continue
		}
		out = append(out, r)
	}
	sortQaris(out, sortMode)
	return out
}

func sortQaris(list []*api.ReciterEntry, sortMode string) {
	mode := normalizeSort(sortMode)
	slices.SortStableFunc(list, func(a, b *api.ReciterEntry) int {
		switch mode {
		case sortName:
			if c := cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)); c != 0 {
				return c
			}
			return cmp.Compare(a.ID, b.ID)
		case sortUpdated:
			if c := cmp.Compare(b.Date, a.Date); c != 0 {
				return c
			}
			if c := cmp.Compare(b.Listens, a.Listens); c != 0 {
				return c
			}
			if c := cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)); c != 0 {
				return c
			}
			return cmp.Compare(a.ID, b.ID)
		default: // sortPopular
			if c := cmp.Compare(b.Listens, a.Listens); c != 0 {
				return c
			}
			if a.Known != b.Known {
				if a.Known {
					return -1
				}
				return 1
			}
			if c := cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)); c != 0 {
				return c
			}
			return cmp.Compare(a.ID, b.ID)
		}
	})
}

// surahQariItem binds a reciter with their best recording for a given surah.
type surahQariItem struct {
	Reciter *api.ReciterEntry
	Moshaf  api.Moshaf
}

func filterQarisForSurah(cat *api.Catalog, favs map[int]bool, surahID int, showAll bool, rewayaID int, sortMode, query string) []surahQariItem {
	if cat == nil || surahID < 1 || surahID > 114 {
		return nil
	}
	var reciters []*api.ReciterEntry
	moshafs := make(map[int]api.Moshaf)
	for _, r := range cat.Reciters {
		if !showAll && !favs[r.ID] && !isWellKnownComplete(r) {
			continue
		}
		if !norm.Match(r.Search, query) {
			continue
		}
		m, ok := bestMoshaf(cat, r, rewayaID, surahID, rewayaID > 0)
		if !ok {
			continue
		}
		reciters = append(reciters, r)
		moshafs[r.ID] = m
	}
	sortQaris(reciters, sortMode)
	out := make([]surahQariItem, len(reciters))
	for i, r := range reciters {
		out[i] = surahQariItem{Reciter: r, Moshaf: moshafs[r.ID]}
	}
	return out
}

// matchSurah checks English/Arabic/display names AND surah number (e.g. "18", "018").
func matchSurah(s api.SurahEntry, query string) bool {
	q := strings.TrimSpace(query)
	if q == "" {
		return true
	}
	hay := fmt.Sprintf("%d %03d %s", s.ID, s.ID, s.Search)
	return norm.Match(hay, q)
}

func filterSurahs(cat *api.Catalog, query string) []api.SurahEntry {
	if cat == nil {
		return nil
	}
	q := strings.TrimSpace(query)
	exactNum, errNum := strconv.Atoi(q)
	out := make([]api.SurahEntry, 0, len(cat.Surahs))
	for _, s := range cat.Surahs {
		if s.ID == 0 {
			continue
		}
		if matchSurah(s, q) {
			out = append(out, s)
		}
	}
	// If the user typed a surah number like "18" or "1", float that exact surah to the top.
	if errNum == nil && exactNum >= 1 && exactNum <= 114 {
		for i, s := range out {
			if s.ID == exactNum && i > 0 {
				copy(out[1:i+1], out[0:i])
				out[0] = s
				break
			}
		}
	}
	return out
}

func filterMoshafSurahs(cat *api.Catalog, m api.Moshaf, query string) []api.SurahEntry {
	if cat == nil {
		return nil
	}
	q := strings.TrimSpace(query)
	exactNum, errNum := strconv.Atoi(q)
	ids := moshafSortedSurahs(m)
	out := make([]api.SurahEntry, 0, len(ids))
	for _, id := range ids {
		if id < 1 || id > len(cat.Surahs) {
			continue
		}
		s := cat.Surahs[id-1]
		if s.ID == 0 {
			s.ID = id
		}
		if matchSurah(s, q) {
			out = append(out, s)
		}
	}
	if errNum == nil && exactNum >= 1 && exactNum <= 114 {
		for i, s := range out {
			if s.ID == exactNum && i > 0 {
				copy(out[1:i+1], out[0:i])
				out[0] = s
				break
			}
		}
	}
	return out
}

type riwayaOption struct {
	ID    int
	Name  string
	Count int
}

func buildRiwayaOptions(cat *api.Catalog) []riwayaOption {
	if cat == nil {
		return []riwayaOption{{ID: 0, Name: "All riwayat", Count: 0}}
	}
	counts := map[int]int{}
	for _, r := range cat.Reciters {
		seen := map[int]bool{}
		for _, m := range r.Moshaf {
			if m.RewayaID > 0 && !seen[m.RewayaID] {
				seen[m.RewayaID] = true
				counts[m.RewayaID]++
			}
		}
	}
	opts := make([]riwayaOption, 0, len(counts)+1)
	for id, c := range counts {
		name := cleanRiwayahName(cat.Riwayat[id])
		if name == "" {
			name = fmt.Sprintf("Riwayah %d", id)
		}
		opts = append(opts, riwayaOption{ID: id, Name: name, Count: c})
	}
	slices.SortStableFunc(opts, func(a, b riwayaOption) int {
		if c := cmp.Compare(b.Count, a.Count); c != 0 {
			return c
		}
		return cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
	})
	return append([]riwayaOption{{ID: 0, Name: "All riwayat", Count: len(cat.Reciters)}}, opts...)
}

func riwayaFilterName(cat *api.Catalog, id int) string {
	if id == 0 {
		return "All"
	}
	if cat != nil {
		if n := cleanRiwayahName(cat.Riwayat[id]); n != "" {
			for _, cut := range []string{" A'n ", " An ", " 'An "} {
				if before, _, ok := strings.Cut(n, cut); ok && strings.TrimSpace(before) != "" {
					return strings.TrimSpace(before)
				}
			}
			return n
		}
	}
	return fmt.Sprintf("#%d", id)
}

func surahProgressSummary(st *store.Store, surahID int) (finished bool, inProgress bool, bestFrac float64) {
	if st == nil {
		return false, false, 0
	}
	for _, p := range st.Progress {
		if p.Surah != surahID {
			continue
		}
		if p.Finished {
			finished = true
			bestFrac = 1
		} else if p.Position >= 5 {
			inProgress = true
			if f := p.Fraction(); f > bestFrac {
				bestFrac = f
			}
		}
	}
	return finished, inProgress, bestFrac
}

func moshafProgressCounts(st *store.Store, reciterID, moshafID int) (finished, inProgress int) {
	if st == nil {
		return 0, 0
	}
	for _, p := range st.Progress {
		if p.Reciter != reciterID || p.Moshaf != moshafID {
			continue
		}
		if p.Finished {
			finished++
		} else if p.Position >= 5 {
			inProgress++
		}
	}
	return finished, inProgress
}

func reciterProgressCounts(st *store.Store, reciterID int) (finished, inProgress int) {
	if st == nil {
		return 0, 0
	}
	finSurahs := map[int]bool{}
	progSurahs := map[int]bool{}
	for _, p := range st.Progress {
		if p.Reciter != reciterID {
			continue
		}
		if p.Finished {
			finSurahs[p.Surah] = true
		} else if p.Position >= 5 {
			progSurahs[p.Surah] = true
		}
	}
	for s := range finSurahs {
		delete(progSurahs, s)
	}
	return len(finSurahs), len(progSurahs)
}

func revelationLabel(makkia int) string {
	if makkia == 1 {
		return "Meccan"
	}
	return "Medinan"
}

func surahPagesLabel(s api.SurahEntry) string {
	if s.StartPage <= 0 {
		return ""
	}
	if s.EndPage <= s.StartPage {
		return fmt.Sprintf("p.%d", s.StartPage)
	}
	return fmt.Sprintf("pp.%d-%d", s.StartPage, s.EndPage)
}

// shortDate normalizes API date strings (e.g. "2024-09-12T10:00:00Z" or "2024-09-12") to "2024-09".
func shortDate(d string) string {
	d = strings.TrimSpace(d)
	if len(d) >= 7 {
		return d[:7]
	}
	return d
}
