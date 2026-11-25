# qari — build the user interface

## The product
`qari` is a terminal Quran player, to be published on GitHub for anyone to use. People open it to listen to
the Quran: pick a qari (reciter) and a surah, or a surah and then a qari, and listen; come back later and
continue where they stopped. It runs in any terminal, dark or light.

## Your job
The backend is done. The user interface does not exist: the `internal/ui` package was removed. Design and
build the whole interface from scratch — layout, navigation, visual design, themes, copy. You have full
creative freedom over how it looks and is organised. The owner wants it to be the most beautiful terminal
app he has seen, with a design that is clearly designed (not a recolored default list), plus themes.

`main.go` currently calls `ui.New(ui.Options{Client, Lang, Store, Player, Popular, NoArabic, Version})` and
runs it with `tea.NewProgram(m, tea.WithAltScreen())`. You may change that call and the options to fit your
design. Stack: Go 1.24+, Bubble Tea v1.3.10, Bubbles v1.0, Lip Gloss v1.1 (already in go.mod). No new
dependencies, no v2 of these libraries.

## Backend you get (read the code; these are the entry points)
- `internal/api`: `api.New(httpClient, cacheDir)`; `api.LoadCatalog(ctx, client, lang)` returns a
  `*Catalog`: `Reciters []*ReciterEntry` (each has `ID`, `Name` in the display language, `Arabic` name,
  `Search` folded text in all loaded languages, `Complete` (has a full 114-surah recording), `Listens`
  and `Known` (to be filled from `popular`), `Date`, and `Moshaf []Moshaf` = its recordings, each with
  `ID`, `Name` like "Rewayat Hafs A'n Assem - Murattal", `RewayaID`, `MoshafType` (riwayah*10+style:
  1 Murattal, 2 Mujawwad, 3 teaching), `SurahTotal`, `Surahs()` set, `URL(n)` mp3 address);
  `Surahs []SurahEntry` (index n-1: `Name`, `Arabic`, `Search`, `Makkia`, `StartPage`, `EndPage`);
  `Riwayat map[int]string`; `Radios []Radio` (`Name`, `URL`); `Lang` (display language actually used);
  `Reciter(id)`, `SurahName(n)`. `client.Offline()` is true when data came from an expired cache.
  Loading can take several seconds on first run (it is then cached for a day).
- `internal/player`: `player.Available()` (is mpv installed); `player.New(relay)`; `Events chan Event`
  (`Pos`, `Duration`, `Paused`, `Loading`, `Ended`, `Err`; position updates about once per second);
  `Play(url, startSeconds, title)`, `TogglePause`, `Seek(rel)`, `SeekTo(abs)`, `Stop`, `AddVolume(d)`
  (returns new volume), `SetMute`, `SetSpeed` (pitch preserved), `Close`.
- `internal/store`: persistent state. `Favorites map[int]bool` (qari ids), `Progress` per `Track{Reciter,
  Moshaf, Surah}` with `Position`, `Duration`, `Finished`, `Updated`, `Fraction()`; `Plays map[int]int`;
  `Last *Track`; `Settings` (add fields you need, e.g. theme). `Update(track,pos,dur)`, `MarkFinished`,
  `ResumeAt(track)` (where to restart: saved position minus 3 s, or 0 if finished), `Recent(n)`
  (unfinished tracks, newest first), `Save()`.
- `internal/popular`: `Load()` → `map[qariID]Entry{Listens, Known}` from real listen counts.
- `internal/norm`: `Fold(s)` (lowercase, strip diacritics, unify Arabic letter variants) and
  `Match(foldedHaystack, query)` for search.
- `internal/keymap`: `LatinKey(r)` maps a character typed on the Arabic keyboard layout to the US key in
  the same position. Apply it to single-key shortcuts (not to text typed into a search box) so every
  shortcut works without switching layout.

## Features the interface must offer
1. Browse qaris; open a qari to choose one of their recordings (riwayah and style), then a surah.
2. Browse the 114 surahs; open a surah to see the qaris who recorded it and play directly.
3. Live Quran radio stations.
4. Search every list, in English or Arabic or the display language (use `Search` + `norm.Match`). Qari
   Arabic names are searchable but must not be displayed; surah Arabic names may be displayed.
5. By default show only well-known qaris (`Known` or `Listens>0`) with a complete Quran; one key toggles
   showing everyone, and its on-screen label must say what pressing it will do now. Favorites always show.
6. Sort qaris by popularity (Listens), name, or recently updated; filter by riwayah. When opening a qari
   or playing for a surah, prefer the chosen riwayah, else Hafs, complete recordings, Murattal style.
7. Favorites.
8. Playback: pause/resume, seek ±10 s and ±1 min, next/previous surah within the recording, volume, mute,
   speed (0.75–2×), stop, after-a-surah mode (autoplay next / repeat / stop), sleep timer (15–60 min or
   end of surah), jump to what is playing.
9. Remember the position of every surah (update the store on player events, save periodically and on
   quit), show progress and finished surahs, and offer "continue listening" prominently.
10. Themes: at least 6, each with a name and its own identity; required among them: one that uses only
    the terminal's 16 ANSI colors (follows the user's terminal theme), one for light backgrounds, one high
    contrast. Switch live with `c`, `--theme NAME` flag / `QARI_THEME`, `--themes` lists them, choice saved.
11. A help screen listing every key. Brief status messages (about 1.5 s) for actions.
12. Clear states for loading, offline (cached data), mpv not installed (browsing still works), network
    error (suggest `--proxy`), and empty search results.
13. `--no-arabic` hides Arabic text for terminals without right-to-left support.

Use well-known keys where they exist: `space` pause, `n`/`p` next/previous, `+`/`-` volume, `m` mute,
`[`/`]` speed (mpv), arrows seek, `/` search, `?` help, `q` quit, `j`/`k`, `g`/`G`, `tab`/number keys.

## Hard technical constraints (learned the hard way)
- Every frame must be exactly terminal width × height at any size from 60×16 to 200×60. A line wider
  than the terminal wraps and pushes the top off-screen. Write a test that renders every screen and theme
  at several sizes from a catalog and asserts this (the API cache JSON can be produced by running once).
- Glyph widths: some symbols render double-width in Konsole/tmux and break alignment: never use ☾ ◑ ◕ ⏸,
  emoji, or Arabic ornaments such as ۝. Box drawing, block elements and ✦ ◔ ✓ ♪ ★ ‖ ▶ ● › are safe.
- Wrap displayed Arabic text in U+2067 … U+2069 isolates so bidi terminals keep the columns in place.
- Don't paint large background areas: many users have translucent terminals over a wallpaper.
- Interface language is English. The reciters are called "qaris" in the interface.
- Run `gofmt -w .`, `go vet ./...`, `go test ./...` (Go: `export PATH="$(mise where go@latest)/bin:$PATH"`).
- README: update the usage/keys/themes sections to match what you build; leave the two screenshot image
  lines in place (the owner will retake them).

## Rules for this run
- Work only inside the repository directory you are given. Do not commit, push, install anything, or touch
  `~/.local`, `~/.cache`, `~/.config`, other tmux sessions, windows, `hyprctl`, or `warp-cli`.
- To try it: the network blocks mp3quran.net, so run with `--proxy socks5://127.0.0.1:40000` and isolated
  data: `XDG_STATE_HOME=$PWD/.design/state XDG_CACHE_HOME=$PWD/.design/cache`, inside a tmux session you
  create (`tmux new-session -d -s gem -x 130 -y 34 '...'`; read it with `tmux capture-pane -p -e`; try
  64×18 too). Kill only your own session at the end. Audio plays on the owner's speakers: pause (`space`)
  or stop right after checking playback.

## Report
Your design concept, the screens and how they work, the theme list (name, idea, key colors), key map,
anything you could not verify.
