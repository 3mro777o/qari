<div align="center">

<img src="docs/banner.jpg" alt="qari: listen to the Quran from your terminal" width="100%">

<p>
<a href="https://github.com/3mro777o/qari/releases/latest"><img src="https://img.shields.io/github/v/release/3mro777o/qari?style=flat-square&color=DFAA4B&label=release" alt="Latest release"></a>
<a href="https://github.com/3mro777o/qari/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/3mro777o/qari/ci.yml?style=flat-square&label=tests" alt="Tests"></a>
<a href="https://github.com/3mro777o/qari/actions/workflows/install.yml"><img src="https://img.shields.io/github/actions/workflow/status/3mro777o/qari/install.yml?style=flat-square&label=install%20check" alt="Install check"></a>
<img src="https://img.shields.io/badge/Linux%20%C2%B7%20macOS%20%C2%B7%20Windows-6BB892?style=flat-square" alt="Linux, macOS, Windows">
<a href="LICENSE"><img src="https://img.shields.io/github/license/3mro777o/qari?style=flat-square&color=AEA596" alt="MIT license"></a>
</p>

**English** · [العربية](README.ar.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [اردو](README.ur.md) · [Français](README.fr.md) · [Deutsch](README.de.md)

</div>

**qari** is a Quran player that lives in your terminal. It lists 240+ qaris (reciters) and all
114 surahs from [mp3quran.net](https://mp3quran.net), plays them through [mpv](https://mpv.io),
remembers where you stopped in every surah, repeats ayahs for memorizing, and keeps surahs on
disk for offline listening. No account, no ads, no tracking.

<div align="center">
<img src="docs/demo.gif" alt="A short tour: search a qari, play a surah, repeat ayahs, change the theme" width="90%">
</div>

## Contents

- [Highlights](#highlights)
- [Install](#install) · [Update](#update) · [Uninstall](#uninstall)
- [Your first five minutes](#your-first-five-minutes)
- [Guides](#guides): [Memorize with ayah repeat](#memorize-with-ayah-repeat) ·
  [Random play and recipes](#random-play-and-recipes) · [Offline listening](#offline-listening) ·
  [Commands and Tab completion](#commands-and-tab-completion) · [Radio, history, favorites](#radio-history-and-favorites) ·
  [Themes](#themes) · [Media keys](#media-keys) · [Arabic text](#arabic-text) · [Blocked networks](#blocked-networks)
- [All keys](#all-keys) · [Options](#options)
- [Troubleshooting](#troubleshooting)
- [Where data is kept](#where-data-is-kept) · [Privacy](#privacy) · [Popularity data](#popularity-data)
- [Build from source](#build-from-source) · [Credits](#credits) · [License](#license)

## Highlights

| | |
| --- | --- |
| 🎙️ **240+ qaris, 114 surahs** | Browse by qari or by surah. Well-known qaris with a full recording come first, ranked by real listening numbers. |
| 🔎 **Search in any script** | `alafasi`, `العفاسي` and `العَفاسي` find the same qari; `kahf`, `18` and `الكهف` find the same surah. |
| 📖 **Memorize (hifz)** | Repeat a range of ayahs: each ayah N times, the whole range M times, with a pause between. |
| ⏯️ **Remembers your place** | Every surah keeps its position. *Continue listening* resumes it; finished surahs get a ✓. |
| 🎲 **Random play** | Recipes such as *Short surahs* or *Daily surahs* (Al-Kahf, Ya-Sin, Al-Mulk…), or your own. |
| 💾 **Offline** | Download a surah or a whole recording; it then plays from disk with no network. |
| ⌨️ **Commands with Tab** | `qari play alafasi kahf` from the shell. Tab completes qari names, surahs and stations. |
| 📻 **Live radio** | The mp3quran.net Quran radio stations. |
| 🎨 **Five themes** | Inspired by Islamic art: Mihrab, Al-Andalus, Samarkand Lapis, Madinah, High Contrast. |
| 🖱️ **Keys and mouse** | vim/mpv-style keys that also work on the Arabic layout, plus clicks, scroll and seek by dragging. |

## Install

One command installs qari **and** mpv (the audio engine qari uses). Copy it into a terminal and
press Enter.

**Linux and macOS**

```sh
curl -fsSL https://3mro777o.github.io/qari/install.sh | sh
```

**Windows** (open *PowerShell* from the Start menu; no admin needed)

```powershell
irm https://3mro777o.github.io/qari/install.ps1 | iex
```

Then type `qari` and press Enter. If the command is not found, open a new terminal window.

<details>
<summary><b>What exactly does the install command do?</b></summary>

You should know what a command does before you pipe it into a shell. Both scripts are short; read
them here: [install.sh](install.sh), [install.ps1](install.ps1).

**Linux** ([install.sh](install.sh))
1. Installs mpv with your package manager (`pacman`, `apt`, `dnf` or `zypper`) if it is missing.
   This step asks for your password. If it fails, qari still installs and tells you to add mpv yourself.
2. Downloads the latest qari release for your CPU (x86-64 or ARM64) from GitHub Releases.
3. Checks the download against the release's `checksums.txt` (SHA-256) and stops if it does not match.
4. Puts the `qari` program in `/usr/local/bin` (asks for your password), which is on the PATH everywhere.
5. Sets up Tab completion for the shells you have: bash, zsh and fish.

**macOS** ([install.sh](install.sh)): uses [Homebrew](https://brew.sh) (install it first). The
script adds this repository as a tap, then runs `brew install qari`, which also installs mpv and Tab
completion. Homebrew then keeps both updated with `brew upgrade`.

**Windows** ([install.ps1](install.ps1))
1. Downloads the latest qari release from GitHub Releases and checks its SHA-256 checksum.
2. Puts `qari.exe` in `%LOCALAPPDATA%\Programs\qari` (your user folder, so no admin rights).
3. If mpv is missing, downloads the portable mpv build by
   [shinchiro](https://github.com/shinchiro/mpv-winbuild-cmake) to `%LOCALAPPDATA%\Programs\mpv`.
4. Adds both folders to your user PATH, so `qari` works in any new terminal.
5. Adds one line to your PowerShell profile that turns on Tab completion.

</details>

> [!NOTE]
> **Windows Smart App Control** blocks programs that are not code-signed, and qari is not signed
> yet. If Windows refuses to run `qari.exe`, that is the reason. See [Troubleshooting](#troubleshooting).

<details>
<summary><b>Other ways to install</b> (Homebrew, Scoop, binary, Go)</summary>

**Homebrew** (macOS and Linux):

```sh
brew tap 3mro777o/qari https://github.com/3mro777o/qari
brew trust 3mro777o/qari   # newer Homebrew versions ask for this
brew install qari
```

**Scoop** (Windows):

```powershell
scoop bucket add qari https://github.com/3mro777o/qari
scoop bucket add extras
scoop install qari mpv
```

**Prebuilt binary:** download the archive for your system from
[Releases](https://github.com/3mro777o/qari/releases), extract it and put `qari` somewhere on your
`PATH`. Install **mpv** yourself (`sudo apt install mpv`, `brew install mpv`, …).

**With Go 1.24+** (also needs mpv):

```sh
go install github.com/3mro777o/qari@latest
```

</details>

**Terminal:** any modern terminal works. On Windows use *Windows Terminal* (the default on
Windows 11). For Arabic names, see [Arabic text](#arabic-text).

### Update

Run the install command again. It always fetches the newest release and replaces the old one;
your favorites, progress and downloads stay. Homebrew users can also run `brew upgrade qari`.

qari checks GitHub once a day and shows *"qari X is out"* in its footer when a new version is ready.

### Uninstall

Your data is separate from the program; delete it too if you want a clean removal
(see [Where data is kept](#where-data-is-kept)).

<details>
<summary><b>Linux, macOS, Windows</b></summary>

**Linux**

```sh
sudo rm /usr/local/bin/qari
rm -f ~/.local/share/bash-completion/completions/qari ~/.config/fish/completions/qari.fish
sudo rm -f /usr/local/share/zsh/site-functions/_qari
rm -rf ~/.local/state/qari ~/.cache/qari ~/.local/share/qari   # your data, optional
```

If zsh completion was added to `~/.zshrc`, delete the line that ends with `# qari Tab completion`.

**macOS**

```sh
brew uninstall qari && brew untap 3mro777o/qari
rm -rf ~/.local/state/qari ~/Library/Caches/qari ~/.local/share/qari   # your data, optional
```

**Windows** (PowerShell)

```powershell
Remove-Item -Recurse "$env:LOCALAPPDATA\Programs\qari"
Remove-Item -Recurse "$env:LOCALAPPDATA\Programs\mpv"            # only if the installer added it
Remove-Item -Recurse "$env:APPDATA\qari", "$env:LOCALAPPDATA\qari" # your data, optional
```

Then remove the two folders from your user PATH (*Settings → System → About → Advanced system
settings → Environment Variables*) and the `qari completion` line from your profile (`notepad $PROFILE`).

</details>

## Your first five minutes

1. **Open it.** Type `qari`. You land on **Home**: *Continue listening*, your favorite qaris and
   the most listened qaris. The tabs on top are `1` Home, `2` Qaris, `3` Surahs, `4` Radio,
   `5` Downloads, `6` History. Press a number or `tab` to switch.

   ![Home: continue listening, favorite qaris and the most listened qaris](docs/home.png)

2. **Find a qari.** Press `2`, then `/` and type part of a name, e.g. `alafasi` or `العفاسي`.
   Move with `j`/`k` or the arrow keys, open with `enter`.
3. **Play a surah.** The qari's page lists their surahs. Pick one and press `enter`. The player at
   the bottom shows the surah, the time and a progress bar. `space` pauses, `←`/`→` seek 10 seconds,
   `n`/`p` go to the next or previous surah.

   ![A qari's page while Al-Kahf plays; the player frame carries the surah name](docs/playing.png)

4. **Make it yours.** `f` marks a favorite qari, `[`/`]` change the speed, `t` sets a sleep timer,
   `c` changes the colors.
5. **Quit with `q`.** Your place is saved. Next time, *Continue listening* or `space` resumes it.

Lost? Press `?` for every key on the screen you are on.

## Guides

### Memorize with ayah repeat

Repeating a few ayahs over and over is the classic way to memorize (hifz). qari knows where each
ayah starts and ends for about 115 recordings, so it can loop exactly the ayahs you want.

1. Play a surah (for example Al-Kahf by Mishary Alafasi).
2. Press `e`. A window opens:

   | Field | Meaning |
   | --- | --- |
   | **From ayah / To ayah** | The range to repeat. |
   | **Each ayah** | How many times each ayah plays before moving on (1–99). |
   | **Whole range** | How many times the whole range runs (`0` = until you stop it). |
   | **Gap** | Seconds of silence before each repeat, time to recite it back yourself. |

   Move between fields with `j`/`k`, change a value with `←`/`→` or type the number, then press
   `enter` to start.

   ![The ayah repeat window](docs/repeat.png)

3. The player shows where you are, e.g. `ayah 3/10 · 2 of 3`. Press `esc` to stop.

**Quick repeat while listening:** press `i` when the first ayah you want starts, and `i` again
during the last one. qari loops that range. A third `i` stops. `u` changes how many times each
ayah plays (1, 2, 3, 5, 7, 10).

Your settings are remembered, and repeat works on downloaded surahs without a network. If a
recording has no ayah timings, qari tells you, and points to another recording of the same qari
that has them (switch with `v`).

### Random play and recipes

Press `z` to pick a **recipe**: a saved description of what to play at random. When a surah ends
another random one starts; no surah by the same qari repeats until the recipe runs out. `n` picks
another one, `p` goes back.

Built-in recipes:

| Recipe | Plays |
| --- | --- |
| *Anything* | Well-known qaris, Hafs, Murattal |
| *Favorite qaris* | Only your favorites |
| *Short surahs* | Juz Amma |
| *Daily surahs* | Al-Kahf, As-Sajdah, Ya-Sin, Al-Waqi'ah, Al-Mulk |
| *Mujawwad* | Mujawwad recitations |
| *Random radio* / *Favorite radios* | Live radio stations |

Make your own with `a`, or copy a built-in one with `c` and change it. A recipe filters by qaris
(all, well-known, favorites or a list you pick), riwayah, style (Murattal, Mujawwad, …) and surahs
(all, Juz Amma or your list). In the editor, `enter` changes a field, `/` searches the lists, and
`s` saves.

### Offline listening

Press `d` on a surah to download it, or on a recording to download all of it (press `d` twice to
confirm). The **Downloads** tab (`5`) shows the queue with progress and speed. A downloaded surah
plays from disk, so it works on a plane, and saves data. `x` deletes a download or cancels one.

![Downloads: the queue with progress and speed, and recordings saved for offline listening](docs/downloads.png)

### Commands and Tab completion

You can start playing straight from the shell. Names may be partial, in English or Arabic; a surah
is a number or a name. If a name fits several qaris, qari lists them.

```sh
qari play alafasi 18              # Al-Kahf by Mishary Alafasi, from where you left off
qari play alafasi                 # resume what you last heard from that qari
qari radio ajmy                   # a live radio station
qari random "short surahs"        # random play from a recipe (no name: the last one used)
qari play alafasi yasin --no-ui   # no interface: one status line, ctrl+c stops
qari help                         # all commands
```

The interface opens already playing, with every key available. With `--no-ui` qari plays in the
background of your terminal and prints one status line; the next surah follows like in the app,
and your place is saved the same way.

**Tab completion** is set up by the install command for bash, zsh, fish and PowerShell. Type
`qari play al` and press Tab: you get matching qari names; after a name, Tab offers that qari's
surahs. It also completes commands, options, radio stations, recipes and themes.

![Tab completion lists qaris and surahs](docs/cli.png)

To set it up by hand, load the output of `qari completion bash|zsh|fish|powershell` in your shell's
startup file. Completion learns the names after you open qari once.

### Radio, history and favorites

- **Radio** (`4`): live Quran radio stations from mp3quran.net. `f` marks a favorite station.
- **History** (`6`): the last 200 surahs and stations you started, grouped by day, with where you
  left off. `enter` plays one again from that point; `x` removes an entry.
- **Favorites:** `f` on a qari adds it to Home and to the *Favorite qaris* recipe.
- **Show everyone:** by default you see well-known qaris with a complete recording. `a` shows all
  of them, including partial and lesser-known recordings. `s` sorts by popularity, name or recently
  updated, `w` filters by riwayah (Hafs, Warsh, Qalun, …), `v` picks one of a qari's recordings.

### Themes

Press `c`: the whole screen previews each theme as you move, `enter` keeps it, `esc` goes back.
No theme paints a background, so a transparent terminal stays transparent.

| Theme | Look |
| --- | --- |
| `mihrab` | **Mihrab** (default): warm gilded gold, alabaster and sage |
| `andalus` | **Al-Andalus**: Alhambra terracotta, glazed turquoise and ivory |
| `lapis` | **Samarkand Lapis**: Timurid lapis azure, gold leaf and moonlit silk |
| `madinah` | **Madinah**: luminous emerald, soft brass and warm pearl |
| `contrast` | **High Contrast**: white, vivid gold and electric cyan |

![The theme picker previewing Al-Andalus](docs/theme.png)

### Media keys

On Linux qari registers as an MPRIS player (`org.mpris.MediaPlayer2.qari`), so keyboard media keys,
desktop widgets and `playerctl -p qari play-pause` control it, and a notification shows each new
surah. On macOS and Windows there are no media keys yet; you still get the notification.

### Arabic text

Names show in English by default, and you can always search in Arabic: `البقره` finds Al-Baqarah.
`--lang` switches qari and surah names to any of the 21 languages the API offers (`--lang ar`,
`--lang fr`, `--lang ur`, …; `--languages` lists them). `--arabic` adds the Arabic surah names next
to the English ones.

Arabic needs a terminal that supports right-to-left text: Konsole, GNOME Terminal and other VTE
terminals, WezTerm, mlterm, Windows Terminal. In other terminals Arabic letters may look reversed
or disconnected.

Shortcuts follow the key position, so they work on the Arabic keyboard layout too: `ض` quits like
`q`, `ظ` searches like `/`.

### Blocked networks

Some school and office networks block mp3quran.net. If the lists do not load, use a proxy you are
allowed to use:

```sh
qari --proxy socks5://127.0.0.1:1080
```

mpv cannot use SOCKS proxies, so qari passes the audio to mpv through a small local relay; seeking
still works, and downloads use the same proxy.

## All keys

Press `?` in the app for this list. The keys follow mpv, cmus and vim.

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
| `n` `p` | Next / previous surah (in random play: another pick / back) |
| `z` | Random: `enter` plays a recipe; `a` new, `e` edit, `c` copy, `x` delete (press twice) |
| `+` `-` | Volume |
| `m` | Mute |
| `[` `]` | Slower / faster (0.75×–2×, pitch stays the same) |
| `r` | After a surah: autoplay next, repeat, or stop |
| `t` | Sleep timer (15–60 min or end of surah) |
| `e` | Repeat ayahs to memorize |
| `i` | Quick repeat: first ayah, last ayah, stop |
| `u` | Times each repeated ayah plays (1–10) |
| `f` | Favorite qari or station |
| `a` | Well-known qaris / all qaris |
| `s` | Sort qaris: popular, A–Z, recently updated |
| `w` | Filter by riwayah |
| `v` | Choose a qari's recording |
| `d` | Download the surah or the whole recording (asks to press again) |
| `x` | Delete or cancel a download; remove a History entry |
| `c` | Color theme |
| `?` | Help |
| `q` / `ctrl+c` | Quit (your place is saved) |

**Mouse:** click tabs and rows, click a row again to open or play it, scroll with the wheel, click
or drag the progress bar to seek.

![The help screen lists every key](docs/help.png)

## Options

```text
qari [options]
qari play <qari> [surah] [--no-ui]
qari radio <station> [--no-ui]
qari random [recipe] [--no-ui]
qari completion bash|zsh|fish|powershell

--lang CODE    language for qari and surah names (default eng; see --languages)
--languages    list the languages the API offers
--theme NAME   color theme (see --themes; overrides the saved choice)
--themes       list the themes
--proxy URL    send all traffic through a proxy, e.g. socks5://127.0.0.1:1080
--arabic       also show the Arabic surah names
--no-notify    no desktop notification when a surah starts
--no-ui        with play, radio or random: play without the interface
--version      print the version
```

`QARI_LANG`, `QARI_THEME` and `QARI_PROXY` set the same defaults from the environment.

## Troubleshooting

<details>
<summary><b>Windows: "An Application Control policy has blocked this file" / Smart App Control</b></summary>

Smart App Control blocks every program that is not code-signed, and qari is not signed yet (free
code signing for open-source projects has been requested). Until then the only way is to turn
Smart App Control off in *Windows Security → App & browser control → Smart App Control settings*.
Most PCs have it off already; it is on only on some new Windows 11 installs.

</details>

<details>
<summary><b>"qari: command not found" right after installing</b></summary>

Open a new terminal: the old one does not see the updated PATH. On Linux check that
`/usr/local/bin` is on your PATH (`echo $PATH`).

</details>

<details>
<summary><b>"Install mpv to play audio" / no sound</b></summary>

qari plays through mpv. Install it with your package manager (`sudo apt install mpv`,
`sudo pacman -S mpv`, `brew install mpv`), or run the install command again. Check that `mpv --version`
works in a new terminal.

</details>

<details>
<summary><b>The lists do not load</b></summary>

mp3quran.net is probably blocked on your network; see [Blocked networks](#blocked-networks). qari
keeps the last loaded lists for offline use, and downloaded surahs always play.

</details>

<details>
<summary><b>Arabic looks reversed or disconnected</b></summary>

Your terminal does not support right-to-left text. Use one of the terminals listed in
[Arabic text](#arabic-text), or run without `--arabic` and `--lang ar`.

</details>

<details>
<summary><b>Tab shows nothing</b></summary>

Open qari once so it saves the lists; completion reads names from them. In bash, the
`bash-completion` package must be installed. Start a new terminal after installing.

</details>

<details>
<summary><b>macOS: "Refusing to load formula from untrusted tap"</b></summary>

Run `brew trust 3mro777o/qari`, then install again. The install command does this for you.

</details>

Something else? [Open an issue](https://github.com/3mro777o/qari/issues) with your system and
what you saw.

## Where data is kept

| What | Linux / macOS | Windows |
| --- | --- | --- |
| Favorites, progress, recipes, history, settings | `~/.local/state/qari/state.json` | `%AppData%\qari\state.json` |
| Catalog cache | `~/.cache/qari/` (`~/Library/Caches/qari/` on macOS) | `%LocalAppData%\qari\` |
| Downloaded surahs | `~/.local/share/qari/` | `%LocalAppData%\qari\` |

On Linux the `XDG_STATE_HOME`, `XDG_CACHE_HOME` and `XDG_DATA_HOME` variables move these folders.

## Privacy

qari has no account and sends no analytics. It talks to two places only: mp3quran.net (the lists,
audio and ayah timings) and GitHub (once a day, to see whether a newer release is out). Everything it
remembers stays on your computer.

## Popularity data

mp3quran.net publishes no play counts. Popularity comes from a snapshot in
[`internal/popular/popular.json`](internal/popular/popular.json), built by
[`tools/genpopular`](tools/genpopular/main.go) from:

- **listen counts** published by [way2quran.com](https://www.way2quran.com/en/reciters?sortBy=-totalListens)
- **"well-known"**: the reciter is also on [quranicaudio.com](https://quranicaudio.com) or
  [quran.com](https://quran.com), or way2quran marks them as a top reciter

Reciters are matched across sites by their Arabic and English names. Refresh the snapshot with
`go run ./tools/genpopular`.

## Build from source

Needs Go 1.24+ and mpv.

```sh
git clone https://github.com/3mro777o/qari && cd qari
go build -o qari . && ./qari
go test ./...
```

The code: `internal/api` (mp3quran.net client and cache), `internal/player` (mpv over JSON IPC),
`internal/store` (saved state), `internal/ui` (the [Bubble Tea](https://github.com/charmbracelet/bubbletea)
interface). Releases are built by [GoReleaser](https://goreleaser.com) when a `v*` tag is pushed.

## Credits

Recitations, names, radio and ayah timings: [mp3quran.net](https://mp3quran.net). Popularity data:
way2quran.com, quranicaudio.com and quran.com. Built with
[Bubble Tea](https://github.com/charmbracelet/bubbletea) and [mpv](https://mpv.io).

## License

[MIT](LICENSE). Translations of this page are kept in sync with the English one, which is the
reference if they differ.
