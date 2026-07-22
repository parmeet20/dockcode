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

New-Item `
    -ItemType Directory `
    -Path $Temp `
    | Out-Null

$Archive = Join-Path $Temp "dockcode.tar.gz"

Invoke-WebRequest `
    -Uri $Url `
    -OutFile $Archive

Write-Host "📂 Extracting..."

tar -xzf $Archive -C $Temp

$BinaryPath = Join-Path $Temp $Binary

if (!(Test-Path $BinaryPath)) {
    throw "❌ $Binary not found inside archive"
}

Write-Host "🚀 Installing DockCode..."

New-Item `
    -ItemType Directory `
    -Path $InstallDir `
    -Force `
    | Out-Null

Copy-Item `
    $BinaryPath `
    (Join-Path $InstallDir $Binary) `
    -Force


# Add DockCode permanently to user PATH
Write-Host "🔧 Updating PATH..."

$UserPath = [Environment]::GetEnvironmentVariable(
    "Path",
    "User"
)

if ($UserPath -notlike "*$InstallDir*") {

    if ([string]::IsNullOrEmpty($UserPath)) {
        $NewPath = $InstallDir
    }
    else {
        $NewPath = "$UserPath;$InstallDir"
    }

    [Environment]::SetEnvironmentVariable(
        "Path",
        $NewPath,
        "User"
    )

    Write-Host "✅ Added DockCode to PATH"
}
else {
    Write-Host "✅ DockCode already exists in PATH"
}


# Update current PowerShell session PATH
if ($env:Path -notlike "*$InstallDir*") {
    $env:Path += ";$InstallDir"
}


# Cleanup
Remove-Item $Temp -Recurse -Force


Write-Host ""
Write-Host "================================="
Write-Host "✅ DockCode installed successfully!"
Write-Host "================================="
Write-Host ""
Write-Host "Version : $Version"
Write-Host "Location: $InstallDir\$Binary"
Write-Host ""
Write-Host "Run:"
Write-Host "  dockcode"