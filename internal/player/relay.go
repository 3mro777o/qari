package player

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net"
	"net/http"
	"net/url"
)

// Relay is a tiny local HTTP server that fetches audio through the app's
// HTTP client (and so its proxy) for mpv, which cannot speak SOCKS.
// Range requests are passed through, so seeking still works.
type Relay struct {
	client *http.Client
	addr   string
	token  string // keeps other local programs from using the relay
}

func NewRelay(client *http.Client) (*Relay, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		ln.Close()
		return nil, err
	}
	r := &Relay{client: client, addr: ln.Addr().String(), token: hex.EncodeToString(b)}
	go http.Serve(ln, r)
	return r, nil
}

func (r *Relay) Wrap(u string) string {
	return "http://" + r.addr + "/" + r.token + "?u=" + url.QueryEscape(u)
}

func (r *Relay) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/"+r.token {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	target := req.URL.Query().Get("u")
	if u, err := url.Parse(target); err != nil || (u.Scheme != "https" && u.Scheme != "http") {
		http.Error(w, "bad url", http.StatusBadRequest)
		return
	}
	out, _ := http.NewRequestWithContext(req.Context(), req.Method, target, nil)
	if rg := req.Header.Get("Range"); rg != "" {
		out.Header.Set("Range", rg)
	}
	resp, err := r.client.Do(out)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	for _, h := range []string{"Content-Type", "Content-Length", "Content-Range", "Accept-Ranges", "icy-name"} {
		if v := resp.Header.Get(h); v != "" {
			w.Header().Set(h, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
