param(
  [string]$Image = "omnitoken-openai-rust-bench:local"
)

$ErrorActionPreference = "Stop"
$Root = Resolve-Path (Join-Path $PSScriptRoot "..\..")
$Results = Join-Path $Root "benchmarks\results"
New-Item -ItemType Directory -Force -Path $Results | Out-Null

$previousErrorActionPreference = $ErrorActionPreference
try {
  $ErrorActionPreference = "Continue"
  docker build -f (Join-Path $Root "benchmarks\docker\Dockerfile.openai-rust") -t $Image $Root 2>&1 | ForEach-Object { $_.ToString() }
}
finally {
  $ErrorActionPreference = $previousErrorActionPreference
}
if ($LASTEXITCODE -ne 0) { throw "rust benchmark image build failed" }
$raw = Join-Path $Results "rust.csv"
$lines = New-Object System.Collections.Generic.List[string]
$previousErrorActionPreference = $ErrorActionPreference
try {
  $ErrorActionPreference = "Continue"
  docker run --rm $Image 2>&1 | ForEach-Object {
    $line = $_.ToString()
    $lines.Add($line)
    $line
  }
}
finally {
  $ErrorActionPreference = $previousErrorActionPreference
}
if ($LASTEXITCODE -ne 0) { throw "rust benchmark run failed" }
[System.IO.File]::WriteAllLines($raw, $lines, [System.Text.UTF8Encoding]::new($false))
