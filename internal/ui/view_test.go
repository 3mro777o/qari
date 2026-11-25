package ui

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/3mro777o/qari/internal/api"
	"github.com/3mro777o/qari/internal/popular"
	"github.com/3mro777o/qari/internal/store"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

// TestViewFitsScreen renders every tab from the on-disk API cache and checks
// that the frame is exactly the terminal size.
func TestViewFitsScreen(t *testing.T) {
	cache, _ := os.UserCacheDir()
	// Read-only use of the user's cache: requests always fail, so nothing is
	// fetched or written and stale entries are served as offline data.
	noNet := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("no network in tests")
	})}
	client := api.New(noNet, filepath.Join(cache, "qari"))
	cat, err := api.LoadCatalog(context.Background(), client, "eng")
	if err != nil {
		t.Skip("no cached catalog:", err)
	}
	st, _ := store.Open(filepath.Join(t.TempDir(), "s.json"))
	for _, size := range [][2]int{{110, 30}, {64, 18}, {160, 45}, {60, 16}} {
		w, h := size[0], size[1]
		var mm tea.Model = New(Options{Client: client, Store: st, Popular: popular.Load()})
		mm, _ = mm.Update(tea.WindowSizeMsg{Width: w, Height: h})
		mm, _ = mm.Update(catalogMsg{cat: cat})
		for _, key := range "1234?" {
			mm, _ = mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{key}})
			lines := strings.Split(mm.View(), "\n")
			if key != '?' && len(lines) != h {
				t.Errorf("%dx%d tab %c: %d lines, want %d", w, h, key, len(lines), h)
			}
			if len(lines) > h {
				t.Errorf("%dx%d help: %d lines > %d", w, h, len(lines), h)
			}
			for i, l := range lines {
				if lw := ansi.StringWidth(l); lw > w {
					t.Errorf("%dx%d key %c line %d: width %d: %q", w, h, key, i, lw, ansi.Strip(l))
				}
			}
			if key == '?' {
				mm, _ = mm.Update(tea.KeyMsg{Type: tea.KeyEsc})
			}
		}
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
