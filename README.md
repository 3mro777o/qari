# qari

**Listen to the Quran from your terminal.** Browse 240+ qaris (reciters) and all 114 surahs from
[mp3quran.net](https://mp3quran.net), search in English or Arabic, and pick up exactly where
you left off. Playback runs through [mpv](https://mpv.io).

![Home: continue listening, favorite qaris and the most listened qaris](docs/home.png)

![A qari's page while Al-Kahf plays; the player frame carries the surah name](docs/playing.png)

## Features

- **Browse by qari or by surah.** Open a qari to see their recordings and surahs, or open
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
- **Random play with recipes.** Press `z` and pick a recipe: a saved filter of qaris (all, well-known,
  favorites or your own list), riwayat, styles (Murattal, Mujawwad, ...) and surahs (all, Juz Amma,
  your own list), or radio stations. When a surah ends another random one starts, no surah by the
  same qari repeats until the recipe runs out, `n` picks another and `p` goes back. Built-in recipes:
  *Anything* (well-known qaris, Hafs, Murattal), *Favorite qaris*, *Short surahs* (Juz Amma),
  *Daily surahs* (Al-Kahf, As-Sajdah, Ya-Sin, Al-Waqi'ah, Al-Mulk), *Mujawwad*, *Random radio* and
  *Favorite radios*. Create your own with `a`, or copy and change a built-in one with `c`.
  In the editor `enter` changes a field, `/` searches the qaris or surahs, and `s` saves.
- **History.** The History tab lists the last 200 surahs and stations you started, by day, with
  where you left off. `enter` plays one again from that point.
- **Offline listening.** Press `d` to download a surah, or a whole recording. The Downloads
  tab shows the queue with progress and speed; downloaded surahs play from disk, with no network.
- **Made for listening:** autoplay the next surah, repeat one surah, sleep timer (15–60 minutes
  or end of surah), speed 0.75×–2× without changing pitch, volume and mute, and riwayah filter (Hafs,
  Warsh, Qalun, …).
- **Media keys and notifications** (Linux): qari is an MPRIS player, so keyboard media keys, bar
  widgets and `playerctl` control it, and a notification shows each new surah.
- **Mouse:** click tabs and rows, click a row again to open or play it, scroll with the wheel,
  click or drag the progress bar to seek.
- **Five themes** with a live preview picker (`c`). No theme paints a background, so translucent
  terminals stay transparent.
- **Favorites** and **live Quran radio** stations.
- **Name language:** qari and surah names in any of the 21 languages the API offers
  (`--lang fr`, `--lang ur`, …). The interface itself is English.

![Downloads: the queue with progress and speed, and recordings saved for offline listening](docs/downloads.png)

## Install

**macOS (Homebrew)**, installs mpv too:

```sh
brew tap 3mro777o/qari https://github.com/3mro777o/qari
brew install qari
```

**Windows (Scoop)**, then mpv from the extras bucket:

```powershell
scoop bucket add qari https://github.com/3mro777o/qari
scoop install qari
scoop bucket add extras; scoop install mpv
```

Elsewhere you need **mpv** first: `sudo pacman -S mpv` · `sudo apt install mpv` · `sudo dnf install mpv`
· `winget install mpv`.

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

Linux, macOS and Windows are supported. Use Windows Terminal on Windows.

## Usage

Run `qari`. Press `?` at any time for every key. The keys follow mpv, cmus and vim:
`space` pauses, `n`/`p` skip surahs, `[`/`]` change speed, `j`/`k` move. Shortcuts also work on the
Arabic keyboard layout: they follow the key position, so `ض` quits like `q` and `ظ` searches like `/`.

| Key | Action |
| --- | --- |
| `1`–`6`, `tab` / `shift+tab` | Home, Qaris, Surahs, Radio, Downloads, History |
| `j` `k` / `↓` `↑` | Move |
| `ctrl+d` `ctrl+u` / `pgdn` `pgup` | Half a page down / up |
| `g` `G` / `home` `end` | First / last row |
| `enter` / `l` | Open, or play |
| `esc` / `h` | Back (clears the search first) |
| `/` | Search the current list (English, Arabic, or surah number) |
| `o` | Go to what is playing |
| `space` | Play / pause (resumes your last surah when idle) |
| `←` `→` | Seek 10 seconds |
| `,` `.` (or `shift+←` `shift+→`) | Seek 1 minute |
| `n` `p` | Next / previous surah in the recording (in random play: another pick / back) |
| `z` | Random: `enter` plays a recipe; `a` new, `e` edit, `c` copy, `x` delete (press twice) |
| `+` `-` | Volume |
| `m` | Mute |
| `[` `]` | Slower / faster (0.75×–2×) |
| `r` | After a surah: autoplay next, repeat, or stop |
| `t` | Sleep timer (15–60 min or end of surah) |
| `f` | Favorite qari |
| `a` | Well-known qaris / all qaris |
| `s` | Sort qaris: popular, A–Z, recently updated |
| `w` | Filter by riwayah |
| `v` | Choose a qari's recording |
| `d` | Download the surah, or the whole recording (on a qari or in the recording list; asks to press again) |
| `x` | Delete a download or cancel it; remove a History entry |
| `c` | Color theme (moving previews it, `enter` keeps it, `esc` goes back) |
| `?` | Help |
| `q` / `ctrl+c` | Quit (your place is saved) |

![The help screen lists every key](docs/help.png)

### Options

```text
--lang CODE    language for qari and surah names (default eng; see --languages)
--languages    list the languages the API offers
--theme NAME   color theme (see --themes; overrides the saved choice)
--themes       list the themes
--proxy URL    send all traffic through a proxy, e.g. socks5://127.0.0.1:1080
--arabic       also show the Arabic surah names
--no-notify    no desktop notification when a surah starts
--version      print the version
```

`QARI_LANG`, `QARI_THEME` and `QARI_PROXY` set the same defaults from the environment.

### Themes

Press `c` to pick a theme: the whole screen previews it as you move, `enter` keeps it (it is
remembered) and `esc` goes back.

| Theme | Identity |
| --- | --- |
| `mihrab` | **Mihrab** (default): warm gilded gold, alabaster and sage |
| `andalus` | **Al-Andalus**: Alhambra terracotta, glazed turquoise and ivory |
| `lapis` | **Samarkand Lapis**: Timurid lapis azure, gold leaf and moonlit silk |
| `madinah` | **Madinah**: luminous emerald, soft brass and warm pearl |
| `contrast` | **High Contrast**: white, vivid gold and electric cyan |

![The theme picker previewing Al-Andalus](docs/theme.png)

### Media keys

On Linux qari registers as `org.mpris.MediaPlayer2.qari`, so media keys, desktop widgets and
`playerctl -p qari play-pause` work while it runs. It starts mpv without scripts, so an installed
mpv-mpris plugin does not show a second player. On macOS and Windows there are no media keys yet;
qari shows a notification when a new surah starts.

### Arabic text

Names are shown in the chosen language (English by default), and you can search in Arabic:
typing `البقره` finds Al-Baqarah. `--arabic` also shows the Arabic surah names, for terminals
that support right-to-left text (Konsole, GNOME Terminal and other VTE terminals, WezTerm,
mlterm); elsewhere Arabic looks reversed.

### Blocked networks

Some school and office networks block mp3quran.net. If the catalog does not load, use a
proxy you are allowed to use:

```sh
qari --proxy socks5://127.0.0.1:1080
```

mpv cannot use SOCKS proxies, so qari forwards the audio to mpv through a small local relay.
Seeking still works. Downloads go through the same proxy.

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
| Favorites, progress, recipes, history, settings | `$XDG_STATE_HOME/qari/state.json` (default `~/.local/state/qari/`) |
| Catalog cache | `$XDG_CACHE_HOME/qari/` (default `~/.cache/qari/`, `~/Library/Caches/qari/` on macOS) |
| Downloaded surahs | `$XDG_DATA_HOME/qari/` (default `~/.local/share/qari/`) |

On Windows: state in `%AppData%\qari\state.json`, cache and downloads in `%LocalAppData%\qari\`.

## Credits

Recitations, names and the API: [mp3quran.net](https://mp3quran.net). Popularity data:
way2quran.com, quranicaudio.com and quran.com. Built with
[Bubble Tea](https://github.com/charmbracelet/bubbletea) and [mpv](https://mpv.io).

## License

[MIT](LICENSE)
