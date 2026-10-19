"""文件用途：通过 Git credential helper 查询当前仓库 CI，不输出凭据。创建者：HIHU-COCO；日期：2026-10-19。"""
import json
import subprocess
import sys
import urllib.error
import urllib.request

credential = subprocess.run(["git", "credential", "fill"], input="protocol=https\nhost=github.com\n\n", capture_output=True, text=True, check=False)
fields = dict(line.split("=", 1) for line in credential.stdout.splitlines() if "=" in line)
token = fields.get("password")
if not token:
    raise SystemExit("GitHub authentication not available from credential helper")
url = "https://api.github.com/repos/HIHU-COCO/HOPNODE/actions/runs?per_page=10"
request = urllib.request.Request(url, headers={"Authorization": f"Bearer {token}", "Accept": "application/vnd.github+json", "User-Agent": "HopNode-CI-check"})
try:
    with urllib.request.urlopen(request, timeout=30) as response:
        data = json.load(response)
except urllib.error.HTTPError as error:
    raise SystemExit(f"GitHub API HTTP {error.code}") from None
for run in data.get("workflow_runs", []):
    print(json.dumps({key: run.get(key) for key in ("id", "name", "head_branch", "head_sha", "status", "conclusion", "html_url")}))
