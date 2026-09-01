<#
.SYNOPSIS
    Cross-compile CATGEN and assemble one clean release folder + zip per target.

.DESCRIPTION
    Each package is a self-contained folder — the binary, the bundled images/,
    an empty exports/ for the user's output, and README.md. No source ships.

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
    @{ OS = "windows"; Arch = "amd64"; Ext = ".exe" }
    @{ OS = "darwin";  Arch = "amd64"; Ext = "" }
    @{ OS = "darwin";  Arch = "arm64"; Ext = "" }
    @{ OS = "linux";   Arch = "amd64"; Ext = "" }
    @{ OS = "linux";   Arch = "arm64"; Ext = "" }
)

$env:CGO_ENABLED = "0"
try {
    foreach ($t in $targets) {
        $name  = "catgen-$Version-$($t.OS)-$($t.Arch)"
        $stage = Join-Path $dist $name
        New-Item -ItemType Directory -Force $stage | Out-Null

        $env:GOOS = $t.OS
        $env:GOARCH = $t.Arch
        Write-Host "building $($t.OS)/$($t.Arch) ..."
        & go build -trimpath -ldflags "-s -w -X main.version=$Version" `
            -o (Join-Path $stage "catgen$($t.Ext)") ./cmd/catgen
        if ($LASTEXITCODE -ne 0) { throw "go build failed for $($t.OS)/$($t.Arch)" }

        Copy-Item -Recurse (Join-Path $root "images") (Join-Path $stage "images")
        Copy-Item (Join-Path $root "README.md") $stage
        New-Item -ItemType Directory -Force (Join-Path $stage "exports") | Out-Null
        Set-Content -Path (Join-Path $stage "exports/README.txt") -Encoding utf8 `
            -Value "Exported .txt / .png art from the studio lands in this folder."

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
