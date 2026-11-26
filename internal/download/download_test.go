package download

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/3mro777o/qari/internal/store"
)

func TestDownloadSaveFailAndDelete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/missing.mp3" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "audio/mpeg")
		w.Write(make([]byte, 100_000))
	}))
	defer srv.Close()
	m := New(t.TempDir(), srv.Client())
	ok, bad := store.Track{Reciter: 1, Moshaf: 2, Surah: 108}, store.Track{Reciter: 1, Moshaf: 2, Surah: 1}
	m.Add(ok, srv.URL+"/108.mp3")
	m.Add(bad, srv.URL+"/missing.mp3")
	if m.Add(ok, srv.URL+"/108.mp3") {
		t.Fatal("same surah queued twice")
	}
	deadline := time.Now().Add(5 * time.Second)
	for len(m.Jobs()) != 1 || m.Jobs()[0].State != Failed {
		if time.Now().After(deadline) {
			t.Fatalf("downloads did not finish: %+v", m.Jobs())
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !m.Has(ok) || m.Has(bad) {
		t.Fatal("wrong saved set")
	}
	if _, err := os.Stat(m.Path(bad) + ".part"); !os.IsNotExist(err) {
		t.Fatal("failed download left a partial file")
	}
	if g := New(m.Dir, nil).Saved(); len(g) != 1 || g[0].Surahs != 1 || g[0].Bytes != 100_000 {
		t.Fatalf("rescan found %+v", g)
	}
	if err := m.Delete(ok); err != nil || m.Has(ok) {
		t.Fatal("delete failed")
	}
}
