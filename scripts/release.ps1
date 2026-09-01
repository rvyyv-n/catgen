<#
.SYNOPSIS
    Cross-compile CATGEN and assemble one clean release folder + zip per target.

.DESCRIPTION
    Each package is a self-contained folder the end user can unzip and run:
    the binary, the bundled images/, an empty exports/ for their output, and
    README.md. No source, no build files.

    Targets: Windows (x64, ARM64), Linux (x64, ARM64), macOS (Apple Silicon).
    macOS packages are labelled "macos" for clarity even though Go's GOOS is
    "darwin".

.EXAMPLE
    pwsh scripts/release.ps1 -Version v2.0.0
#>
param(
    [string]$Version = "v2.0.0"
)

$ErrorActionPreference = "Stop"
$root = Split-Path $PSScriptRoot -Parent
$dist = Join-Path $root "dist"

Remove-Item -Recurse -Force $dist -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Force $dist | Out-Null

$targets = @(
    @{ GOOS = "windows"; GOARCH = "amd64"; Label = "windows-x64";   Ext = ".exe" }
    @{ GOOS = "windows"; GOARCH = "arm64"; Label = "windows-arm64"; Ext = ".exe" }
    @{ GOOS = "linux";   GOARCH = "amd64"; Label = "linux-x64";     Ext = "" }
    @{ GOOS = "linux";   GOARCH = "arm64"; Label = "linux-arm64";   Ext = "" }
    @{ GOOS = "darwin";  GOARCH = "arm64"; Label = "macos-arm64";   Ext = "" }
)

$env:CGO_ENABLED = "0"
try {
    foreach ($t in $targets) {
        $name  = "catgen-$Version-$($t.Label)"
        $stage = Join-Path $dist $name
        New-Item -ItemType Directory -Force $stage | Out-Null

        $env:GOOS = $t.GOOS
        $env:GOARCH = $t.GOARCH
        Write-Host "building $($t.Label) ..."
        & go build -trimpath -ldflags "-s -w -X main.version=$Version" `
            -o (Join-Path $stage "catgen$($t.Ext)") ./cmd/catgen
        if ($LASTEXITCODE -ne 0) { throw "go build failed for $($t.Label)" }

        Copy-Item -Recurse (Join-Path $root "images") (Join-Path $stage "images")
        Copy-Item (Join-Path $root "README.md") $stage
        New-Item -ItemType Directory -Force (Join-Path $stage "exports") | Out-Null
        Set-Content -Path (Join-Path $stage "exports/README.txt") -Encoding utf8 `
            -Value "Your exported .txt and .png art is saved here."

        $zip = Join-Path $dist "$name.zip"
        Compress-Archive -Path (Join-Path $stage "*") -DestinationPath $zip -Force
        Write-Host "  -> $zip"
    }
}
finally {
    Remove-Item Env:GOOS, Env:GOARCH, Env:CGO_ENABLED -ErrorAction SilentlyContinue
}

Write-Host ""
Write-Host "release folders and zips are in $dist"
