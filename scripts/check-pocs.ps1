# 文件用途：分别验证独立 POC module，避免根 go test 漏测嵌套模块。
# 创建者：HIHU-COCO；创建日期：2026-10-19
$ErrorActionPreference = 'Stop'
$taskRoot = Split-Path -Parent $PSScriptRoot
$taskGo = Get-Command go -ErrorAction SilentlyContinue
$taskGoPath = if ($taskGo) { $taskGo.Source } else { Join-Path $taskRoot '.tools/go/bin/go.exe' }
if (!(Test-Path -LiteralPath $taskGoPath)) { throw 'Go 1.27.2 is required' }
foreach ($taskModule in Get-ChildItem -LiteralPath (Join-Path $taskRoot 'tests/pocs') -Directory) {
  if (!(Test-Path -LiteralPath (Join-Path $taskModule.FullName 'go.mod'))) { continue }
  Push-Location $taskModule.FullName
  try {
    Write-Output "Checking POC: $($taskModule.Name)"
    & $taskGoPath test ./...
    if ($LASTEXITCODE -ne 0) { throw "POC test failed: $($taskModule.Name)" }
    & $taskGoPath vet ./...
    if ($LASTEXITCODE -ne 0) { throw "POC vet failed: $($taskModule.Name)" }
  } finally { Pop-Location }
}
Push-Location $taskRoot
try {
  python -m unittest discover -s tests/pocs/network -p 'test_*.py' -v
  if ($LASTEXITCODE -ne 0) { throw 'Network preflight tests failed' }
} finally { Pop-Location }
