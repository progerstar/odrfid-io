[CmdletBinding()]
param([string] $DistDir = "")

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$RootDir = Split-Path -Parent $PSScriptRoot
$AppName = "odrfid-io"
$Version = if ($env:APP_VERSION) { $env:APP_VERSION } else { "dev" }
if ($Version -cnotmatch '^[A-Za-z0-9][A-Za-z0-9._-]*$') {
    throw "Invalid APP_VERSION: $Version"
}
if (-not $DistDir) { $DistDir = Join-Path $RootDir "dist\windows" }
if (-not [IO.Path]::IsPathRooted($DistDir)) { $DistDir = Join-Path $RootDir $DistDir }
New-Item -ItemType Directory -Force -Path $DistDir | Out-Null

$BuildDir = Join-Path ([IO.Path]::GetTempPath()) ("odrfid-io-" + [guid]::NewGuid().ToString("N"))
$PackageDir = Join-Path $BuildDir $AppName
$Binary = Join-Path $PackageDir "$AppName.exe"
$ArchiveName = "$AppName-$Version-windows-x86_64.zip"
$Archive = Join-Path $DistDir $ArchiveName

try {
    New-Item -ItemType Directory -Force -Path $PackageDir | Out-Null
    $env:CGO_ENABLED = "1"
    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    $env:CC = "gcc"
    Push-Location $RootDir
    try {
        & go build -trimpath -buildvcs=false -ldflags '-s -w' -o $Binary .
        if ($LASTEXITCODE -ne 0) { throw "go build failed: $LASTEXITCODE" }
    }
    finally { Pop-Location }

    & $Binary -h 2>&1 | Out-Null
    if ($LASTEXITCODE -ne 0) { throw "Executable help smoke test failed: $LASTEXITCODE" }

    Copy-Item -LiteralPath (Join-Path $RootDir "README.md"), (Join-Path $RootDir "LICENSE"), (Join-Path $RootDir "start.bat") -Destination $PackageDir
    if (Test-Path -LiteralPath $Archive) { Remove-Item -LiteralPath $Archive -Force }
    Compress-Archive -Path $PackageDir -DestinationPath $Archive -CompressionLevel Optimal

    $Hash = (Get-FileHash -Algorithm SHA256 -LiteralPath $Archive).Hash.ToLowerInvariant()
    [IO.File]::WriteAllText("$Archive.sha256", "$Hash  $ArchiveName`n", [Text.ASCIIEncoding]::new())
    Write-Host "Created $Archive"
}
finally {
    if (Test-Path -LiteralPath $BuildDir) { Remove-Item -LiteralPath $BuildDir -Recurse -Force }
}
