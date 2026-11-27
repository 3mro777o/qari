<div align="center">

<img src="docs/banner.jpg" alt="qari: Kur'an'ı terminalinizden dinleyin" width="100%">

<p>
<a href="https://github.com/3mro777o/qari/releases/latest"><img src="https://img.shields.io/github/v/release/3mro777o/qari?style=flat-square&color=DFAA4B&label=release" alt="Latest release"></a>
<a href="https://github.com/3mro777o/qari/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/3mro777o/qari/ci.yml?style=flat-square&label=tests" alt="Tests"></a>
<a href="https://github.com/3mro777o/qari/actions/workflows/install.yml"><img src="https://img.shields.io/github/actions/workflow/status/3mro777o/qari/install.yml?style=flat-square&label=install%20check" alt="Install check"></a>
<img src="https://img.shields.io/badge/Linux%20%C2%B7%20macOS%20%C2%B7%20Windows-6BB892?style=flat-square" alt="Linux, macOS, Windows">
<a href="LICENSE"><img src="https://img.shields.io/github/license/3mro777o/qari?style=flat-square&color=AEA596" alt="MIT license"></a>
</p>

[English](README.md) · [العربية](README.ar.md) · **Türkçe** · [Bahasa Indonesia](README.id.md) · [اردو](README.ur.md) · [Français](README.fr.md) · [Deutsch](README.de.md)

</div>

**qari**, terminalinizde yaşayan bir Kur'an oynatıcısıdır. [mp3quran.net](https://mp3quran.net)
üzerindeki 240'tan fazla kâriyi ve 114 surenin tamamını listeler, bunları [mpv](https://mpv.io)
ile çalar, her surede nerede kaldığınızı hatırlar, hıfz için ayetleri tekrarlar ve internetsiz
dinlemek için sureleri diske kaydeder. Hesap yok, reklam yok, takip yok.

<div align="center">
<img src="docs/demo.gif" alt="Kısa bir tur: kâri arama, sure çalma, ayet tekrarı, tema değiştirme" width="90%">
</div>

## İçindekiler

- [Öne çıkanlar](#öne-çıkanlar)
- [Kurulum](#kurulum) · [Güncelleme](#güncelleme) · [Kaldırma](#kaldırma)
- [Başlarken: ilk beş dakika](#başlarken-ilk-beş-dakika)
- [Rehberler](#rehberler): [Ayet tekrarıyla ezberleme](#ayet-tekrarıyla-ezberleme) ·
  [Rastgele çalma ve tarifler](#rastgele-çalma-ve-tarifler) · [Çevrimdışı dinleme](#çevrimdışı-dinleme) ·
  [Komutlar ve Tab tamamlama](#komutlar-ve-tab-tamamlama) · [Radyo, geçmiş ve favoriler](#radyo-geçmiş-ve-favoriler) ·
  [Temalar](#temalar) · [Medya tuşları](#medya-tuşları) · [Arapça metin](#arapça-metin) · [Engellenen ağlar](#engellenen-ağlar)
- [Tüm tuşlar](#tüm-tuşlar) · [Seçenekler](#seçenekler)
- [Sorun giderme](#sorun-giderme)
- [Veriler nerede tutulur](#veriler-nerede-tutulur) · [Gizlilik](#gizlilik) · [Popülerlik verisi](#popülerlik-verisi)
- [Kaynaktan derleme](#kaynaktan-derleme) · [Teşekkürler](#teşekkürler) · [Lisans](#lisans)

## Öne çıkanlar

| | |
| --- | --- |
| 🎙️ **240+ kâri, 114 sure** | Kâriye ya da sureye göre göz atın. Tam kaydı olan tanınmış kâriler, gerçek dinlenme sayılarına göre sıralanarak en üstte gelir. |
| 🔎 **Her yazıyla arama** | `alafasi`, `العفاسي` ve `العَفاسي` aynı kâriyi bulur; `kahf`, `18` ve `الكهف` aynı sureyi bulur. |
| 📖 **Ezberleme (hıfz)** | Bir ayet aralığını tekrarlayın: her ayet N kez, tüm aralık M kez, aralarında bir duraklamayla. |
| ⏯️ **Kaldığınız yeri hatırlar** | Her sure kendi konumunu saklar. *Continue listening* (Dinlemeye devam et) oradan sürdürür; biten surelere ✓ konur. |
| 🎲 **Rastgele çalma** | *Short surahs* (Kısa sureler) ya da *Daily surahs* (Günlük sureler: Kehf, Yâsîn, Mülk…) gibi tarifler veya kendi tarifleriniz. |
| 💾 **Çevrimdışı** | Bir sureyi ya da bütün bir kaydı indirin; sonra internet olmadan diskten çalar. |
| ⌨️ **Tab destekli komutlar** | Kabuktan `qari play alafasi kahf`. Tab; kâri adlarını, sureleri ve istasyonları tamamlar. |
| 📻 **Canlı radyo** | mp3quran.net'in Kur'an radyo istasyonları. |
| 🎨 **Beş tema** | İslam sanatından ilham alır: Mihrab, Al-Andalus, Samarkand Lapis, Madinah, High Contrast. |
| 🖱️ **Klavye ve fare** | Arapça klavye düzeninde de çalışan vim/mpv tarzı tuşlar; ayrıca tıklama, kaydırma ve sürükleyerek ileri/geri sarma. |

## Kurulum

Tek bir komut hem qari'yi **hem de** mpv'yi (qari'nin kullandığı ses motoru) kurar. Komutu bir
terminale yapıştırıp Enter'a basın.

**Linux ve macOS**

```sh
curl -fsSL https://3mro777o.github.io/qari/install.sh | sh
```

**Windows** (Başlat menüsünden *PowerShell*'i açın; yönetici yetkisi gerekmez)

```powershell
irm https://3mro777o.github.io/qari/install.ps1 | iex
```

Ardından `qari` yazıp Enter'a basın. Komut bulunamazsa yeni bir terminal penceresi açın.

<details>
<summary><b>Kurulum komutu tam olarak ne yapıyor?</b></summary>

Bir komutu kabuğa aktarmadan önce ne yaptığını bilmelisiniz. İki betik de kısa; buradan
okuyabilirsiniz: [install.sh](install.sh), [install.ps1](install.ps1).

**Linux** ([install.sh](install.sh))
1. mpv yoksa paket yöneticinizle (`pacman`, `apt`, `dnf` ya da `zypper`) kurar.
   Bu adım parolanızı sorar. Başarısız olursa qari yine de kurulur ve mpv'yi kendiniz eklemenizi söyler.
2. İşlemcinize (x86-64 ya da ARM64) uygun en son qari sürümünü GitHub Releases'tan indirir.
3. İndirilen dosyayı sürümün `checksums.txt` dosyasıyla (SHA-256) karşılaştırır; eşleşmezse durur.
4. `qari` programını her yerde PATH üzerinde olan `/usr/local/bin` klasörüne koyar (parolanızı sorar).
5. Sizde bulunan kabuklar için Tab tamamlamayı kurar: bash, zsh ve fish.

**macOS** ([install.sh](install.sh)): [Homebrew](https://brew.sh) kullanır (önce onu kurun).
Betik bu depoyu bir tap olarak ekler, sonra `brew install qari` çalıştırır; bu da mpv'yi ve Tab
tamamlamayı kurar. Homebrew ikisini de `brew upgrade` ile güncel tutar.

**Windows** ([install.ps1](install.ps1))
1. En son qari sürümünü GitHub Releases'tan indirir ve SHA-256 sağlama toplamını denetler.
2. `qari.exe` dosyasını `%LOCALAPPDATA%\Programs\qari` klasörüne koyar (kullanıcı klasörünüz, yani yönetici yetkisi gerekmez).
3. mpv yoksa [shinchiro](https://github.com/shinchiro/mpv-winbuild-cmake) tarafından hazırlanan
   taşınabilir mpv sürümünü `%LOCALAPPDATA%\Programs\mpv` klasörüne indirir.
4. İki klasörü de kullanıcı PATH'inize ekler; böylece `qari` her yeni terminalde çalışır.
5. PowerShell profilinize Tab tamamlamayı açan tek bir satır ekler.

</details>

> [!NOTE]
> **Windows Akıllı Uygulama Denetimi (Smart App Control)** kod imzası olmayan programları engeller
> ve qari henüz imzalı değil. Windows `qari.exe` dosyasını çalıştırmayı reddederse sebebi budur.
> Bkz. [Sorun giderme](#sorun-giderme).

<details>
<summary><b>Diğer kurulum yolları</b> (Homebrew, Scoop, hazır derleme, Go)</summary>

**Homebrew** (macOS ve Linux):

```sh
brew tap 3mro777o/qari https://github.com/3mro777o/qari
brew trust 3mro777o/qari   # yeni Homebrew sürümleri bunu ister
brew install qari
```

**Scoop** (Windows):

```powershell
scoop bucket add qari https://github.com/3mro777o/qari
scoop bucket add extras
scoop install qari mpv
```

**Hazır derleme:** sisteminize uygun arşivi [Releases](https://github.com/3mro777o/qari/releases)
sayfasından indirin, açın ve `qari` dosyasını `PATH` üzerindeki bir yere koyun. **mpv**'yi kendiniz
kurun (`sudo apt install mpv`, `brew install mpv`, …).

**Go 1.24+ ile** (mpv de gerekir):

```sh
go install github.com/3mro777o/qari@latest
```

</details>

**Terminal:** modern terminallerin hepsi olur. Windows'ta *Windows Terminal* kullanın (Windows 11'de
varsayılan). Arapça adlar için bkz. [Arapça metin](#arapça-metin).

### Güncelleme

Kurulum komutunu yeniden çalıştırın. Her zaman en yeni sürümü indirip eskisinin yerine koyar;
favorileriniz, ilerlemeniz ve indirmeleriniz yerinde kalır. Homebrew kullananlar `brew upgrade qari`
de çalıştırabilir.

qari günde bir kez GitHub'ı denetler ve yeni sürüm çıktığında alt çubukta *"qari X is out"*
(qari X çıktı) gösterir.

### Kaldırma

Verileriniz programdan ayrı tutulur; tertemiz bir kaldırma istiyorsanız onları da silin
(bkz. [Veriler nerede tutulur](#veriler-nerede-tutulur)).

<details>
<summary><b>Linux, macOS, Windows</b></summary>

**Linux**

```sh
sudo rm /usr/local/bin/qari
rm -f ~/.local/share/bash-completion/completions/qari ~/.config/fish/completions/qari.fish
sudo rm -f /usr/local/share/zsh/site-functions/_qari
rm -rf ~/.local/state/qari ~/.cache/qari ~/.local/share/qari   # verileriniz, isteğe bağlı
```

zsh tamamlama `~/.zshrc` dosyasına eklendiyse, `# qari Tab completion` ile biten satırı silin.

**macOS**

```sh
brew uninstall qari && brew untap 3mro777o/qari
rm -rf ~/.local/state/qari ~/Library/Caches/qari ~/.local/share/qari   # verileriniz, isteğe bağlı
```

**Windows** (PowerShell)

```powershell
Remove-Item -Recurse "$env:LOCALAPPDATA\Programs\qari"
Remove-Item -Recurse "$env:LOCALAPPDATA\Programs\mpv"            # yalnızca kurulum betiği eklediyse
Remove-Item -Recurse "$env:APPDATA\qari", "$env:LOCALAPPDATA\qari" # verileriniz, isteğe bağlı
```

Ardından iki klasörü kullanıcı PATH'inizden (*Ayarlar → Sistem → Hakkında → Gelişmiş sistem
ayarları → Ortam Değişkenleri*) ve `qari completion` satırını profilinizden (`notepad $PROFILE`) kaldırın.

</details>

## Başlarken: ilk beş dakika

1. **Açın.** `qari` yazın. **Home** (Ana sayfa) ekranına gelirsiniz: *Continue listening*, favori
   kârileriniz ve en çok dinlenen kâriler. Üstteki sekmeler: `1` Home, `2` Qaris (Kâriler),
   `3` Surahs (Sureler), `4` Radio, `5` Downloads (İndirilenler), `6` History (Geçmiş), `7` Settings. Geçmek için
   bir rakama ya da `tab` tuşuna basın.

   ![Ana sayfa: dinlemeye devam et, favori kâriler ve en çok dinlenen kâriler](docs/home.png)

2. **Bir kâri bulun.** `2`'ye, sonra `/`'ye basın ve adın bir kısmını yazın, örneğin `alafasi` ya
   da `العفاسي`. `j`/`k` ya da ok tuşlarıyla gezinin, `enter` ile açın.
3. **Bir sure çalın.** Kârinin sayfasında sureleri listelenir. Birini seçip `enter`'a basın. Alttaki
   oynatıcı sureyi, süreyi ve bir ilerleme çubuğunu gösterir. `space` duraklatır, `←`/`→` 10 saniye
   sarar, `n`/`p` sonraki ya da önceki sureye geçer.

   ![Kehf suresi çalarken bir kârinin sayfası; oynatıcı çerçevesinde surenin adı yazar](docs/playing.png)

4. **Kendinize göre ayarlayın.** `f` kâriyi favorilere ekler, `[`/`]` hızı değiştirir, `t` uyku
   zamanlayıcısı kurar, `c` renkleri değiştirir.
5. **`q` ile çıkın.** Kaldığınız yer kaydedilir. Bir dahaki sefere *Continue listening* ya da
   `space` oradan devam eder.

Kayboldunuz mu? Bulunduğunuz ekrandaki tüm tuşlar için `?`'ye basın.

## Rehberler

### Ayet tekrarıyla ezberleme

Birkaç ayeti tekrar tekrar dinlemek, ezberlemenin (hıfz) bilinen yoludur. qari yaklaşık 115 kayıt
için her ayetin nerede başlayıp bittiğini bilir; böylece tam istediğiniz ayetleri döngüye alabilir.

1. Bir sure çalın (örneğin Mişari el-Afasi'den Kehf suresi).
2. `e`'ye basın. Bir pencere açılır:

   | Alan | Anlamı |
   | --- | --- |
   | **From ayah / To ayah** | Tekrarlanacak aralık (başlangıç ve bitiş ayeti). |
   | **Each ayah** | Her ayetin sonrakine geçmeden kaç kez çalacağı (1–99). |
   | **Whole range** | Tüm aralığın kaç kez döneceği (`0` = siz durdurana kadar). |
   | **Gap** | Her tekrardan önceki sessizlik (saniye): ayeti kendiniz okumanız için süre. |

   Alanlar arasında `j`/`k` ile gezinin, değeri `←`/`→` ile değiştirin ya da sayıyı yazın, sonra
   başlatmak için `enter`'a basın.

   ![Ayet tekrarı penceresi](docs/repeat.png)

3. Oynatıcı nerede olduğunuzu gösterir, örneğin `ayah 3/10 · 2 of 3`. Durdurmak için `esc`'ye basın.

**Dinlerken hızlı tekrar:** istediğiniz ilk ayet başladığında `i`'ye, son ayet okunurken yeniden
`i`'ye basın. qari bu aralığı döngüye alır. Üçüncü `i` durdurur. `u` her ayetin kaç kez çalacağını
değiştirir (1, 2, 3, 5, 7, 10).

Ayarlarınız hatırlanır ve tekrar özelliği indirilmiş surelerde internet olmadan da çalışır. Bir
kaydın ayet zamanlamaları yoksa qari bunu söyler ve aynı kârinin zamanlamaları olan başka bir
kaydını gösterir (`v` ile geçin).

### Rastgele çalma ve tarifler

Bir **tarif** seçmek için `z`'ye basın: tarif, rastgele neyin çalınacağını tarif eden kayıtlı bir
tanımdır. Bir sure bitince rastgele bir başkası başlar; tarif tükenene kadar aynı kâriden aynı sure
tekrar gelmez. `n` başka birini seçer, `p` geri döner.

Hazır tarifler:

| Tarif | Çaldıkları |
| --- | --- |
| *Anything* (Herhangi) | Tanınmış kâriler, Hafs rivayeti, Mürettel |
| *Favorite qaris* (Favori kâriler) | Yalnızca favorileriniz |
| *Short surahs* (Kısa sureler) | Amme cüzü |
| *Daily surahs* (Günlük sureler) | Kehf, Secde, Yâsîn, Vâkıa, Mülk |
| *Mujawwad* | Mücevved tilavetler |
| *Random radio* / *Favorite radios* | Canlı radyo istasyonları |

`a` ile kendi tarifinizi yapın ya da `c` ile hazır bir tarifi kopyalayıp değiştirin. Bir tarif;
kârilere (hepsi, tanınmışlar, favoriler ya da seçtiğiniz bir liste), rivayete, okuma tarzına
(Mürettel, Mücevved, …) ve surelere (hepsi, Amme cüzü ya da sizin listeniz) göre süzer. Düzenleyicide
`enter` bir alanı değiştirir, `/` listelerde arar, `s` kaydeder.

### Çevrimdışı dinleme

Bir surede `d`'ye basarak onu, bir kayıtta basarak kaydın tamamını indirin (onaylamak için `d`'ye
iki kez basın). **Downloads** sekmesi (`5`) kuyruğu ilerleme ve hızla birlikte gösterir. İndirilen
sure diskten çalar; böylece uçakta da çalışır ve internet kotanızı korur. `x` bir indirmeyi siler
ya da iptal eder.

![İndirilenler: ilerleme ve hızıyla kuyruk ve çevrimdışı dinleme için kaydedilen kayıtlar](docs/downloads.png)

### Komutlar ve Tab tamamlama

Doğrudan kabuktan çalmaya başlayabilirsiniz. Adlar kısmi olabilir, İngilizce ya da Arapça; sure
bir numara ya da bir addır. Bir ad birden çok kâriye uyuyorsa qari onları listeler.

```sh
qari play alafasi 18              # Mişari el-Afasi'den Kehf suresi, kaldığınız yerden
qari play alafasi                 # o kâriden en son dinlediğinize devam
qari radio ajmy                   # canlı bir radyo istasyonu
qari random "short surahs"        # bir tariften rastgele çalma (ad yoksa: son kullanılan)
qari play alafasi yasin --no-ui   # arayüz yok: tek durum satırı, ctrl+c durdurur
qari help                         # tüm komutlar
```

Arayüz çalmaya başlamış olarak ve tüm tuşlar kullanılabilir hâlde açılır. `--no-ui` ile qari
terminalinizin arka planında çalar ve tek bir durum satırı yazar; sonraki sure uygulamadaki gibi
gelir ve kaldığınız yer aynı şekilde kaydedilir.

**Tab tamamlama** kurulum komutu tarafından bash, zsh, fish ve PowerShell için kurulur. `qari play al`
yazıp Tab'a basın: eşleşen kâri adları gelir; bir addan sonra Tab o kârinin surelerini önerir.
Komutları, seçenekleri, radyo istasyonlarını, tarifleri ve temaları da tamamlar.

![Tab tamamlama kârileri ve sureleri listeler](docs/cli.png)

Elle kurmak için `qari completion bash|zsh|fish|powershell` çıktısını kabuğunuzun başlangıç
dosyasında yükleyin. Tamamlama, adları qari'yi bir kez açtıktan sonra öğrenir.

### Radyo, geçmiş ve favoriler

- **Radio** (`4`): mp3quran.net'in canlı Kur'an radyo istasyonları. `f` istasyonu favorilere ekler.
- **History** (`6`): başlattığınız son 200 sure ve istasyon, güne göre gruplanmış ve nerede
  kaldığınızla birlikte. `enter` birini o noktadan yeniden çalar; `x` bir kaydı kaldırır.
- **Settings** (`7`, Ayarlar): tampon (mpv'nin bir surenin ne kadarını bellekte tuttuğu, varsayılan
  25 MB), ses ve hız (hatırlanır), isim dili, Arapça isimler, bildirimler, tema, rivayet, sıralama ve
  tekrar. `enter` ya da `→` bir satırı değiştirir, `←` geri alır. `--lang` gibi seçenekler önceliklidir.
- **Favoriler:** bir kâride `f`, onu Home'a ve *Favorite qaris* tarifine ekler.
- **Herkesi göster:** varsayılan olarak tam kaydı olan tanınmış kârileri görürsünüz. `a` eksik ve
  daha az bilinen kayıtlar dahil hepsini gösterir. `s` popülerliğe, ada ya da son güncellenmeye
  göre sıralar, `w` rivayete göre süzer (Hafs, Verş, Kalun, …), `v` kârinin kayıtlarından birini seçer.

### Temalar

`c`'ye basın: siz gezindikçe ekranın tamamı her temanın önizlemesini gösterir, `enter` seçili temayı
tutar, `esc` geri döner. Hiçbir tema arka plan boyamaz; saydam bir terminal saydam kalır.

| Tema | Görünüm |
| --- | --- |
| `mihrab` | **Mihrab** (varsayılan): sıcak yaldızlı altın, kaymaktaşı ve adaçayı yeşili |
| `andalus` | **Al-Andalus**: Elhamra pişmiş toprağı, sırlı turkuaz ve fildişi |
| `lapis` | **Samarkand Lapis**: Timurlu lacivert taşı mavisi, altın varak ve ay ışığında ipek |
| `madinah` | **Madinah**: ışıltılı zümrüt, yumuşak pirinç ve sıcak inci |
| `contrast` | **High Contrast**: beyaz, canlı altın ve elektrik camgöbeği |

![Al-Andalus'u önizleyen tema seçici](docs/theme.png)

### Medya tuşları

Linux'ta qari kendini bir MPRIS oynatıcısı (`org.mpris.MediaPlayer2.qari`) olarak kaydeder; böylece
klavyedeki medya tuşları, masaüstü bileşenleri ve `playerctl -p qari play-pause` onu denetler ve
her yeni surede bir bildirim gösterilir. macOS ve Windows'ta henüz medya tuşu desteği yok; bildirim
yine de gelir.

### Arapça metin

Adlar varsayılan olarak İngilizce görünür ve her zaman Arapça arayabilirsiniz: `البقره` Bakara
suresini bulur. `--lang`, kâri ve sure adlarını API'nin sunduğu 21 dilden herhangi birine çevirir
(`--lang ar`, `--lang fr`, `--lang ur`, …; `--languages` bunları listeler). `--arabic`, İngilizce
sure adlarının yanına Arapçalarını da ekler.

Arapça için sağdan sola yazıyı destekleyen bir terminal gerekir: Konsole, GNOME Terminal ve diğer
VTE terminalleri, WezTerm, mlterm, Windows Terminal. Diğer terminallerde Arapça harfler ters ya da
kopuk görünebilir.

Kısayollar tuşun konumunu izler, bu yüzden Arapça klavye düzeninde de çalışır: `ض`, `q` gibi çıkar;
`ظ`, `/` gibi arar.

### Engellenen ağlar

Bazı okul ve iş yeri ağları mp3quran.net'i engeller. Listeler yüklenmiyorsa kullanmanıza izin
verilen bir proxy kullanın:

```sh
qari --proxy socks5://127.0.0.1:1080
```

mpv SOCKS proxy kullanamaz; bu yüzden qari sesi mpv'ye küçük bir yerel aktarıcı üzerinden iletir.
İleri/geri sarma yine çalışır ve indirmeler de aynı proxy'yi kullanır.

## Tüm tuşlar

Bu liste için uygulamada `?`'ye basın. Tuşlar mpv, cmus ve vim'i izler.

| Tuş | İşlev |
| --- | --- |
| `1`–`7`, `tab` / `shift+tab` | Home, Qaris, Surahs, Radio, Downloads, History, Settings |
| `j` `k` / `↓` `↑` | Gezin |
| `ctrl+d` `ctrl+u` / `pgdn` `pgup` | Yarım sayfa aşağı / yukarı |
| `g` `G` / `home` `end` | İlk / son satır |
| `enter` / `l` | Aç ya da çal |
| `esc` / `h` | Geri (önce aramayı temizler) |
| `/` | Geçerli listede ara (İngilizce, Arapça ya da sure numarası) |
| `o` | Çalmakta olana git |
| `space` | Çal / duraklat (boştayken son surenize devam eder) |
| `←` `→` | 10 saniye sar |
| `,` `.` (ya da `shift+←` `shift+→`) | 1 dakika sar |
| `n` `p` | Sonraki / önceki sure (rastgele çalmada: başka seçim / geri) |
| `z` | Rastgele: `enter` bir tarifi çalar; `a` yeni, `e` düzenle, `c` kopyala, `x` sil (iki kez basın) |
| `+` `-` | Ses düzeyi |
| `m` | Sessiz |
| `[` `]` | Yavaş / hızlı (0.75×–2×, ses perdesi değişmez) |
| `r` | Sure bitince: sonrakini otomatik çal, tekrarla ya da dur |
| `t` | Uyku zamanlayıcısı (15–60 dk ya da sure sonu) |
| `e` | Ezber için ayet tekrarı |
| `i` | Hızlı tekrar: ilk ayet, son ayet, dur |
| `u` | Tekrarlanan her ayetin kaç kez çalacağı (1–10) |
| `f` | Favori kâri ya da istasyon |
| `a` | Tanınmış kâriler / tüm kâriler |
| `s` | Kârileri sırala: popüler, A–Z, son güncellenen |
| `w` | Rivayete göre süz |
| `v` | Kârinin kaydını seç |
| `d` | Sureyi ya da kaydın tamamını indir (yeniden basmanızı ister) |
| `x` | İndirmeyi sil ya da iptal et; History'den bir kaydı kaldır |
| `c` | Renk teması |
| `?` | Yardım |
| `q` / `ctrl+c` | Çık (kaldığınız yer kaydedilir) |

**Fare:** sekmelere ve satırlara tıklayın, açmak ya da çalmak için bir satıra yeniden tıklayın,
tekerlekle kaydırın, sarmak için ilerleme çubuğuna tıklayın ya da onu sürükleyin.

![Yardım ekranı tüm tuşları listeler](docs/help.png)

## Seçenekler

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

`QARI_LANG`, `QARI_THEME` ve `QARI_PROXY` aynı varsayılanları ortam değişkenlerinden ayarlar.

## Sorun giderme

<details>
<summary><b>Windows: "An Application Control policy has blocked this file" / Akıllı Uygulama Denetimi</b></summary>

Akıllı Uygulama Denetimi (Smart App Control) kod imzası olmayan her programı engeller ve qari henüz
imzalı değil (açık kaynak projeler için ücretsiz kod imzalama başvurusu yapıldı). O zamana kadar tek
yol, *Windows Güvenliği → Uygulama ve tarayıcı denetimi → Akıllı Uygulama Denetimi ayarları*
bölümünden bu özelliği kapatmaktır. Çoğu bilgisayarda zaten kapalıdır; yalnızca bazı yeni Windows 11
kurulumlarında açıktır.

</details>

<details>
<summary><b>Kurulumdan hemen sonra "qari: command not found"</b></summary>

Yeni bir terminal açın: eskisi güncellenen PATH'i görmez. Linux'ta `/usr/local/bin` klasörünün
PATH'inizde olduğunu denetleyin (`echo $PATH`).

</details>

<details>
<summary><b>"Install mpv to play audio" / ses yok</b></summary>

qari mpv üzerinden çalar. mpv'yi paket yöneticinizle kurun (`sudo apt install mpv`,
`sudo pacman -S mpv`, `brew install mpv`) ya da kurulum komutunu yeniden çalıştırın. Yeni bir
terminalde `mpv --version` komutunun çalıştığını denetleyin.

</details>

<details>
<summary><b>Listeler yüklenmiyor</b></summary>

mp3quran.net büyük olasılıkla ağınızda engelli; bkz. [Engellenen ağlar](#engellenen-ağlar). qari son
yüklenen listeleri çevrimdışı kullanım için saklar ve indirilmiş sureler her zaman çalar.

</details>

<details>
<summary><b>Arapça ters ya da kopuk görünüyor</b></summary>

Terminaliniz sağdan sola yazıyı desteklemiyor. [Arapça metin](#arapça-metin) bölümündeki
terminallerden birini kullanın ya da `--arabic` ve `--lang ar` olmadan çalıştırın.

</details>

<details>
<summary><b>Tab hiçbir şey göstermiyor</b></summary>

Listeleri kaydetmesi için qari'yi bir kez açın; tamamlama adları bu listelerden okur. bash'te
`bash-completion` paketinin kurulu olması gerekir. Kurulumdan sonra yeni bir terminal başlatın.

</details>

<details>
<summary><b>macOS: "Refusing to load formula from untrusted tap"</b></summary>

`brew trust 3mro777o/qari` çalıştırın, sonra yeniden kurun. Kurulum komutu bunu sizin yerinize yapar.

</details>

Başka bir sorun mu var? Sisteminizi ve ne gördüğünüzü yazarak
[bir issue açın](https://github.com/3mro777o/qari/issues).

## Veriler nerede tutulur

| Ne | Linux / macOS | Windows |
| --- | --- | --- |
| Favoriler, ilerleme, tarifler, geçmiş, ayarlar | `~/.local/state/qari/state.json` | `%AppData%\qari\state.json` |
| Katalog önbelleği | `~/.cache/qari/` (macOS'ta `~/Library/Caches/qari/`) | `%LocalAppData%\qari\` |
| İndirilen sureler | `~/.local/share/qari/` | `%LocalAppData%\qari\` |

Linux'ta `XDG_STATE_HOME`, `XDG_CACHE_HOME` ve `XDG_DATA_HOME` değişkenleri bu klasörlerin yerini
değiştirir.

## Gizlilik

qari'de hesap yoktur ve hiçbir analiz verisi göndermez. Yalnızca iki yerle konuşur: mp3quran.net
(listeler, ses ve ayet zamanlamaları) ve GitHub (günde bir kez, daha yeni bir sürüm çıkıp çıkmadığına
bakmak için). Hatırladığı her şey bilgisayarınızda kalır.

## Popülerlik verisi

mp3quran.net dinlenme sayısı yayımlamaz. Popülerlik,
[`tools/genpopular`](tools/genpopular/main.go) tarafından oluşturulan
[`internal/popular/popular.json`](internal/popular/popular.json) anlık görüntüsünden gelir ve şu
kaynaklara dayanır:

- [way2quran.com](https://www.way2quran.com/en/reciters?sortBy=-totalListens) tarafından yayımlanan **dinlenme sayıları**
- **"tanınmış"**: kâri [quranicaudio.com](https://quranicaudio.com) ya da
  [quran.com](https://quran.com) sitesinde de bulunuyorsa ya da way2quran onu en iyi kâriler arasında
  gösteriyorsa

Kâriler siteler arasında Arapça ve İngilizce adlarıyla eşleştirilir. Anlık görüntüyü
`go run ./tools/genpopular` ile yenileyin.

## Kaynaktan derleme

Go 1.24+ ve mpv gerekir.

```sh
git clone https://github.com/3mro777o/qari && cd qari
go build -o qari . && ./qari
go test ./...
```

Kod: `internal/api` (mp3quran.net istemcisi ve önbelleği), `internal/player` (JSON IPC üzerinden mpv),
`internal/store` (kayıtlı durum), `internal/ui` ([Bubble Tea](https://github.com/charmbracelet/bubbletea)
arayüzü). Sürümler, bir `v*` etiketi gönderildiğinde [GoReleaser](https://goreleaser.com) ile derlenir.

## Teşekkürler

Tilavetler, adlar, radyo ve ayet zamanlamaları: [mp3quran.net](https://mp3quran.net). Popülerlik
verisi: way2quran.com, quranicaudio.com ve quran.com.
[Bubble Tea](https://github.com/charmbracelet/bubbletea) ve [mpv](https://mpv.io) ile yapıldı.

## Lisans

[MIT](LICENSE). Çeviriler İngilizce sürümün gerisinde kalabilir; aralarında fark olursa İngilizce
README esas alınır.
