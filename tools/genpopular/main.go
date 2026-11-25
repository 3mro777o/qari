// genpopular builds internal/popular/popular.json: a listen count and a
// "well-known" flag for each mp3quran reciter.
//
// mp3quran.net publishes no play counts, so the data comes from elsewhere:
//   - listens: way2quran.com reciter pages, sorted by total listens
//   - known:   the reciter is also published by quranicaudio.com or
//     quran.com, or way2quran marks them as a top reciter
//
// Reciters are matched across sites by their folded Arabic, then English name.
// Run: go run ./tools/genpopular   (honours HTTPS_PROXY, incl. socks5://)
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/3mro777o/qari/internal/norm"
)

type entry struct {
	Listens int  `json:"listens,omitempty"`
	Known   bool `json:"known,omitempty"`
}

type output struct {
	Generated string           `json:"generated"`
	Sources   []string         `json:"sources"`
	Reciters  map[string]entry `json:"reciters"` // key: mp3quran reciter id
}

var client = &http.Client{Timeout: 60 * time.Second}

func get(url string) []byte {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "qari-genpopular (+https://github.com/3mro777o/qari)")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("%s: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Fatalf("%s: HTTP %d", url, resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return b
}

type w2q struct {
	Arabic  string `json:"arabicName"`
	English string `json:"englishName"`
	Slug    string `json:"slug"`
	Listens int    `json:"totalListens"`
	Top     bool   `json:"isTopReciter"`
}

var rscChunk = regexp.MustCompile(`self\.__next_f\.push\(\[1,("(?:[^"\\]|\\.)*")\]\)`)

// way2quranPage extracts the reciter objects embedded in a Next.js page.
func way2quranPage(html []byte) []w2q {
	var payload strings.Builder
	for _, m := range rscChunk.FindAllSubmatch(html, -1) {
		var s string
		if json.Unmarshal(m[1], &s) == nil {
			payload.WriteString(s)
		}
	}
	text := payload.String()
	var out []w2q
	for i := 0; ; {
		j := strings.Index(text[i:], `{"_id":`)
		if j < 0 {
			break
		}
		i += j
		dec := json.NewDecoder(strings.NewReader(text[i:]))
		var raw map[string]json.RawMessage
		if dec.Decode(&raw) == nil {
			if _, ok := raw["totalListens"]; ok {
				var r w2q
				b, _ := json.Marshal(raw)
				json.Unmarshal(b, &r)
				out = append(out, r)
			}
		}
		i++
	}
	return out
}

type person struct {
	id     int
	ar, en string
}

// matcher finds the mp3quran reciter for a name from another site.
type matcher struct {
	byAr, byEn map[string][]int
	people     []person
}

func newMatcher(ps []person) *matcher {
	m := &matcher{byAr: map[string][]int{}, byEn: map[string][]int{}, people: ps}
	for _, p := range ps {
		m.byAr[key(p.ar)] = append(m.byAr[key(p.ar)], p.id)
		m.byEn[key(p.en)] = append(m.byEn[key(p.en)], p.id)
	}
	return m
}

// key drops "ال"/"al" articles and spaces so word splits and articles don't matter.
func key(s string) string {
	var parts []string
	for _, w := range strings.Fields(norm.Fold(s)) {
		w = strings.TrimPrefix(w, "ال")
		if strings.HasPrefix(w, "al") && len(w) > 4 {
			w = w[2:]
		}
		parts = append(parts, w)
	}
	// "عبدالباسط" is often written "عبد الباسط"
	return strings.ReplaceAll(strings.Join(parts, ""), "عبدال", "عبد")
}

func (m *matcher) find(ar, en string) int {
	if ids := m.byAr[key(ar)]; ar != "" && len(ids) == 1 {
		return ids[0]
	}
	if ids := m.byEn[key(en)]; en != "" && len(ids) == 1 {
		return ids[0]
	}
	// All words of the shorter Arabic name appear in exactly one candidate.
	words := strings.Fields(norm.Fold(ar))
	if len(words) < 2 {
		return 0
	}
	found := 0
	for _, p := range m.people {
		have := strings.Fields(norm.Fold(p.ar))
		if subset(words, have) || subset(have, words) && len(have) >= 2 {
			if found != 0 {
				return 0 // ambiguous
			}
			found = p.id
		}
	}
	return found
}

func subset(a, b []string) bool {
	set := map[string]bool{}
	for _, w := range b {
		set[strings.TrimPrefix(w, "ال")] = true
	}
	for _, w := range a {
		if !set[strings.TrimPrefix(w, "ال")] {
			return false
		}
	}
	return true
}

func main() {
	pages := flag.Int("pages", 8, "way2quran pages to read (most listened first)")
	out := flag.String("o", "internal/popular/popular.json", "output file")
	flag.Parse()

	var ar, en struct {
		Reciters []struct {
			ID   int
			Name string
		}
	}
	json.Unmarshal(get("https://mp3quran.net/api/v3/reciters?language=ar"), &ar)
	json.Unmarshal(get("https://mp3quran.net/api/v3/reciters?language=eng"), &en)
	enName := map[int]string{}
	for _, r := range en.Reciters {
		enName[r.ID] = r.Name
	}
	var ps []person
	for _, r := range ar.Reciters {
		ps = append(ps, person{r.ID, r.Name, enName[r.ID]})
	}
	if len(ps) == 0 {
		log.Fatal("no reciters from mp3quran")
	}
	m := newMatcher(ps)
	res := map[int]entry{}
	unmatched := 0

	for p := 1; p <= *pages; p++ {
		url := fmt.Sprintf("https://www.way2quran.com/en/reciters?sortBy=-totalListens&currentPage=%d", p)
		rs := way2quranPage(get(url))
		log.Printf("way2quran page %d: %d reciters", p, len(rs))
		for _, r := range rs {
			id := m.find(r.Arabic, r.English)
			if id == 0 {
				unmatched++
				continue
			}
			e := res[id]
			e.Listens = max(e.Listens, r.Listens)
			e.Known = e.Known || r.Top || r.Listens >= 10000
			res[id] = e
		}
	}

	var qa []struct {
		Name   string `json:"name"`
		Arabic string `json:"arabic_name"`
	}
	json.Unmarshal(get("https://quranicaudio.com/api/qaris"), &qa)
	for _, q := range qa {
		if id := m.find(q.Arabic, q.Name); id != 0 {
			e := res[id]
			e.Known = true
			res[id] = e
		}
	}
	var qc struct {
		Recitations []struct {
			Name string `json:"reciter_name"`
		} `json:"recitations"`
	}
	json.Unmarshal(get("https://api.quran.com/api/v4/resources/recitations"), &qc)
	for _, q := range qc.Recitations {
		if id := m.find("", q.Name); id != 0 {
			e := res[id]
			e.Known = true
			res[id] = e
		}
	}

	o := output{
		Generated: time.Now().UTC().Format("2006-01-02"),
		Sources:   []string{"way2quran.com (listens)", "quranicaudio.com", "quran.com"},
		Reciters:  map[string]entry{},
	}
	ids := make([]int, 0, len(res))
	for id := range res {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	known := 0
	for _, id := range ids {
		o.Reciters[fmt.Sprint(id)] = res[id]
		if res[id].Known {
			known++
		}
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", " ")
	enc.Encode(o)
	if err := os.WriteFile(*out, buf.Bytes(), 0o644); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote %s: %d reciters matched (%d known), %d way2quran names unmatched", *out, len(res), known, unmatched)
}
