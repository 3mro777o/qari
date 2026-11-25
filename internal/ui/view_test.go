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
	var mm tea.Model = New(Options{Client: client, Store: st, Popular: popular.Load()})
	mm, _ = mm.Update(tea.WindowSizeMsg{Width: 110, Height: 30})
	mm, _ = mm.Update(catalogMsg{cat: cat})
	for tab := '1'; tab <= '4'; tab++ {
		mm, _ = mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{tab}})
		lines := strings.Split(mm.View(), "\n")
		if len(lines) != 30 {
			t.Errorf("tab %c: %d lines, want 30", tab, len(lines))
		}
		for i, l := range lines {
			if w := ansi.StringWidth(l); w > 110 {
				t.Errorf("tab %c line %d: width %d > 110: %q", tab, i, w, ansi.Strip(l))
			}
		}
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
