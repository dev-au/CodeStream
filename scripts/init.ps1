param (
    [Parameter(Mandatory=$true)]
    [string]$NewName
)

$scriptPath = Split-Path -Parent $MyInvocation.MyCommand.Definition
Set-Location "$scriptPath\.."

Write-Host "🚀 Initializing project: $NewName" -ForegroundColor Cyan

$Target = "gotemplate"


$files = Get-ChildItem -Recurse -File | Where-Object { 
    $_.FullName -notlike "*\.git\*" -and $_.FullName -notlike "*\scripts\*" 
}

foreach ($file in $files) {
    $content = Get-Content $file.FullName -Raw
    if ($content -match [regex]::Escape($Target)) {
        $content -replace [regex]::Escape($Target), $NewName | Set-Content $file.FullName -NoNewline
        Write-Host "  Updated: $($file.Name)" -ForegroundColor Gray
    }
}

if (Test-Path "go.mod") {
    Write-Host "📦 Tidying Go modules..." -ForegroundColor DarkCyan
    go mod tidy
}


if (Test-Path ".git") {
    Write-Host "⚠️ Existing .git directory found. Skipping 'git init' to protect your history." -ForegroundColor Yellow
    Write-Host "   If you want a fresh start, run 'rm -rf .git; git init' manually." -ForegroundColor Yellow
} else {
    Write-Host "🌱 No .git found. Initializing new repository..." -ForegroundColor Green
    git init | Out-Null
    Write-Host "✅ Git initialized!" -ForegroundColor Green

    Write-Host "🪝 Setting up git hooks directly..." -ForegroundColor Blue
    
    $HookFile = ".git/hooks/pre-commit"
    
    $HookContent = @"
#!/bin/bash
echo "Running pre-commit checks..."
make precommit
if [ `$? -ne 0 ]; then
    echo "❌ Pre-commit checks failed. Commit aborted."
    exit 1
fi
"@

    Set-Content -Path $HookFile -Value $HookContent -Encoding Ascii
    
    Write-Host "✅ Hook installed successfully at $HookFile" -ForegroundColor Green
}

Write-Host "✅ Initialization complete!" -ForegroundColor Green