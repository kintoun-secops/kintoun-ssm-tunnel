# kintoun-ssm-tunnel

`SSMPortForward=true` 태그가 붙은 EC2 에 SSM 포트포워딩 터널을 여는 TUI 입니다.
프로필 선택, `aws login`, 인스턴스 조회, 터널 열기를 한 흐름으로 처리합니다.
실행 명령은 `ktun` 입니다.

## 필요한 것

- AWS CLI v2
- [Session Manager 플러그인](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html)

둘 중 하나라도 없으면 실행 시 운영체제에 맞는 설치 링크를 출력하고 종료합니다.

| 운영체제 | AWS CLI v2 | Session Manager 플러그인 |
| --- | --- | --- |
| Windows | [설치 안내](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html) | [설치 안내](https://docs.aws.amazon.com/systems-manager/latest/userguide/install-plugin-windows.html) |
| macOS | [설치 안내](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html) | [설치 안내](https://docs.aws.amazon.com/systems-manager/latest/userguide/install-plugin-macos-overview.html) |
| Linux | [설치 안내](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html) | [설치 안내](https://docs.aws.amazon.com/systems-manager/latest/userguide/install-plugin-linux-overview.html) |

## 설치, 업데이트, 삭제

패키지 매니저를 쓰는 방법을 권장합니다. 업데이트와 삭제를 패키지 매니저가 처리합니다.

### macOS: Homebrew

```shell
# 설치
brew tap kintoun-secops/ssm https://github.com/kintoun-secops/kintoun-ssm-tunnel
brew trust --cask kintoun-secops/ssm/ktun
brew install --cask kintoun-secops/ssm/ktun

# 업데이트
brew update
brew upgrade --cask kintoun-secops/ssm/ktun

# 삭제
brew uninstall --cask kintoun-secops/ssm/ktun
brew untrust --cask kintoun-secops/ssm/ktun
brew untap kintoun-secops/ssm
```

Homebrew 는 공식 tap 이 아닌 cask 를 `brew trust` 로 신뢰하기 전에는 불러오지 않습니다. 신뢰 목록은 `~/.homebrew/trust.json` 에 저장됩니다.
항상 `kintoun-secops/ssm/` 을 붙인 전체 이름을 사용합니다.

macOS 에서 서명되지 않은 바이너리라는 이유로 실행이 막히면 격리 속성을 제거합니다.

```shell
xattr -d com.apple.quarantine "$(which ktun)"
```

### Windows: Scoop

```powershell
# 설치
scoop bucket add kintoun https://github.com/kintoun-secops/kintoun-ssm-tunnel
scoop install ktun

# 업데이트
scoop update
scoop update ktun

# 삭제
scoop uninstall ktun
scoop bucket rm kintoun
```

### 설치 스크립트: Linux 및 패키지 매니저를 쓰지 않는 경우

최신 릴리스를 내려받아 체크섬을 확인하고 설치합니다. `VERSION` 환경 변수로 버전을, `INSTALL_DIR` 로 설치 위치를 바꿀 수 있습니다.

Linux, macOS 는 `~/.local/bin` 에 설치합니다. PATH 에 없으면 사용하는 셸에 맞는 `export PATH` 추가 명령을 출력하며, 셸 설정 파일은 직접 수정하지 않습니다.

```shell
# 설치와 업데이트: 같은 명령을 다시 실행하면 최신 버전으로 바뀝니다
curl -fsSL https://github.com/kintoun-secops/kintoun-ssm-tunnel/releases/latest/download/install.sh | sh

# 삭제: PATH 에 추가한 export 줄이 있으면 셸 설정 파일에서 직접 지웁니다
rm ~/.local/bin/ktun
```

Windows 는 `%LOCALAPPDATA%\ktun` 에 설치하고 사용자 PATH 에 추가합니다.

```powershell
# 설치와 업데이트: 같은 명령을 다시 실행하면 최신 버전으로 바뀝니다
irm https://github.com/kintoun-secops/kintoun-ssm-tunnel/releases/latest/download/install.ps1 | iex

# 삭제: 설치 폴더와 사용자 PATH 항목을 함께 지웁니다
$dir = "$env:LOCALAPPDATA\ktun"
Remove-Item -Recurse -Force $dir
$paths = [Environment]::GetEnvironmentVariable('Path', 'User') -split ';' | Where-Object { $_ -and $_ -ne $dir }
[Environment]::SetEnvironmentVariable('Path', ($paths -join ';'), 'User')
```

### 내려받은 파일 검증

[Releases](https://github.com/kintoun-secops/kintoun-ssm-tunnel/releases) 에서 직접 내려받은 파일은 같은 릴리스의 `checksums.txt` 로 해시를 확인할 수 있습니다.
릴리스 파일은 GitHub Actions 에서 빌드되며 빌드 출처 증명이 함께 발급됩니다. 아래 명령으로 이 저장소의 워크플로우에서 만든 파일인지 확인합니다.

```shell
gh attestation verify <내려받은 파일> --repo kintoun-secops/kintoun-ssm-tunnel
```

## 실행

```shell
ktun
```

1. 프로필을 고릅니다. 목록에 없으면 직접 입력합니다. `default` 프로필은 목록에 나오지 않습니다.
2. 자격증명이 없거나 만료되었으면 `aws login --profile <프로필>` 을 실행합니다.
3. 접속할 인스턴스를 고르면 터널이 열리고 접속 주소가 표시됩니다.
4. `Ctrl+C` 로 터널을 닫으면 인스턴스 목록으로 돌아옵니다.

## 기본 포트

인스턴스 `Name` 태그에 아래 문자열이 들어 있으면 포트를 묻지 않습니다.

| 이름에 포함 | 원격 포트 | 로컬 포트 | 접속 주소 |
| --- | --- | --- | --- |
| `kali` | 6080 | 6080 | `http://localhost:6080/vnc.html` |
| `wazuh` | 443 | 56789 | `https://localhost:56789` |
| `velociraptor` | 8889 | 8889 | `https://localhost:8889` |

그 외 인스턴스는 포트를 입력받아 원격과 로컬에 같은 값을 사용합니다.

## 개발

```shell
go run .
go test ./...
```

## 릴리스

`main` 에 Go 코드(`*.go`, `go.mod`, `go.sum`), `.goreleaser.yaml`, `install.sh`, `install.ps1` 변경이 push되면 `Tag` 워크플로우가 다음 버전 태그(`vX.Y.Z`)를 만들고 `Release` 워크플로우를 실행합니다.
README 나 워크플로우 파일만 바뀐 경우에는 릴리스가 만들어지지 않습니다.

### 버전 올리는 기준

마지막 태그 이후의 커밋 메시지를 모두 보고 가장 높은 단계를 적용합니다.

| 커밋 | 올라가는 자리 | 예 |
| --- | --- | --- |
| `feat!:` 처럼 type 뒤에 `!` 가 있거나 본문에 `BREAKING CHANGE:` 가 있음 | major (major 가 0 인 동안은 minor) | 명령 이름 변경, 설정 형식 변경 |
| `feat:` | minor | 새 기능 |
| 그 밖의 모든 커밋 (`fix:`, `refactor:`, `perf:`, `docs:`, `ci:`, `chore:` 등) | patch | 버그 수정, 내부 정리 |

- 커밋 메시지는 `type(scope): 설명` 형식을 지킵니다. 형식이 다르면 patch 로 계산됩니다.
- 기능 추가나 호환성이 깨지는 변경은 `feat:` 나 `feat!:` 로 써야 minor 이상으로 올라갑니다.
- 기준과 다르게 올리려면 Actions 의 `Tag` 워크플로우를 수동으로 실행하고 `bump` 에 `auto`, `patch`, `minor`, `major` 중 하나를 고릅니다.
- 계산 규칙은 `.github/scripts/next-tag.sh` 에 있습니다.

### 동작
- `v*.*.*` 태그를 직접 push해도 `Release` 워크플로우가 실행됩니다.
- `Release` 는 GoReleaser 로 빌드와 GitHub 릴리스를 만들고, 같은 실행에서 이 저장소의 `Casks/` 와 `bucket/` 에 Homebrew cask 와 Scoop 매니페스트를 커밋합니다.
- 워크플로우가 만든 태그는 `push` 이벤트를 발생시키지 않으므로, `Tag` 가 `workflow_dispatch` 로 `Release` 를 직접 호출합니다.
