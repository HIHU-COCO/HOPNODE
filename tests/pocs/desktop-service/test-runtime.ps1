# Purpose: isolated Windows independent process/CLI/Rust IPC smoke test. Creator: HIHU-COCO. Created: 2026-10-19.
param([Parameter(Mandatory=$true)][string]$GoExecutable, [string]$CargoExecutable)
$ErrorActionPreference = 'Stop'
$pocDir = $PSScriptRoot
$testDir = Join-Path ([IO.Path]::GetTempPath()) ('hopnode-p04-' + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $testDir | Out-Null
$binary = Join-Path $testDir 'hopnodectl-poc.exe'
$previousEndpoint = $env:HOPNODE_POC_ENDPOINT
$env:HOPNODE_POC_ENDPOINT = '\\.\pipe\hopnode-poc-test-' + [guid]::NewGuid().ToString('N')
$controller = $null
Push-Location $pocDir
try {
    & $GoExecutable build -o $binary .
    if ($LASTEXITCODE -ne 0) { throw 'POC build failed' }
    $controller = Start-Process -FilePath $binary -ArgumentList 'serve' -PassThru -WindowStyle Hidden -RedirectStandardOutput (Join-Path $testDir 'stdout.txt') -RedirectStandardError (Join-Path $testDir 'stderr.txt')
    $status = $null
    for ($attempt=0; $attempt -lt 30; $attempt++) {
        $output = & $binary status --json 2>$null
        if ($LASTEXITCODE -eq 0) { $status = $output | ConvertFrom-Json; break }
        Start-Sleep -Milliseconds 100
    }
    if (-not $status -or $status.pid -ne $controller.Id -or $status.api_version -ne 'poc.v0' -or -not $status.identity) { throw 'Independent Core status did not match the launched process' }
    Write-Output 'PASS: independent Go controller -> CLI Named Pipe; reported PID and real process identity verified'
    if ($CargoExecutable) {
        $manifest = Join-Path $pocDir '../../../apps/desktop/src-tauri/Cargo.toml'
        & $CargoExecutable test --locked --manifest-path $manifest real_controller_roundtrip -- --ignored --nocapture
        if ($LASTEXITCODE -ne 0) { throw 'Rust -> Go Named Pipe failed' }
    }
    Stop-Process -Id $controller.Id
    $controller.WaitForExit()
    & $binary status --json 2>$null | Out-Null
    if ($LASTEXITCODE -eq 0) { throw 'Disconnected Core returned a success' }
    Write-Output 'PASS: stopped controller produces real CLI failure'
} finally {
    if ($controller -and -not $controller.HasExited) { Stop-Process -Id $controller.Id }
    $env:HOPNODE_POC_ENDPOINT = $previousEndpoint
    Pop-Location
    # Keep this isolated test directory as local evidence. No user data is removed.
}
