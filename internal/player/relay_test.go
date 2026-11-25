package player

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRelay(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Range") != "bytes=2-" {
			t.Errorf("Range not forwarded: %q", r.Header.Get("Range"))
		}
		w.Header().Set("Content-Range", "bytes 2-4/5")
		w.WriteHeader(http.StatusPartialContent)
		io.WriteString(w, "llo")
	}))
	defer up.Close()
	r, err := NewRelay(http.DefaultClient)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest("GET", r.Wrap(up.URL+"/a.mp3"), nil)
	req.Header.Set("Range", "bytes=2-")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 206 || string(body) != "llo" || resp.Header.Get("Content-Range") == "" {
		t.Errorf("got %d %q %v", resp.StatusCode, body, resp.Header)
	}
	// Without the token the relay refuses.
	bad := strings.Replace(r.Wrap(up.URL), r.token, "x", 1)
	if resp, _ := http.Get(bad); resp.StatusCode != 404 {
		t.Errorf("request without token: %d, want 404", resp.StatusCode)
	}
}
