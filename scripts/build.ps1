# Windows PowerShell Build Script for TokMan Gateway Mesh
$ErrorActionPreference = "Stop"

Write-Host "================================================================="
Write-Host "   TOKMAN AI GATEWAY MESH: PRODUCTION NATIVE BINARY COMPILER     "
Write-Host "================================================================="

$rootDir = Split-Path -Parent $PSScriptRoot
Set-Location $rootDir

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Error "Go toolchain not found in PATH. Please install Go from https://go.dev/dl/."
    exit 1
}

New-Item -ItemType Directory -Force -Path "build\bin" | Out-Null

Write-Host "--> Compiling standalone native Windows binary (Gateway + Embedded SQLite + Fyne GUI)..."
go build -trimpath -ldflags="-s -w" -o "build\bin\tokman.exe" .

if (Test-Path "build\bin\tokman.exe") {
    $fileInfo = Get-Item "build\bin\tokman.exe"
    $sizeMB = [math]::Round($fileInfo.Length / 1MB, 1)
    Write-Host "--> Build complete. Binary located at: build\bin\tokman.exe ($sizeMB MB)"
} else {
    Write-Error "Build failed: output binary not found."
    exit 1
}
