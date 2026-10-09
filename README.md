# kintoun-ssm-tunnel

`SSMPortForward=true` 태그가 붙은 EC2 에 SSM 포트포워딩 터널을 여는 TUI 입니다.
프로필 선택, `aws login`, 인스턴스 조회, 터널 열기를 한 흐름으로 처리합니다.

## 필요한 것

- AWS CLI v2
- [Session Manager 플러그인](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html)

둘 중 하나라도 없으면 실행 시 운영체제에 맞는 설치 링크를 출력하고 종료합니다.

| 운영체제 | AWS CLI v2 | Session Manager 플러그인 |
| --- | --- | --- |
| Windows | [설치 안내](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html) | [설치 안내](https://docs.aws.amazon.com/systems-manager/latest/userguide/install-plugin-windows.html) |
| macOS | [설치 안내](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html) | [설치 안내](https://docs.aws.amazon.com/systems-manager/latest/userguide/install-plugin-macos-overview.html) |
| Linux | [설치 안내](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html) | [설치 안내](https://docs.aws.amazon.com/systems-manager/latest/userguide/install-plugin-linux-overview.html) |

## 설치

최신 릴리스를 내려받아 체크섬을 확인하고 설치합니다.

Linux, macOS 는 `~/.local/bin` 에 설치합니다.

```shell
curl -fsSL https://github.com/kintoun-secops/kintoun-ssm-tunnel/releases/latest/download/install.sh | sh
```

Windows 는 `%LOCALAPPDATA%\kintoun-ssm-tunnel` 에 설치하고 사용자 PATH 에 추가합니다.

```powershell
irm https://github.com/kintoun-secops/kintoun-ssm-tunnel/releases/latest/download/install.ps1 | iex
```

`VERSION` 환경 변수로 버전을, `INSTALL_DIR` 로 설치 위치를 바꿀 수 있습니다.
직접 내려받으려면 [Releases](https://github.com/kintoun-secops/kintoun-ssm-tunnel/releases) 에서 운영체제에 맞는 압축 파일을 받고, 같은 릴리스의 `checksums.txt` 로 해시를 확인합니다.

## 실행

```shell
go run .
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
