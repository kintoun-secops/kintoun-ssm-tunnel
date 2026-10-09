# kintoun-ssm-tunnel 설치 스크립트 (Windows)
#   irm https://github.com/kintoun-secops/kintoun-ssm-tunnel/releases/latest/download/install.ps1 | iex
# VERSION 으로 버전을, INSTALL_DIR 로 설치 위치를 바꿀 수 있다.
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

$repo = 'kintoun-secops/kintoun-ssm-tunnel'
$bin = 'kintoun-ssm-tunnel'
$installDir = if ($env:INSTALL_DIR) { $env:INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA $bin }

$version = $env:VERSION
if (-not $version) {
    $version = (Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest").tag_name
}

$name = "${bin}_${version}_windows_amd64"
$base = "https://github.com/$repo/releases/download/$version"
$tmp = Join-Path ([IO.Path]::GetTempPath()) ([Guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $tmp | Out-Null

try {
    Write-Host "$version (windows/amd64) 을 내려받는 중..."
    Invoke-WebRequest "$base/$name.zip" -OutFile (Join-Path $tmp "$name.zip")
    Invoke-WebRequest "$base/checksums.txt" -OutFile (Join-Path $tmp 'checksums.txt')

    $line = Get-Content (Join-Path $tmp 'checksums.txt') | Where-Object { $_ -match "\s$([regex]::Escape("$name.zip"))$" }
    if (-not $line) { throw "checksums.txt 에 $name.zip 항목이 없습니다." }
    $expected = ($line -split '\s+')[0]
    $actual = (Get-FileHash (Join-Path $tmp "$name.zip") -Algorithm SHA256).Hash
    if ($expected -ne $actual) { throw '체크섬이 일치하지 않습니다.' }

    Expand-Archive (Join-Path $tmp "$name.zip") -DestinationPath $tmp
    New-Item -ItemType Directory -Force $installDir | Out-Null
    Copy-Item (Join-Path $tmp "$name\$bin.exe") $installDir -Force
}
finally {
    Remove-Item $tmp -Recurse -Force -ErrorAction SilentlyContinue
}

Write-Host "$installDir\$bin.exe 에 설치했습니다."
$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if (($userPath -split ';') -notcontains $installDir) {
    [Environment]::SetEnvironmentVariable('Path', "$userPath;$installDir", 'User')
    Write-Host 'PATH 에 추가했습니다. 터미널을 새로 연 뒤 실행하세요.'
}
Write-Host "실행: $bin"
