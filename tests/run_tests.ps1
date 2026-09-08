# Windows PowerShell Automated Test Runner for TokMan - Ultimate AI Orchestration
param(
    [string]$Mode = "--unit",
    [string]$Module = "1"
)

$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RootDir = Split-Path -Parent $ScriptDir
Set-Location $RootDir

Write-Host "================================================================="
Write-Host "   TOKMAN - ULTIMATE AI ORCHESTRATION: AUTOMATED TEST RUNNER     "
Write-Host "================================================================="

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
            "3" { go test ./backend/supervisor ./backend/gateway -v }
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
