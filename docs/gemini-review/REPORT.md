Built the complete `internal/ui` interface for `qari` from scratch, wired theme persistence and CLI flags in [main.go](file:///tmp/claude-1000/-home-nova/b833c7fc-f492-427b-aeb0-b55ff2b3c32f/scratchpad/gemini/wt/main.go) and [store.go](file:///tmp/claude-1000/-home-nova/b833c7fc-f492-427b-aeb0-b55ff2b3c32f/scratchpad/gemini/wt/internal/store/store.go), added layout and behavior tests in [ui_test.go](file:///tmp/claude-1000/-home-nova/b833c7fc-f492-427b-aeb0-b55ff2b3c32f/scratchpad/gemini/wt/internal/ui/ui_test.go), and updated [README.md](file:///tmp/claude-1000/-home-nova/b833c7fc-f492-427b-aeb0-b55ff2b3c32f/scratchpad/gemini/wt/README.md).

---

### Design Concept — *Mushaf & Mihrab*

- **Quiet architectural framing, zero background blocks**: No panel, row, or screen paints background fills, so translucent terminals over wallpapers remain completely transparent. Hierarchy comes from foreground color temperature, typographic weight, hairline dividers (`─`, `│`), a crisp left selection rail (`▌›`), and an architectural accent underline (`━`) anchored directly beneath the active tab.
- **The Living Player Frame**: A single rounded bottom dock (`╭─ ... ─╮`) shifts color and presence with playback state — muted hairline when idle, glowing theme accent when playing a surah, calm secondary accent when paused, and vivid live coral when streaming radio. The surah title (plus isolated Arabic name on wide terminals) lives directly in the top border, a sleek playhead bar (`━●─`) fills the dock interior, and transient action toasts (`~1.6 s`) embed seamlessly into the bottom border without shifting layout.
- **Adaptive & exact `W × H` geometry**: Every frame is guaranteed to match exact terminal dimensions from `60×16` to `200×60` via ANSI-aware cell measurement and padding ([layout.go](file:///tmp/claude-1000/-home-nova/b833c7fc-f492-427b-aeb0-b55ff2b3c32f/scratchpad/gemini/wt/internal/ui/layout.go)). Wide terminals (`≥ 96` / `≥ 100` cols) expand into editorial master-detail / two-column layouts; compact heights (`< 20` rows) automatically collapse the player dock from 4 lines to 3 lines while preserving progress bars, controls, and full navigation.
- **Vertical Arabic calligraphic ribbon**: In Surah lists and Qari detail pages, English surah names occupy a fixed column so isolated Arabic surah names (`\u2067`…`\u2069`) form a straight vertical column. Qari Arabic names are strictly searchable-only and never rendered.

---

### Screens & How They Work

Implemented across [view.go](file:///tmp/claude-1000/-home-nova/b833c7fc-f492-427b-aeb0-b55ff2b3c32f/scratchpad/gemini/wt/internal/ui/view.go), [view_home.go](file:///tmp/claude-1000/-home-nova/b833c7fc-f492-427b-aeb0-b55ff2b3c32f/scratchpad/gemini/wt/internal/ui/view_home.go), [view_qaris.go](file:///tmp/claude-1000/-home-nova/b833c7fc-f492-427b-aeb0-b55ff2b3c32f/scratchpad/gemini/wt/internal/ui/view_qaris.go), [view_surahs.go](file:///tmp/claude-1000/-home-nova/b833c7fc-f492-427b-aeb0-b55ff2b3c32f/scratchpad/gemini/wt/internal/ui/view_surahs.go), [view_radio.go](file:///tmp/claude-1000/-home-nova/b833c7fc-f492-427b-aeb0-b55ff2b3c32f/scratchpad/gemini/wt/internal/ui/view_radio.go), and [view_overlays.go](file:///tmp/claude-1000/-home-nova/b833c7fc-f492-427b-aeb0-b55ff2b3c32f/scratchpad/gemini/wt/internal/ui/view_overlays.go):

1. **Loading & Network Error States**: Centered framed spinner while fetching/caching from `mp3quran.net`; on network error, a calm error card explains blocked-network recovery (`qari --proxy socks5://127.0.0.1:1080`) with `enter`/`r` to retry. Stale cached data displays a subtle `OFFLINE` header badge; missing `mpv` switches the dock to browse mode while keeping all catalog browsing and search intact.
2. **Home (`1`)**: On wide terminals, left column features **Continue Listening** (2-line progress cards with remaining time and mini progress bar, or a welcoming quick-start card on first run) and **Favorite Qaris**; right column displays **Most Listened** qaris dynamically sized to fill the terminal height with rank, recording count/riwayah, and listen counts. Collapses cleanly into a single scrollable column on narrow terminals.
3. **Qaris (`2`)**: Filter ribbon displays active scope, count, sort mode, riwayah filter, and an explicit dynamic action label (`a show all` ↔ `a well-known only`). Favorites always appear. On wide screens (`≥ 100` cols), a live **Qari Preview** pane shows completion progress and every recording (moshaf) with the preferred one marked `✦`.
4. **Qari Detail (`enter` on a qari)**: Breadcrumb masthead with favorite toggle, active recording bar with `[index/total e switch]` and recording completion bar (`✓ N/114`), and the surah list for that recording showing `▶`/`‖`/`✓`/`◔`, isolated Arabic names, inline progress bars, and revelation/page metadata. Preferred moshaf selection follows: chosen riwayah → Hafs → complete (114) → Murattal ([catalog_util.go](file:///tmp/claude-1000/-home-nova/b833c7fc-f492-427b-aeb0-b55ff2b3c32f/scratchpad/gemini/wt/internal/ui/catalog_util.go)).
5. **Surahs (`3`)**: All 114 surahs searchable by English, Arabic (with/without diacritics), display language, or surah number (exact numeric match floats to the top). Wide screens show a **Surah Preview** card with isolated Arabic title, revelation/pages, listening status, and top available qaris.
6. **Surah Detail (`enter` on a surah)**: Lists qaris who recorded that surah (with their best matching recording), supporting the same `a`/`s`/`w`/`f` controls. Pressing `enter` plays directly; `i` opens that qari's full page.
7. **Radio (`4`)**: Searchable list of live Quran radio stations with `● LIVE` state, dedicated live-stream dock styling, and a station preview card on wide screens; radio playback does not pollute surah progress.
8. **Overlays**:
   - **Help (`?`)**: Responsive 2-column floating shortcut card on wide screens (`≥ 100×24`, no scrolling needed) and scrollable single-column card on compact terminals.
   - **Riwayah Picker (`w`)**: Scrollable modal listing `All riwayat` and every riwayah in the catalog sorted by qari count.
   - **Recording Picker (`e` in Qari Detail)**: Lists all recordings (riwayah, style, surah count) for the active qari; `v` cycles recordings directly.

---

### Themes (7 Built-In)

Defined in [theme.go](file:///tmp/claude-1000/-home-nova/b833c7fc-f492-427b-aeb0-b55ff2b3c32f/scratchpad/gemini/wt/internal/ui/theme.go). Switch live with `c` (saved in `state.json`), override with `--theme NAME` / `QARI_THEME`, or list via `--themes`:

| ID | Name | Identity | Key Colors |
| --- | --- | --- | --- |
| `mihrab` | **Mihrab** *(default)* | Warm gilded manuscript sanctuary on dark terminals | Accent `#DFAA4B`, Secondary/Success `#6BB892`, Text `#EDE6D6`, Border `#524A3E`, Live `#E06C5A` |
| `andalus` | **Al-Andalus** | Alhambra terracotta & glazed Nasrid turquoise | Accent `#E08A5A`, Secondary `#4ABBB2`, Success `#6EC490`, Text `#EFECE4`, Border `#404E56` |
| `lapis` | **Samarkand Lapis** | Timurid lapis azure & antique gold leaf | Accent `#74A8F8`, Secondary/Fav `#E3BC64`, Success `#6FC79A`, Text `#E6ECF6`, Border `#3D4B66` |
| `madinah` | **Madinah** | Luminous emerald & soft brass gold | Accent `#52CA8E`, Secondary `#DCBC6C`, Text `#E8F0EC`, Border `#344C40`, Live `#E27268` |
| `parchment` | **Parchment** *(light)* | Light-background mushaf paper: walnut ink & ochre | Accent `#9C560C`, Secondary/Success `#1A6B49`, Text `#241E17`, Border `#B8AB96`, Live `#B8332A` |
| `ansi` | **Terminal** *(16 ANSI)* | Strictly 16 ANSI colors (`0`–`15` + default fg); follows any terminal palette | Accent `3`, Accent2 `6`, Success `2`, Text default `""`, Muted/Border `8`, Fav `11`, Live `1` |
| `contrast` | **High Contrast** | Maximum legibility with crisp borders & underlined active focus | Accent `#FFD700`, Accent2 `#00F0FF`, Success `#39FF14`, Text `#FFFFFF`, Border `#B8B8B8`, Live `#FF5555` |

---

### Key Map

All single-key shortcuts pass through `keymap.LatinKey` (so Arabic keyboard layout works by key position without switching layouts); search input receives raw text directly ([keys.go](file:///tmp/claude-1000/-home-nova/b833c7fc-f492-427b-aeb0-b55ff2b3c32f/scratchpad/gemini/wt/internal/ui/keys.go)):

| Area | Keys | Action |
| --- | --- | --- |
| **Navigation** | `1`–`4`, `tab` / `shift+tab` | Switch tabs: Home, Qaris, Surahs, Radio |
| | `j`/`k`, `↑`/`↓`, `pgup`/`pgdn`, `ctrl+u`/`ctrl+d`, `g`/`G` | Move, page, jump top/bottom |
| | `enter` | Open qari/surah or play |
| | `/` | Search current list (English, Arabic, or surah number) |
| | `esc` | Close popup → clear search → go back |
| **Playback** | `space` | Pause / resume (or resume last track when idle) |
| | `←`/`→`, `shift+←`/`shift+→` (`<`/`>`) | Seek ±10 s / ±1 min |
| | `n`/`p` | Next / previous surah in current recording |
| | `+` (`=`) / `-`, `m` | Volume ±5%, mute toggle |
| | `[` / `]` | Speed `0.75x`–`2.0x` (pitch preserved) |
| | `r`, `t`, `x`, `o` | After-surah mode (auto/repeat/stop), sleep timer (15–60m / end of surah), stop, jump to playing |
| **Library** | `f`, `a`, `s`, `w` | Favorite qari, toggle well-known vs all, sort (popular/A–Z/updated), riwayah picker |
| | `e` / `v`, `i` | Choose/cycle recording in Qari Detail; open Qari page from Surah Detail |
| **App** | `c`, `A`, `?`, `q` / `ctrl+c` | Cycle theme live, toggle Arabic surah names, help overlay, quit (saves state) |

---

### Verification & Notes

- **Automated verification**: `gofmt -w .`, `go vet ./...`, and `go test -count=1 ./...` all pass — including exact `W × H` checks, safe-glyph & no-background-fill audits, bidi isolate balance checks, and no-qari-Arabic checks across all 7 themes, 18 screen/overlay states, and 6 terminal sizes (`60×16` to `200×60`), verified against both self-contained fixtures and the live 244-qari / 114-surah / 177-radio API cache.
- **Live tmux verification**: Tested live via `--proxy socks5://127.0.0.1:40000` with isolated state/cache at `130×34` and `64×18`, covering Home, Qaris, Qari Detail, Recording Picker, Surahs, Surah Detail, Radio, Riwayah Picker, 2-column Help, theme switching, favorites, and live audio start/pause/stop.
- **Not verified visually**: Physical rendering inside a GUI light-background terminal emulator (e.g., Konsole with a light profile or translucent wallpaper) and `mpv-mpris` external media keys were not inspected directly outside headless tmux; `parchment` and `ansi` color tokens and no-background-fill invariants are verified programmatically in [ui_test.go](file:///tmp/claude-1000/-home-nova/b833c7fc-f492-427b-aeb0-b55ff2b3c32f/scratchpad/gemini/wt/internal/ui/ui_test.go).
