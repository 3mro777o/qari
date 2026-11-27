package api

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadForMatchesServer(t *testing.T) {
	reads := []TimingRead{{1, "https://cdn.mp3quran.net/audio/ibrahim-akhdar/r1/"}, {5, "https://www.Server8.mp3quran.net/afs"}}
	for server, want := range map[string]int{
		"https://cdn.mp3quran.net/audio/ibrahim-akhdar/r1/": 1,
		"http://cdn.mp3quran.net/audio/ibrahim-akhdar/r1":   1,
		"https://server8.mp3quran.net/afs/":                 5,
		"https://cdn.mp3quran.net/audio/ibrahim-akhdar/r2/": 0,
		"": 0,
	} {
		if got := ReadFor(reads, Moshaf{Server: server}); got != want {
			t.Errorf("%q: read %d, want %d", server, got, want)
		}
	}
}

func TestAyahTimingsFromCache(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "ayat_timing_read-1_surah-112.json"),
		[]byte(`[{"ayah":1,"start_time":7895,"end_time":11168},{"ayah":0,"start_time":0,"end_time":7895}]`), 0o644)
	os.WriteFile(filepath.Join(dir, "ayat_timing_reads.json"), []byte(`[{"id":1,"folder_url":"https://x/a/"}]`), 0o644)
	c := New(nil, dir) // no HTTP client: only the cache can answer
	tm, err := c.AyahTimings(t.Context(), 1, 112)
	if err != nil || len(tm) != 2 || tm[0].Ayah != 0 || tm[1].Start != 7895 {
		t.Fatalf("timings %+v, %v", tm, err)
	}
	if r, err := c.TimingReads(t.Context()); err != nil || len(r) != 1 || r[0].ID != 1 {
		t.Fatalf("reads %+v, %v", r, err)
	}
}
