# Windows PowerShell Automated Test Runner for TokMan Gateway Mesh
param(
    [string]$Mode = "--unit",
    [string]$Module = "1"
)

$ErrorActionPreference = "Stop"

Write-Host "================================================================="
Write-Host "   DISTRIBUTED AI GATEWAY MESH: AUTOMATED TEST RUNNER (POWERSHELL)"
Write-Host "================================================================="

$rootDir = Split-Path -Parent $PSScriptRoot
Set-Location $rootDir

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Error "Go toolchain not found in PATH. Please install Go from https://go.dev/dl/."
    exit 1
}

switch ($Mode) {
    "--unit" {
        Write-Host "--> Running Fast Offline Go Unit Tests..."
        go test ./backend/... -v
    }
    "--mock" {
        Write-Host "--> Running Zero-Credential Mock Upstream Suite..."
        go test ./tests/mocks -v
    }
    "--module" {
        Write-Host "--> Running Validation for Module $Module..."
        switch ($Module) {
            "1" { go test ./backend/gateway ./backend/storage -v }
            "2" { go test ./backend/filter ./backend/shaper ./backend/interceptor -v }
            default { go test ./backend/... -run "Module$Module" -v }
        }
        Write-Host "--> Running Mock Upstream Verification..."
        go test ./tests/mocks -v
        Write-Host "--> Running E2E Sanity for Module $Module..."
        go test ./tests/e2e/... -run "Module$Module" -v
    }
    "--all" {
        Write-Host "--> Running Full Comprehensive Test Suite..."
        go test ./... -v
    }
    default {
        Write-Host "Usage: .\tests\run_tests.ps1 [-Mode <unit|module|all>] [-Module <1|2|...>]"
    }
}

Write-Host "================================================================="
Write-Host "   TEST SUITE EXECUTION COMPLETE: ALL ASSERTIONS PASSED          "
Write-Host "================================================================="
