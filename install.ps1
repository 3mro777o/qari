# Install or update qari (and mpv) on Windows, no admin needed:
#   irm https://3mro777o.github.io/qari/install.ps1 | iex
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue' # the progress bar makes downloads very slow
$repo = '3mro777o/qari'
$arch = if ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64') { 'arm64' } else { 'amd64' }
$root = Join-Path $env:LOCALAPPDATA 'Programs'
$qdir = Join-Path $root 'qari'
$mdir = Join-Path $root 'mpv'
$tmp = Join-Path $env:TEMP "qari-install-$PID"
New-Item -ItemType Directory -Force $tmp, $qdir | Out-Null

try {
	$rel = Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest"
	$ver = $rel.tag_name.TrimStart('v')
	$zip = "qari_${ver}_windows_$arch.zip"
	Write-Host "Downloading qari $ver..."
	Invoke-WebRequest "https://github.com/$repo/releases/download/$($rel.tag_name)/$zip" -OutFile "$tmp\$zip"
	Invoke-WebRequest "https://github.com/$repo/releases/download/$($rel.tag_name)/checksums.txt" -OutFile "$tmp\checksums.txt"
	$want = (Get-Content "$tmp\checksums.txt" | Where-Object { $_ -match " $([regex]::Escape($zip))$" }) -replace ' .*', ''
	if ((Get-FileHash "$tmp\$zip" -Algorithm SHA256).Hash -ne $want.Trim().ToUpper()) { throw 'checksum mismatch' }
	Expand-Archive "$tmp\$zip" $tmp -Force
	Get-Process qari -ErrorAction SilentlyContinue | Where-Object Path -like "$qdir*" | Stop-Process -Force
	Copy-Item "$tmp\qari.exe" $qdir -Force

	if (-not (Get-Command mpv -ErrorAction SilentlyContinue) -and -not (Test-Path "$mdir\mpv.exe")) {
		Write-Host 'Downloading mpv...'
		$mrel = Invoke-RestMethod 'https://api.github.com/repos/shinchiro/mpv-winbuild-cmake/releases/latest'
		$name = if ($arch -eq 'arm64') { '^mpv-aarch64-\d' } else { '^mpv-x86_64-\d' }
		$asset = $mrel.assets | Where-Object name -match $name | Select-Object -First 1
		Invoke-WebRequest $asset.browser_download_url -OutFile "$tmp\mpv.7z"
		New-Item -ItemType Directory -Force $mdir | Out-Null
		tar -xf "$tmp\mpv.7z" -C $mdir # Windows tar (bsdtar) reads 7z
	}

	# PATH for new terminals, and for this one so "qari" works right away.
	$path = [Environment]::GetEnvironmentVariable('Path', 'User')
	foreach ($d in $qdir, $mdir) {
		if ((Test-Path $d) -and ($path -split ';') -notcontains $d) { $path = "$path;$d".TrimStart(';') }
		if (($env:Path -split ';') -notcontains $d) { $env:Path += ";$d" }
	}
	[Environment]::SetEnvironmentVariable('Path', $path, 'User')

	# Tab completion: one line in the PowerShell profile. Older releases would open the player
	# instead, so check --help first (it prints to stderr, which must not stop the script).
	$help = & { $ErrorActionPreference = 'Continue'; & "$qdir\qari.exe" --help 2>&1 | Out-String }
	if ($help -match 'completion') {
		$line = 'if (Get-Command qari -ErrorAction SilentlyContinue) { qari completion powershell | Out-String | Invoke-Expression }'
		if (-not (Test-Path $PROFILE)) { New-Item -ItemType File -Force $PROFILE | Out-Null }
		if (-not (Select-String -Path $PROFILE -SimpleMatch $line -Quiet)) { Add-Content $PROFILE $line }
	}
	Write-Host "Done: qari $ver. Type: qari (open a new terminal if it is not found)"
} finally {
	Remove-Item $tmp -Recurse -Force -ErrorAction SilentlyContinue
}
