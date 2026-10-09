$ErrorActionPreference = "Stop"

$cacheDir = Join-Path $PSScriptRoot ".gocache"
New-Item -ItemType Directory -Path $cacheDir -Force | Out-Null
$env:GOCACHE = $cacheDir

$tasks = @(
    "task01-hello",
    "task02-flow",
    "task03-collection",
    "task04-struct",
    "task05-http",
    "task06-concurrency",
    "task07-gin",
    "task08-gorm",
    "task09-viper"
)

foreach ($task in $tasks) {
    Write-Host "== $task =="
    Push-Location (Join-Path $PSScriptRoot $task)
    try {
        go test .
        if ($LASTEXITCODE -ne 0) {
            exit $LASTEXITCODE
        }
    }
    finally {
        Pop-Location
    }
}
