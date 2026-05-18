# commit-on-stop — commit tracked changes at session end so work survives restarts.
# Uses `commit -am` (tracked files only); the agent git-adds new source files itself.
$ErrorActionPreference = 'SilentlyContinue'
$inside = (git rev-parse --is-inside-work-tree 2>$null)
if ($inside -eq 'true') {
  $changes = (git status --porcelain 2>$null)
  if ($changes) {
    git commit -am ("session checkpoint: " + (Get-Date -Format 'yyyy-MM-dd HH:mm')) 2>$null | Out-Null
  }
}
exit 0
