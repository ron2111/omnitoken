param(
  [int]$Count = 3,
  [string]$Benchtime = "1s",
  [switch]$Rust
)

$ErrorActionPreference = "Stop"
$Root = Resolve-Path (Join-Path $PSScriptRoot "..\..")
$Results = Join-Path $Root "benchmarks\results"
$Reports = Join-Path $Root "benchmarks\reports"
New-Item -ItemType Directory -Force -Path $Results | Out-Null
New-Item -ItemType Directory -Force -Path $Reports | Out-Null

& (Join-Path $Root "benchmarks\scripts\run-go.ps1") -Count $Count -Benchtime $Benchtime

$inputs = @((Join-Path $Results "go.jsonl"))
if ($Rust) {
  & (Join-Path $Root "benchmarks\scripts\run-rust.ps1")
  $inputs += (Join-Path $Results "rust.csv")
}

python (Join-Path $Root "benchmarks\scripts\combine.py") --inputs ($inputs -join ",") --output (Join-Path $Results "combined.csv")
if ($LASTEXITCODE -ne 0) { throw "benchmark combine failed" }
python (Join-Path $Root "benchmarks\scripts\plot.py") --input (Join-Path $Results "combined.csv") --out-dir $Reports
if ($LASTEXITCODE -ne 0) { throw "benchmark plot failed" }

"Benchmark report: $Reports"
