# qari

**Listen to the Quran from your terminal.** Browse 240+ qaris (reciters) and all 114 surahs from
[mp3quran.net](https://mp3quran.net), search in English or Arabic, and pick up exactly where
you left off. Playback runs through [mpv](https://mpv.io).

![Home screen: continue listening, favorite qaris and the most listened qaris](docs/home.png)

![A qari's surah list while Ya-Sin is playing; the player frame turns gold and carries the surah name](docs/playing.png)

## Features

- **Browse by qari or by surah.** Open a qari to see their recitations and surahs, or open
  a surah to see every qari who recorded it.
- **Search in any script.** `afasy`, `العفاسي` and `العَفاسي` all find the same qari;
  `kahf`, `18` and `الكهف` all find the same surah. Diacritics and letter variants (أ/إ/ا, ة/ه, ى/ي)
  are ignored.
- **Sensible default list.** By default you see well-known qaris with a complete recording
  of the Quran. Press `a` to show everyone, including partial and lesser-known recordings.
- **Most listened first.** Qaris are ranked by real listening numbers (see
  [Popularity data](#popularity-data)). Sort by name or by recently updated with `s`.
- **Remembers your place.** Every surah keeps its position. *Continue listening* on the home
  screen resumes it, and finished surahs get a ✓.
- **Made for listening:** autoplay the next surah, repeat one surah, sleep timer (15–60 minutes
  or end of surah), speed 0.75×–2× without changing pitch, volume and mute, and riwayah filter (Hafs,
  Warsh, Qalun, …).
- **Favorites** and **live Quran radio** stations.
- **Seven built-in themes** including dark manuscript palettes, a light-background theme, a 16-color
  terminal-native theme, and a high-contrast theme — switch live with `c`, with zero full-screen
  background fill so translucent terminals stay transparent.
- **Media keys** work if you have [mpv-mpris](https://github.com/hoyon/mpv-mpris) installed.
- **Works offline** for browsing: the catalog is cached for a day and used when you are offline.
- **Name language:** qari and surah names in any of the 21 languages the API offers
  (`--lang fr`, `--lang ur`, …). The interface itself is English.

## Install

You need **mpv**: `sudo pacman -S mpv` · `sudo apt install mpv` · `sudo dnf install mpv` ·
`brew install mpv`.

**Prebuilt binary:** download the archive for your system from
[Releases](https://github.com/3mro777o/qari/releases), extract it and put `qari` in your `PATH`.

**With Go 1.24+:**

```sh
go install github.com/3mro777o/qari@latest
```

**From source:**

```sh
git clone https://github.com/3mro777o/qari && cd qari
go build -o qari . && ./qari
```

Linux and macOS are supported. Windows is not yet (mpv's IPC uses named pipes there).

## Usage

Run `qari`. Press `?` at any time for the full list of keys. The keys follow other terminal players
(spotify_player, termusic) and mpv: `space` pauses, `n`/`p` skip surahs, `[`/`]` change speed. Shortcuts also work on the Arabic
keyboard layout: they follow the key position, so `ض` quits like `q` and `ظ` searches like `/`.

| Key | Action |
| --- | --- |
| `1`–`4`, `tab` / `shift+tab` | Home, Qaris, Surahs, Radio |
| `↑` `↓` / `j` `k` | Move (`pgup`/`pgdn` or `ctrl+u`/`ctrl+d` to page) |
| `g` `G` | Jump to top / bottom of list |
| `enter` | Open, or play |
| `/` | Search the current list (English, Arabic, or surah number) |
| `esc` | Close popup, clear the search, then go back |
| `space` | Pause / resume (or resume last surah when idle) |
| `←` `→` | Seek ±10 seconds |
| `shift+←` `shift+→` (or `<` `>`) | Seek ±1 minute |
| `n` `p` | Next / previous surah in the current recording |
| `f` | Add or remove a qari from favorites |
| `a` | Toggle well-known complete qaris vs all qaris |
| `s` | Sort qaris: popular, A–Z, recently updated |
| `w` | Filter by riwayah |
| `e` / `v` | Choose or cycle recording inside a qari's page |
| `i` | Open a qari's full page from a surah's qari list |
| `r` | After a surah: autoplay next, repeat, or stop |
| `t` | Sleep timer (15–60 min or end of surah) |
| `-` `+` | Volume |
| `m` | Mute |
| `[` `]` | Slower / faster (0.75×–2.0×, like mpv) |
| `x` | Stop playback |
| `o` | Jump to what is playing |
| `c` | Cycle color theme live (saved automatically) |
| `A` | Show or hide Arabic surah names |
| `?` | Show or hide the help screen |
| `q` / `ctrl+c` | Quit (your place is saved) |

### Options

```text
--lang CODE    language for qari and surah names (default eng; see --languages)
--languages    list the languages the API offers
--theme NAME   color theme name (see --themes; overrides saved choice)
--themes       list available color themes
--proxy URL    send all traffic through a proxy, e.g. socks5://127.0.0.1:1080
--no-arabic    hide the Arabic surah names
--version      print the version
```

`QARI_LANG`, `QARI_THEME` and `QARI_PROXY` set the same defaults from the environment.

### Themes

Press `c` inside the app to switch themes live (your choice is remembered), or pass `--theme NAME` / `QARI_THEME`. None of the themes paint large background blocks, so translucent terminals remain transparent.

| Theme | Identity |
| --- | --- |
| `mihrab` | **Mihrab** (default) — warm gilded gold, alabaster & sage on dark terminals |
| `andalus` | **Al-Andalus** — Alhambra terracotta, glazed turquoise & courtyard ivory |
| `lapis` | **Samarkand Lapis** — Timurid lapis azure, antique gold leaf & moonlit silk |
| `madinah` | **Madinah** — luminous emerald, soft brass gold & warm pearl |
| `parchment` | **Parchment** — light-background manuscript palette: walnut ink, ochre & forest green |
| `ansi` | **Terminal** — uses only the terminal's 16 ANSI colors (follows your terminal theme) |
| `contrast` | **High Contrast** — maximum contrast white, vivid gold, electric cyan & crisp borders |

### Arabic text

Qari names are searchable in Arabic but shown only in the chosen language. Arabic surah
names show up correctly in terminals that support right-to-left text, such as Konsole,
GNOME Terminal and other VTE terminals, WezTerm and mlterm. In terminals without it, Arabic looks
reversed. Use `--no-arabic` or press `A` there. Searching in Arabic works in every terminal.

### Blocked networks

Some school and office networks block mp3quran.net. If the catalog does not load, use a
proxy you are allowed to use:

```sh
qari --proxy socks5://127.0.0.1:1080
```

mpv cannot use SOCKS proxies, so qari forwards the audio to mpv through a small local relay.
Seeking still works.

## Popularity data

mp3quran.net publishes no play counts. Popularity comes from a snapshot in
[`internal/popular/popular.json`](internal/popular/popular.json), built by
[`tools/genpopular`](tools/genpopular/main.go) from:

- **listen counts** published by [way2quran.com](https://www.way2quran.com/en/reciters?sortBy=-totalListens)
- **"well-known"**: the reciter is also on [quranicaudio.com](https://quranicaudio.com) or
  [quran.com](https://quran.com), or way2quran marks them as a top reciter

Reciters are matched across sites by their Arabic and English names. To refresh the snapshot:

```sh
go run ./tools/genpopular
```

## Where data is kept

| What | Where |
| --- | --- |
| Favorites, progress, settings | `$XDG_STATE_HOME/qari/state.json` (default `~/.local/state/qari/`) |
| Catalog cache | `$XDG_CACHE_HOME/qari/` (default `~/.cache/qari/`, `~/Library/Caches/qari/` on macOS) |

## Credits

Recitations, names and the API: [mp3quran.net](https://mp3quran.net). Popularity data:
way2quran.com, quranicaudio.com and quran.com. Built with
[Bubble Tea](https://github.com/charmbracelet/bubbletea) and [mpv](https://mpv.io).

## License

[MIT](LICENSE)
