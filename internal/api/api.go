// Package api talks to the public mp3quran.net v3 API and caches its answers
// on disk so the app starts instantly and keeps working offline.
package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

const (
	BaseURL  = "https://mp3quran.net/api/v3/"
	cacheTTL = 24 * time.Hour
)

type Moshaf struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	RewayaID   int    `json:"rewaya_id"`
	Server     string `json:"server"`
	SurahTotal int    `json:"surah_total"`
	MoshafType int    `json:"moshaf_type"`
	SurahList  string `json:"surah_list"`
}

// Surahs parses SurahList ("1,2,3") into a set.
func (m Moshaf) Surahs() map[int]bool {
	set := make(map[int]bool, m.SurahTotal)
	for _, s := range strings.Split(m.SurahList, ",") {
		if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
			set[n] = true
		}
	}
	return set
}

// URL is the mp3 address of one surah in this moshaf.
func (m Moshaf) URL(surah int) string {
	return fmt.Sprintf("%s%03d.mp3", strings.TrimRight(m.Server, "/")+"/", surah)
}

type Reciter struct {
	ID     int      `json:"id"`
	Name   string   `json:"name"`
	Moshaf []Moshaf `json:"moshaf"`
	Date   string   `json:"date"` // when the reciter was last updated
}

type Surah struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Makkia    int    `json:"makkia"`
	StartPage int    `json:"start_page"`
	EndPage   int    `json:"end_page"`
}

type Riwaya struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Radio struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Language struct {
	Locale   string `json:"locale"`
	Language string `json:"language"`
	Native   string `json:"native"`
}

type Client struct {
	HTTP     *http.Client
	CacheDir string
	offline  atomic.Bool // set when a request fell back to stale cache
}

// Offline reports whether any answer came from an expired cache because the
// network failed.
func (c *Client) Offline() bool { return c.offline.Load() }

func New(httpClient *http.Client, cacheDir string) *Client {
	return &Client{HTTP: httpClient, CacheDir: cacheDir}
}

// get returns the body of an API endpoint: fresh cache, then network, then
// stale cache if the network fails.
func (c *Client) get(ctx context.Context, endpoint string, q url.Values) ([]byte, error) {
	key := endpoint
	if len(q) > 0 {
		key += "_" + q.Encode()
	}
	path := filepath.Join(c.CacheDir, strings.NewReplacer("/", "_", "=", "-", "&", "_").Replace(key)+".json")
	if st, err := os.Stat(path); err == nil && time.Since(st.ModTime()) < cacheTTL {
		if b, err := os.ReadFile(path); err == nil {
			return b, nil
		}
	}
	b, err := c.fetch(ctx, endpoint, q)
	if err != nil {
		if old, rerr := os.ReadFile(path); rerr == nil {
			c.offline.Store(true)
			return old, nil
		}
		return nil, err
	}
	if err := os.MkdirAll(c.CacheDir, 0o755); err == nil {
		tmp := path + ".tmp"
		if os.WriteFile(tmp, b, 0o644) == nil {
			os.Rename(tmp, path)
		}
	}
	return b, nil
}

func (c *Client) fetch(ctx context.Context, endpoint string, q url.Values) ([]byte, error) {
	u := BaseURL + endpoint
	if len(q) > 0 {
		u += "?" + q.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: HTTP %d", endpoint, resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return nil, err
	}
	if !json.Valid(b) {
		return nil, errors.New(endpoint + ": response is not JSON (blocked by a network filter?)")
	}
	return b, nil
}

func getInto[T any](ctx context.Context, c *Client, endpoint, field, lang string) ([]T, error) {
	q := url.Values{}
	if lang != "" {
		q.Set("language", lang)
	}
	b, err := c.get(ctx, endpoint, q)
	if err != nil {
		return nil, err
	}
	var wrap map[string]json.RawMessage
	if err := json.Unmarshal(b, &wrap); err != nil {
		return nil, fmt.Errorf("%s: %w", endpoint, err)
	}
	var out []T
	if err := json.Unmarshal(wrap[field], &out); err != nil {
		return nil, fmt.Errorf("%s: %w", endpoint, err)
	}
	return out, nil
}

func (c *Client) Reciters(ctx context.Context, lang string) ([]Reciter, error) {
	return getInto[Reciter](ctx, c, "reciters", "reciters", lang)
}
func (c *Client) Suwar(ctx context.Context, lang string) ([]Surah, error) {
	return getInto[Surah](ctx, c, "suwar", "suwar", lang)
}
func (c *Client) Riwayat(ctx context.Context, lang string) ([]Riwaya, error) {
	return getInto[Riwaya](ctx, c, "riwayat", "riwayat", lang)
}
func (c *Client) Radios(ctx context.Context, lang string) ([]Radio, error) {
	return getInto[Radio](ctx, c, "radios", "radios", lang)
}
func (c *Client) Languages(ctx context.Context) ([]Language, error) {
	return getInto[Language](ctx, c, "languages", "language", "")
}
