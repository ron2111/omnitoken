param(
  [int]$Count = 3,
  [string]$Benchtime = "1s"
)

$ErrorActionPreference = "Stop"
$Root = Resolve-Path (Join-Path $PSScriptRoot "..\..")
$Results = Join-Path $Root "benchmarks\results"
New-Item -ItemType Directory -Force -Path $Results | Out-Null

Push-Location (Join-Path $Root "tools\benchmark_harness")
try {
  $env:GOWORK = Join-Path $Root "go.work"
  $raw = Join-Path $Results "go.raw.txt"
  $lines = New-Object System.Collections.Generic.List[string]
  go test -run "^$" -bench "BenchmarkTokenizer" -benchmem "-count=$Count" "-benchtime=$Benchtime" 2>&1 | ForEach-Object {
    $line = $_.ToString()
    $lines.Add($line)
    $line
  }
  if ($LASTEXITCODE -ne 0) { throw "go benchmark failed" }
  [System.IO.File]::WriteAllLines($raw, $lines, [System.Text.UTF8Encoding]::new($false))
}
finally {
  Pop-Location
}

python (Join-Path $Root "benchmarks\scripts\parse_go.py") --input (Join-Path $Results "go.raw.txt") --output (Join-Path $Results "go.jsonl")
if ($LASTEXITCODE -ne 0) { throw "go benchmark parse failed" }
