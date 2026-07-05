param(
  [int]$Count = 3,
  [string]$Benchtime = "1s",
  [string]$Timeout = "60m",
  [switch]$DockerGo,
  [switch]$Rust,
  [string]$Stage = "all",
  [switch]$Resume,
  [switch]$Force
)

$ErrorActionPreference = "Stop"
$Root = Resolve-Path (Join-Path $PSScriptRoot "..\..")
$Results = Join-Path $Root "benchmarks\results"
$Reports = Join-Path $Root "benchmarks\reports"
$Checkpoints = Join-Path $Results "checkpoints"
$Logs = Join-Path $Results "logs"
New-Item -ItemType Directory -Force -Path $Results,$Reports,$Checkpoints,$Logs | Out-Null

function Write-Metadata {
  $Metadata = Join-Path $Results "metadata.json"
  $cpu = Get-CimInstance Win32_Processor | Select-Object -First 1
  $computer = Get-CimInstance Win32_ComputerSystem
  $os = Get-CimInstance Win32_OperatingSystem
  $cpuLoad = $null
  try { $cpuLoad = (Get-Counter '\Processor(_Total)\% Processor Time' -SampleInterval 1 -MaxSamples 1).CounterSamples.CookedValue } catch { $cpuLoad = $null }
  $metadataObject = [ordered]@{
    timestamp_utc = (Get-Date).ToUniversalTime().ToString("o")
    git_commit = (git rev-parse HEAD)
    git_status = @(git status --short)
    go_version = (go version)
    goos = (go env GOOS)
    goarch = (go env GOARCH)
    docker_version = (docker --version)
    os = [ordered]@{ caption = $os.Caption; version = $os.Version }
    cpu = [ordered]@{ name = $cpu.Name; cores = $cpu.NumberOfCores; logical_processors = $cpu.NumberOfLogicalProcessors; max_clock_mhz = $cpu.MaxClockSpeed; load_percent_sample = $cpuLoad }
    memory = [ordered]@{ total_physical_bytes = [int64]$computer.TotalPhysicalMemory; free_physical_kb = [int64]$os.FreePhysicalMemory }
    benchmark = [ordered]@{ count = $Count; benchtime = $Benchtime; timeout = $Timeout; docker_go = [bool]$DockerGo; rust = [bool]$Rust; note = "Checkpointed benchmark run. Background processes and thermal state may affect results." }
  }
  $metadataObject | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath $Metadata -Encoding UTF8
}

function Invoke-BenchStage {
  param(
    [string]$Name,
    [scriptblock]$Body
  )
  $checkpoint = Join-Path $Checkpoints "$Name.done"
  $log = Join-Path $Logs "$Name.log"
  if ($Resume -and -not $Force -and (Test-Path $checkpoint)) {
    "Skipping completed stage: $Name"
    return
  }
  Remove-Item -LiteralPath $checkpoint -Force -ErrorAction SilentlyContinue
  "Starting stage: $Name"
  try {
    & $Body *>&1 | Tee-Object -FilePath $log
    if ($LASTEXITCODE -ne 0) { throw "stage $Name failed with exit code $LASTEXITCODE" }
    [System.IO.File]::WriteAllText($checkpoint, (Get-Date).ToUniversalTime().ToString("o"), [System.Text.UTF8Encoding]::new($false))
    "Completed stage: $Name"
  } catch {
    "Failed stage: $Name" | Tee-Object -FilePath $log -Append
    $_ | Out-String | Tee-Object -FilePath $log -Append
    throw
  }
}

function Should-RunStage([string]$Name) {
  return $Stage -eq "all" -or $Stage -eq $Name
}

if (Should-RunStage "metadata") {
  Invoke-BenchStage "00-metadata" { Write-Metadata }
}

if (Should-RunStage "native-go") {
  Invoke-BenchStage "01-native-go" { & (Join-Path $Root "benchmarks\scripts\run-go.ps1") -Count $Count -Benchtime $Benchtime -Timeout $Timeout -OutputPrefix "go" }
}

if ($DockerGo -and (Should-RunStage "docker-go")) {
  Invoke-BenchStage "02-docker-go" { & (Join-Path $Root "benchmarks\scripts\run-docker-go.ps1") -Count $Count -Benchtime $Benchtime -Timeout $Timeout -OutputPrefix "go-docker" }
}

if ($Rust -and (Should-RunStage "rust")) {
  Invoke-BenchStage "03-rust" { & (Join-Path $Root "benchmarks\scripts\run-rust.ps1") }
}

if (Should-RunStage "combine") {
  Invoke-BenchStage "04-combine" {
    $inputs = @()
    $native = Join-Path $Results "go.jsonl"
    if (Test-Path $native) { $inputs += $native }
    $docker = Join-Path $Results "go-docker.jsonl"
    if ($DockerGo -and (Test-Path $docker)) { $inputs += $docker }
    $rustCsv = Join-Path $Results "rust.csv"
    if ($Rust -and (Test-Path $rustCsv)) { $inputs += $rustCsv }
    if ($inputs.Count -eq 0) { throw "no benchmark inputs found" }
    python (Join-Path $Root "benchmarks\scripts\combine.py") --inputs ($inputs -join ",") --output (Join-Path $Results "combined.csv")
  }
}

if (Should-RunStage "report") {
  Invoke-BenchStage "05-report" {
    python (Join-Path $Root "benchmarks\scripts\plot.py") --input (Join-Path $Results "combined.csv") --metadata (Join-Path $Results "metadata.json") --out-dir $Reports
  }
}

"Benchmark report: $Reports"
