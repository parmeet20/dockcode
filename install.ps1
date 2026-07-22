$ErrorActionPreference = "Stop"

$Repo = "parmeet20/dockcode"
$Binary = "dockcode.exe"

$InstallDir = "$env:LOCALAPPDATA\Programs\DockCode"

Write-Host "🔍 Fetching latest DockCode release..."

$Release = Invoke-RestMethod `
    "https://api.github.com/repos/$Repo/releases/latest"

$Version = $Release.tag_name

$Asset = "dockcode_windows_amd64.tar.gz"

$Url = "https://github.com/$Repo/releases/download/$Version/$Asset"

Write-Host "📦 Downloading $Asset..."

$Temp = Join-Path $env:TEMP "dockcode-install"

if (Test-Path $Temp) {
    Remove-Item $Temp -Recurse -Force
}

New-Item -ItemType Directory -Path $Temp | Out-Null

$Archive = Join-Path $Temp "dockcode.tar.gz"

Invoke-WebRequest `
    -Uri $Url `
    -OutFile $Archive

Write-Host "📂 Extracting..."

tar -xzf $Archive -C $Temp

if (!(Test-Path "$Temp\$Binary")) {
    throw "dockcode.exe not found in archive"
}

New-Item `
    -ItemType Directory `
    -Path $InstallDir `
    -Force | Out-Null

Copy-Item `
    "$Temp\$Binary" `
    "$InstallDir\$Binary" `
    -Force

Write-Host ""
Write-Host "✅ DockCode installed!"
Write-Host ""
Write-Host "Installed at:"
Write-Host "$InstallDir\$Binary"
Write-Host ""

Write-Host "Add this folder to PATH:"
Write-Host "$InstallDir"