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

$Metadata = Join-Path $Results "metadata.json"
$cpu = Get-CimInstance Win32_Processor | Select-Object -First 1
$computer = Get-CimInstance Win32_ComputerSystem
$os = Get-CimInstance Win32_OperatingSystem
$cpuLoad = $null
try {
  $cpuLoad = (Get-Counter '\Processor(_Total)\% Processor Time' -SampleInterval 1 -MaxSamples 1).CounterSamples.CookedValue
} catch {
  $cpuLoad = $null
}
$gitStatus = git status --short
$metadataObject = [ordered]@{
  timestamp_utc = (Get-Date).ToUniversalTime().ToString("o")
  git_commit = (git rev-parse HEAD)
  git_status = @($gitStatus)
  go_version = (go version)
  goos = (go env GOOS)
  goarch = (go env GOARCH)
  docker_version = (docker --version)
  os = [ordered]@{
    caption = $os.Caption
    version = $os.Version
  }
  cpu = [ordered]@{
    name = $cpu.Name
    cores = $cpu.NumberOfCores
    logical_processors = $cpu.NumberOfLogicalProcessors
    max_clock_mhz = $cpu.MaxClockSpeed
    load_percent_sample = $cpuLoad
  }
  memory = [ordered]@{
    total_physical_bytes = [int64]$computer.TotalPhysicalMemory
    free_physical_kb = [int64]$os.FreePhysicalMemory
  }
  benchmark = [ordered]@{
    count = $Count
    benchtime = $Benchtime
    rust = [bool]$Rust
    note = "Single local run. Background processes and thermal state may affect results. Use higher Count/Benchtime for publishable numbers."
  }
}
$metadataObject | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath $Metadata -Encoding UTF8

& (Join-Path $Root "benchmarks\scripts\run-go.ps1") -Count $Count -Benchtime $Benchtime

$inputs = @((Join-Path $Results "go.jsonl"))
if ($Rust) {
  & (Join-Path $Root "benchmarks\scripts\run-rust.ps1")
  $inputs += (Join-Path $Results "rust.csv")
}

python (Join-Path $Root "benchmarks\scripts\combine.py") --inputs ($inputs -join ",") --output (Join-Path $Results "combined.csv")
if ($LASTEXITCODE -ne 0) { throw "benchmark combine failed" }
python (Join-Path $Root "benchmarks\scripts\plot.py") --input (Join-Path $Results "combined.csv") --metadata $Metadata --out-dir $Reports
if ($LASTEXITCODE -ne 0) { throw "benchmark plot failed" }

"Benchmark report: $Reports"
