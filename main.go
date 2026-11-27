// qari is a terminal Quran player: browse reciters and surahs from
// mp3quran.net, and listen through mpv with progress remembered.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/3mro777o/qari/internal/api"
	"github.com/3mro777o/qari/internal/desktop"
	"github.com/3mro777o/qari/internal/download"
	"github.com/3mro777o/qari/internal/player"
	"github.com/3mro777o/qari/internal/popular"
	"github.com/3mro777o/qari/internal/store"
	"github.com/3mro777o/qari/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

var version = "dev"

var (
	lang        = flag.String("lang", envOr("QARI_LANG", "eng"), "language for reciter and surah names (see --languages)")
	proxy       = flag.String("proxy", os.Getenv("QARI_PROXY"), "proxy URL for all traffic, e.g. socks5://127.0.0.1:1080")
	theme       = flag.String("theme", os.Getenv("QARI_THEME"), "color theme name (see --themes; overrides saved choice)")
	listThemes  = flag.Bool("themes", false, "list available color themes and exit")
	arabic      = flag.Bool("arabic", false, "also show Arabic surah names (needs a terminal with Arabic text support)")
	noNotify    = flag.Bool("no-notify", false, "no desktop notification when a surah starts")
	noUI        = flag.Bool("no-ui", false, "with play, radio or random: play without the interface, one status line")
	languages   = flag.Bool("languages", false, "list the name languages the API offers and exit")
	showVersion = flag.Bool("version", false, "print the version and exit")
)

const usageText = `qari %s: listen to the Quran in your terminal

Usage:
  qari [flags]                      open the player
  qari play <qari> [surah] [flags]  play a surah (no surah: resume that qari)
  qari radio <station> [flags]      play a live radio station
  qari random [recipe] [flags]      random play from a recipe (default: the last one)
  qari completion bash|zsh|fish|powershell
                                    print the Tab-completion script

Names can be partial, in English or Arabic; a surah is a number or a name.
Press Tab to complete them. Flags go before or after the command.

Examples:
  qari play alafasi 18
  qari play "abdulbasit abdulsamad" yasin --no-ui
  qari radio ajmy
  qari random "short surahs"

Flags:
`

func main() {
	if len(os.Args) > 1 && os.Args[1] == "__complete" {
		complete(os.Stdout, os.Args[2:])
		return
	}
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), usageText, version)
		flag.PrintDefaults()
		fmt.Fprintln(flag.CommandLine.Output(), "\nPress ? inside the app for keys. Needs mpv for playback.")
	}
	args := parseArgs(flag.CommandLine, os.Args[1:])
	if version == "dev" { // installed with go install: use the module version
		if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
			version = bi.Main.Version
		}
	}
	if *showVersion {
		fmt.Println("qari", version)
		return
	}
	if *listThemes {
		for _, t := range ui.Themes() {
			fmt.Printf("%-11s %-18s %s\n", t.ID, t.Name, t.Desc)
		}
		return
	}
	if *theme != "" && !ui.ValidTheme(*theme) {
		fail(fmt.Errorf("unknown theme %q (run --themes to list available themes)", *theme))
	}
	cmd := ""
	if len(args) > 0 {
		cmd, args = args[0], args[1:]
	}
	switch cmd {
	case "", "play", "radio", "random":
	case "help":
		flag.CommandLine.SetOutput(os.Stdout)
		flag.Usage()
		return
	case "completion":
		if len(args) != 1 || scripts[args[0]] == "" {
			fail(fmt.Errorf("usage: qari completion bash|zsh|fish|powershell"))
		}
		fmt.Print(scripts[args[0]])
		return
	default:
		fail(fmt.Errorf("unknown command %q (run qari help)", cmd))
	}
	if *noUI && cmd == "" {
		fail(fmt.Errorf("--no-ui needs a command: qari play, radio or random"))
	}

	httpClient, err := newHTTPClient(*proxy)
	if err != nil {
		fail(err)
	}
	cacheDir, stateFile, dataDir := paths()
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
	var desk *desktop.Desktop
	if player.Available() {
		var relay *player.Relay
		if *proxy != "" || os.Getenv("HTTPS_PROXY") != "" || os.Getenv("https_proxy") != "" {
			if relay, err = player.NewRelay(httpClient); err != nil {
				fail(err)
			}
		}
		p = player.New(relay)
		desk = desktop.Start(!*noNotify)
		defer desk.Close()
	}

	opts := ui.Options{
		Client: client, Lang: *lang, Store: st, Player: p, Desktop: desk,
		Downloads: download.New(dataDir, httpClient),
		Popular:   popular.Load(), NoArabic: !*arabic, Theme: *theme, Version: version, Headless: *noUI,
	}
	if cmd != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		opts.Catalog, err = api.LoadCatalog(ctx, client, *lang)
		cancel()
		if err != nil {
			fail(fmt.Errorf("loading the catalog: %w", err))
		}
		if opts.Start, err = ui.Resolve(opts, cmd, args); err != nil {
			fail(err)
		}
	}
	notice := checkUpdate(httpClient, st)
	var prog *tea.Program
	if *noUI {
		prog = tea.NewProgram(ui.New(opts), tea.WithInput(nil))
	} else {
		prog = tea.NewProgram(ui.New(opts), tea.WithAltScreen(), tea.WithMouseCellMotion())
		go func() {
			if s := <-notice; s != "" {
				prog.Send(ui.Notice(s))
			}
		}()
	}
	// Closing the terminal (SIGHUP) or a kill (SIGTERM) must not leave mpv
	// playing on its own.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGTERM)
	go func() { <-sig; prog.Quit() }()
	final, err := prog.Run()
	if errors.Is(err, tea.ErrInterrupted) { // ctrl+c in --no-ui mode
		err = nil
	}
	if p != nil {
		p.Close()
	}
	if serr := st.Save(); serr != nil && err == nil {
		err = serr
	}
	if err != nil {
		fail(err)
	}
	if *noUI {
		fmt.Println(final.View()) // the renderer erases its line on exit
		select {
		case s := <-notice:
			if s != "" {
				fmt.Fprintln(os.Stderr, s)
			}
		default:
		}
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

// paths follows the XDG layout: the catalog is cache, progress is state,
// downloaded surahs are data.
func paths() (cacheDir, stateFile, dataDir string) {
	cache, err := os.UserCacheDir()
	if err != nil {
		cache = os.TempDir()
	}
	if runtime.GOOS == "windows" { // %LocalAppData%\qari and %AppData%\qari
		roaming, err := os.UserConfigDir()
		if err != nil {
			roaming = cache
		}
		return filepath.Join(cache, "qari", "cache"), filepath.Join(roaming, "qari", "state.json"), filepath.Join(cache, "qari", "downloads")
	}
	home, _ := os.UserHomeDir()
	xdg := func(env, def string) string {
		if v := os.Getenv(env); v != "" {
			return v
		}
		if home == "" {
			return cache
		}
		return filepath.Join(home, def)
	}
	return filepath.Join(cache, "qari"), filepath.Join(xdg("XDG_STATE_HOME", ".local/state"), "qari", "state.json"),
		filepath.Join(xdg("XDG_DATA_HOME", ".local/share"), "qari")
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
