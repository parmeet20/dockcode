$ErrorActionPreference = "Stop"

$Repo = "parmeet20/dockcode"
$Binary = "dockcode.exe"

$InstallDir = "$env:LOCALAPPDATA\Programs\DockCode"

Write-Host "🔍 Fetching latest DockCode release..."

$Release = Invoke-RestMethod `
    "https://api.github.com/repos/$Repo/releases/latest"

$Version = $Release.tag_name

$Asset = "dockcode_windows_amd64.tar.gz"

$DownloadUrl = "https://github.com/$Repo/releases/download/$Version/$Asset"

Write-Host "📦 Downloading $Asset..."

$TempDir = Join-Path $env:TEMP "dockcode-install"

if (Test-Path $TempDir) {
    Remove-Item $TempDir -Recurse -Force
}

New-Item `
    -ItemType Directory `
    -Path $TempDir `
    | Out-Null


$Archive = Join-Path $TempDir "dockcode.tar.gz"


Invoke-WebRequest `
    -Uri $DownloadUrl `
    -OutFile $Archive


Write-Host "📂 Extracting..."

tar -xzf $Archive -C $TempDir


$SourceBinary = Join-Path $TempDir $Binary


if (!(Test-Path $SourceBinary)) {
    throw "❌ $Binary not found inside release archive"
}


Write-Host "🚀 Installing DockCode..."


New-Item `
    -ItemType Directory `
    -Path $InstallDir `
    -Force `
    | Out-Null


Copy-Item `
    $SourceBinary `
    "$InstallDir\$Binary" `
    -Force



# ----------------------------------------
# Add DockCode permanently to User PATH
# ----------------------------------------

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

    Write-Host "✅ Added DockCode to permanent PATH"

}
else {

    Write-Host "✅ DockCode already exists in PATH"

}



# Refresh PATH for current PowerShell session

$env:Path = (
    [Environment]::GetEnvironmentVariable("Path", "Machine") +
    ";" +
    [Environment]::GetEnvironmentVariable("Path", "User")
)



# Cleanup

Remove-Item $TempDir -Recurse -Force



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