# verify-gate — deny writes to test-results.json unless evidence was Read this cycle.
# Teaching example, not a security boundary: only guards Write/Edit, basename match.
$ErrorActionPreference = 'SilentlyContinue'
$log = if ($env:VERIFY_READ_LOG) { $env:VERIFY_READ_LOG } else { './.claude/.evidence-reads' }
$results = if ($env:RESULTS_FILE) { $env:RESULTS_FILE } else { 'test-results.json' }
$raw = [Console]::In.ReadToEnd()
$target = ''
try { $target = ($raw | ConvertFrom-Json).tool_input.file_path } catch { }
if (-not $target) { exit 0 }
if ((Split-Path -Leaf $target) -ne $results) { exit 0 }
$hasEvidence = (Test-Path -LiteralPath $log) -and ((Get-Item -LiteralPath $log).Length -gt 0)
if (-not $hasEvidence) {
  Write-Output '{"decision":"block","reason":"Cannot modify test-results.json: no evidence (build log, test output, or screenshot) has been Read this cycle. Open the evidence file with the Read tool first, then retry."}'
  exit 0
}
Clear-Content -LiteralPath $log   # consume evidence; the next change needs fresh proof
exit 0
