# steer — surface STEER.md content to the agent once, then clear it.
# Write to STEER.md to redirect the agent mid-run without restarting.
$ErrorActionPreference = 'SilentlyContinue'
$f = if ($env:AGENT_STEER_FILE) { $env:AGENT_STEER_FILE } else { './STEER.md' }
if ((Test-Path -LiteralPath $f) -and ((Get-Item -LiteralPath $f).Length -gt 0)) {
  $note = (Get-Content -LiteralPath $f -Raw)
  $reason = "OPERATOR STEERING: $note`n`nPause what you were about to do, incorporate this guidance, then continue toward the goal."
  Write-Output (@{ decision = 'block'; reason = $reason } | ConvertTo-Json -Compress)
  Clear-Content -LiteralPath $f
}
exit 0
