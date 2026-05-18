# track-read — record evidence files (build/test logs, screenshots) the agent opens.
# verify-gate.ps1 consults this list before allowing test-results.json to be marked.
$ErrorActionPreference = 'SilentlyContinue'
$log = if ($env:VERIFY_READ_LOG) { $env:VERIFY_READ_LOG } else { './.claude/.evidence-reads' }
$raw = [Console]::In.ReadToEnd()
$path = ''
try { $path = ($raw | ConvertFrom-Json).tool_input.file_path } catch { }
if ($path -and (($path -match 'evidence[\\/]') -or ($path -match '\.png$') -or ($path -match '-build\.txt$') -or ($path -match '-test\.txt$'))) {
  if (Test-Path -LiteralPath $path) { Add-Content -LiteralPath $log -Value $path }
}
exit 0
