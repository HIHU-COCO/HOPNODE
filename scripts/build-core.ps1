# 文件用途：可重复构建 Windows Go 基础程序。
# 创建者：HIHU-COCO；创建日期：2026-10-19
$ErrorActionPreference = 'Stop'
$taskRoot = Split-Path -Parent $PSScriptRoot
Push-Location $taskRoot
try {
  $taskGo = Get-Command go -ErrorAction SilentlyContinue
  if ($taskGo) { $taskGoPath = $taskGo.Source } else { $taskGoPath = Join-Path $taskRoot '.tools/go/bin/go.exe' }
  if (!(Test-Path -LiteralPath $taskGoPath)) { throw 'Install Go 1.27.2 or place it under .tools/go.' }
  New-Item -ItemType Directory -Force artifacts/windows | Out-Null
  foreach ($taskName in 'hopnode-core','hopnodectl','hopnode-directory') {
    & $taskGoPath build -trimpath -o "artifacts/windows/$taskName.exe" "./cmd/$taskName"
    if ($LASTEXITCODE -ne 0) { throw "Build failed: $taskName" }
  }
} finally { Pop-Location }
