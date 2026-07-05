param(
  [int]$Count = 3,
  [string]$Benchtime = "1s",
  [string]$Timeout = "60m",
  [string]$OutputPrefix = "go-docker",
  [string]$Image = "omnitoken-go-bench:local"
)

$ErrorActionPreference = "Stop"
$Root = Resolve-Path (Join-Path $PSScriptRoot "..\..")
$Results = Join-Path $Root "benchmarks\results"
New-Item -ItemType Directory -Force -Path $Results | Out-Null

docker build -f (Join-Path $Root "benchmarks\docker\Dockerfile.go-bench") -t $Image $Root
if ($LASTEXITCODE -ne 0) { throw "go benchmark image build failed" }

$raw = Join-Path $Results "$OutputPrefix.raw.txt"
$lines = New-Object System.Collections.Generic.List[string]
$previousErrorActionPreference = $ErrorActionPreference
try {
  $ErrorActionPreference = "Continue"
  docker run --rm -e "BENCH_COUNT=$Count" -e "BENCH_TIME=$Benchtime" -e "BENCH_TIMEOUT=$Timeout" $Image 2>&1 | ForEach-Object {
    $line = $_.ToString()
    $lines.Add($line)
    $line
  }
}
finally {
  $ErrorActionPreference = $previousErrorActionPreference
}
if ($LASTEXITCODE -ne 0) { throw "go benchmark docker run failed" }
[System.IO.File]::WriteAllLines($raw, $lines, [System.Text.UTF8Encoding]::new($false))

python (Join-Path $Root "benchmarks\scripts\parse_go.py") --input $raw --output (Join-Path $Results "$OutputPrefix.jsonl") --runner-suffix "_docker"
if ($LASTEXITCODE -ne 0) { throw "go docker benchmark parse failed" }
