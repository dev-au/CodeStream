$unformatted = & "$env:CI_PROJECT_DIR\tools\goimports.exe" -local github.com/dev-au/CodeStream -l .
if ($unformatted) {
    Write-Host "These files need formatting:"
    $unformatted | ForEach-Object { Write-Host $_ }
    exit 1
}
Write-Host "All files properly formatted."