package api

import (
	"context"
	"strings"
	"sync"
	"unicode"

	"github.com/3mro777o/qari/internal/norm"
)

// ReciterEntry is a reciter with names merged from several API languages.
type ReciterEntry struct {
	Reciter         // names in the display language
	Arabic   string // Arabic name, always searchable
	Search   string // folded names in every loaded language
	Complete bool   // at least one moshaf has all 114 surahs
	Listens  int    // popularity score (0 = unknown)
	Known    bool   // well-known reciter, see internal/popular
}

// SurahEntry is a surah with its Arabic name and search key.
type SurahEntry struct {
	Surah
	Arabic string
	Search string
}

type Catalog struct {
	Lang     string
	Reciters []*ReciterEntry
	Surahs   []SurahEntry   // Surahs[i] is surah i+1
	Riwayat  map[int]string // rewaya id -> name
	Radios   []Radio
	byID     map[int]*ReciterEntry
}

func (c *Catalog) Reciter(id int) *ReciterEntry { return c.byID[id] }

// SurahName returns the display name of surah n (1-based).
func (c *Catalog) SurahName(n int) string {
	if n >= 1 && n <= len(c.Surahs) {
		return c.Surahs[n-1].Name
	}
	return "?"
}

// LoadCatalog fetches everything the UI needs. Names are fetched in English,
// Arabic and lang so search works in any of them; lang is used for display.
func LoadCatalog(ctx context.Context, c *Client, lang string) (*Catalog, error) {
	langs := []string{"eng", "ar"}
	if lang != "eng" && lang != "ar" {
		langs = append(langs, lang)
	}
	var (
		mu       sync.Mutex
		wg       sync.WaitGroup
		firstErr error
		reciters = map[string][]Reciter{}
		suwar    = map[string][]Surah{}
		riwayat  []Riwaya
		radios   []Radio
	)
	run := func(f func() error) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := f(); err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}()
	}
	for _, l := range langs {
		run(func() error {
			r, err := c.Reciters(ctx, l)
			mu.Lock()
			reciters[l] = r
			mu.Unlock()
			return err
		})
		run(func() error {
			s, err := c.Suwar(ctx, l)
			mu.Lock()
			suwar[l] = s
			mu.Unlock()
			return err
		})
	}
	run(func() (err error) { riwayat, err = c.Riwayat(ctx, lang); return })
	run(func() (err error) { radios, err = c.Radios(ctx, lang); return })
	wg.Wait()
	// Only English is required; other languages just add names.
	if len(reciters["eng"]) == 0 || len(suwar["eng"]) == 0 {
		return nil, firstErr
	}
	display := lang
	if len(reciters[display]) == 0 || len(suwar[display]) == 0 {
		display = "eng"
	}

	cat := &Catalog{Lang: display, Riwayat: map[int]string{}, Radios: radios, byID: map[int]*ReciterEntry{}}
	names := map[int][]string{}
	arabic := map[int]string{}
	for _, l := range langs {
		for _, r := range reciters[l] {
			names[r.ID] = append(names[r.ID], r.Name)
			if l == "ar" {
				arabic[r.ID] = r.Name
			}
		}
	}
	for _, r := range reciters[display] {
		e := &ReciterEntry{Reciter: r, Arabic: arabic[r.ID], Search: norm.Fold(strings.Join(names[r.ID], " "))}
		e.Name, e.Arabic = clean(e.Name), clean(e.Arabic)
		for _, m := range r.Moshaf {
			if m.SurahTotal >= 114 {
				e.Complete = true
			}
		}
		cat.Reciters = append(cat.Reciters, e)
		cat.byID[r.ID] = e
	}

	snames := map[int][]string{}
	sarabic := map[int]string{}
	for _, l := range langs {
		for _, s := range suwar[l] {
			snames[s.ID] = append(snames[s.ID], s.Name)
			if l == "ar" {
				sarabic[s.ID] = clean(s.Name)
			}
		}
	}
	cat.Surahs = make([]SurahEntry, 114)
	for _, s := range suwar[display] {
		if s.ID < 1 || s.ID > 114 {
			continue
		}
		s.Name = clean(s.Name)
		cat.Surahs[s.ID-1] = SurahEntry{Surah: s, Arabic: sarabic[s.ID], Search: norm.Fold(strings.Join(snames[s.ID], " "))}
	}
	for _, r := range riwayat {
		cat.Riwayat[r.ID] = r.Name
	}
	return cat, nil
}

// clean drops invisible formatting characters the API sometimes includes
// (soft hyphens, direction marks) and collapses runs of spaces.
func clean(s string) string {
	s = strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Cf, r) {
			return -1
		}
		return r
	}, s)
	return strings.Join(strings.Fields(s), " ")
}
