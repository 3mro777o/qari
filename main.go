// qari is a terminal Quran player: browse reciters and surahs from
// mp3quran.net, and listen through mpv with progress remembered.
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/3mro777o/qari/internal/api"
	"github.com/3mro777o/qari/internal/player"
	"github.com/3mro777o/qari/internal/popular"
	"github.com/3mro777o/qari/internal/store"
	"github.com/3mro777o/qari/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

var version = "dev"

func main() {
	lang := flag.String("lang", envOr("QARI_LANG", "eng"), "language for reciter and surah names (see --languages)")
	proxy := flag.String("proxy", os.Getenv("QARI_PROXY"), "proxy URL for all traffic, e.g. socks5://127.0.0.1:1080")
	noArabic := flag.Bool("no-arabic", false, "hide Arabic names (for terminals without Arabic text support)")
	languages := flag.Bool("languages", false, "list the name languages the API offers and exit")
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "qari %s: listen to the Quran in your terminal\n\nUsage: qari [flags]\n\n", version)
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nPress ? inside the app for keys. Needs mpv for playback.")
	}
	flag.Parse()
	if version == "dev" { // installed with go install: use the module version
		if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
			version = bi.Main.Version
		}
	}
	if *showVersion {
		fmt.Println("qari", version)
		return
	}

	httpClient, err := newHTTPClient(*proxy)
	if err != nil {
		fail(err)
	}
	cacheDir, stateFile := paths()
	client := api.New(httpClient, cacheDir)

	if *languages {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		ls, err := client.Languages(ctx)
		if err != nil {
			fail(err)
		}
		for _, l := range ls {
			fmt.Printf("%-5s %s (%s)\n", l.Locale, l.Language, l.Native)
		}
		return
	}

	st, err := store.Open(stateFile)
	if err != nil {
		fail(err)
	}

	var p *player.Player
	if player.Available() {
		var relay *player.Relay
		if *proxy != "" || os.Getenv("HTTPS_PROXY") != "" || os.Getenv("https_proxy") != "" {
			if relay, err = player.NewRelay(httpClient); err != nil {
				fail(err)
			}
		}
		p = player.New(relay)
	}

	m := ui.New(ui.Options{
		Client: client, Lang: *lang, Store: st, Player: p,
		Popular: popular.Load(), NoArabic: *noArabic, Version: version,
	})
	prog := tea.NewProgram(m, tea.WithAltScreen())
	// Closing the terminal (SIGHUP) or a kill (SIGTERM) must not leave mpv
	// playing on its own.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGTERM)
	go func() { <-sig; prog.Quit() }()
	_, err = prog.Run()
	if p != nil {
		p.Close()
	}
	if serr := st.Save(); serr != nil && err == nil {
		err = serr
	}
	if err != nil {
		fail(err)
	}
}

func newHTTPClient(proxy string) (*http.Client, error) {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	if proxy != "" {
		u, err := url.Parse(proxy)
		if err != nil || u.Host == "" {
			return nil, fmt.Errorf("invalid --proxy %q (want e.g. socks5://127.0.0.1:1080 or http://host:port)", proxy)
		}
		tr.Proxy = http.ProxyURL(u)
	}
	// No overall timeout: audio streams through this client for a long time.
	tr.ResponseHeaderTimeout = 30 * time.Second
	return &http.Client{Transport: tr}, nil
}

// paths follows the XDG layout: the catalog is cache, progress is state.
func paths() (cacheDir, stateFile string) {
	cache, err := os.UserCacheDir()
	if err != nil {
		cache = os.TempDir()
	}
	state := os.Getenv("XDG_STATE_HOME")
	if state == "" {
		if home, err := os.UserHomeDir(); err == nil {
			state = filepath.Join(home, ".local", "state")
		} else {
			state = cache
		}
	}
	return filepath.Join(cache, "qari"), filepath.Join(state, "qari", "state.json")
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "qari:", err)
	os.Exit(1)
}
