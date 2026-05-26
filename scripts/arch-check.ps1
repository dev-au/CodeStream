$FAIL = 0

function Check-Layer($Layer, $Dir, $Packages) {
  foreach ($pkg in $Packages) {
    $results = Get-ChildItem -Recurse -Filter "*.go" $Dir -ErrorAction SilentlyContinue |
      Select-String -Pattern "`"$pkg"
    if ($results) {
      Write-Host "VIOLATION [$Layer]: '$pkg' found in ${Dir}/"
      $results | ForEach-Object { Write-Host $_.ToString() }
      $script:FAIL = 1
    }
  }
}

Check-Layer "domain" "internal/domain" @(
  "github.com/dev-au/CodeStream/internal/usecase",
  "github.com/dev-au/CodeStream/internal/adapter",
  "github.com/dev-au/CodeStream/internal/infrastructure"
)

Check-Layer "usecase" "internal/usecase" @(
  "github.com/dev-au/CodeStream/internal/adapter",
  "github.com/dev-au/CodeStream/internal/infrastructure"
)

Check-Layer "adapter" "internal/adapter" @(
  "github.com/dev-au/CodeStream/internal/infrastructure"
)

if ($FAIL -eq 0) { Write-Host "Architecture check passed." }
exit $FAIL