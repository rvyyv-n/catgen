<#
.SYNOPSIS
    Cross-compile CATGEN and assemble the release assets: one standalone
    binary per platform plus one source-code zip per platform.

.DESCRIPTION
    Each binary is a single self-contained file — the 27 sample images are
    embedded via //go:embed (internal/samples), so there is no images/ folder
    or exports/ folder to ship alongside it. --dir / IMAGES_DIR still let a
    user point catgen at a real folder of their own images.

    Targets: Windows (x64), Linux (x64), macOS (Apple Silicon).

    Each source zip is the same git HEAD snapshot (via `git archive`),
    produced once and copied under a platform-labeled name so every binary
    has a matching source download beside it.

.EXAMPLE
    pwsh scripts/release.ps1 -Version v1.0.0
#>
param(
    [string]$Version = "v1.0.0"
)

$ErrorActionPreference = "Stop"
$root = Split-Path $PSScriptRoot -Parent
$dist = Join-Path $root "dist"

Remove-Item -Recurse -Force $dist -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Force $dist | Out-Null

$targets = @(
    @{ GOOS = "windows"; GOARCH = "amd64"; Label = "windows-x64"; Ext = ".exe" }
    @{ GOOS = "linux";   GOARCH = "amd64"; Label = "linux-x64";   Ext = ""     }
    @{ GOOS = "darwin";  GOARCH = "arm64"; Label = "macos-arm64"; Ext = ""     }
)

Write-Host "archiving source (git HEAD) ..."
$srcZip = Join-Path $dist "_src.zip"
& git -C $root archive --format=zip -o $srcZip HEAD
if ($LASTEXITCODE -ne 0) { throw "git archive failed" }

$env:CGO_ENABLED = "0"
try {
    foreach ($t in $targets) {
        $bin = "catgen-$Version-$($t.Label)$($t.Ext)"
        $src = "catgen-$Version-src-$($t.Label).zip"

        Write-Host "building $($t.Label) ..."
        $env:GOOS = $t.GOOS
        $env:GOARCH = $t.GOARCH
        & go build -trimpath -ldflags "-s -w -X main.version=$Version" `
            -o (Join-Path $dist $bin) ./cmd/catgen
        if ($LASTEXITCODE -ne 0) { throw "go build failed for $($t.Label)" }

        Copy-Item $srcZip (Join-Path $dist $src)
        Write-Host "  -> dist/$bin"
        Write-Host "  -> dist/$src"
    }
}
finally {
    Remove-Item Env:GOOS, Env:GOARCH, Env:CGO_ENABLED -ErrorAction SilentlyContinue
}

Remove-Item $srcZip -ErrorAction SilentlyContinue

Write-Host ""
Write-Host "release binaries and source zips are in $dist"
