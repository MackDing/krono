# kill-switch — halt every tool call while ./AGENT_STOP exists.
# Engage:  New-Item AGENT_STOP    Resume:  Remove-Item AGENT_STOP
$ErrorActionPreference = 'SilentlyContinue'
$stop = if ($env:AGENT_STOP_FILE) { $env:AGENT_STOP_FILE } else { './AGENT_STOP' }
if (Test-Path -LiteralPath $stop) {
  Write-Output '{"decision":"block","reason":"Kill switch engaged: AGENT_STOP file exists. Agent is halted. Remove the file to resume."}'
}
exit 0
