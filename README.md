# kintoun-ssm-tunnel

`SSMPortForward=true` 태그가 붙은 EC2 에 SSM 포트포워딩 터널을 여는 TUI 입니다.
프로필 선택, `aws login`, 인스턴스 조회, 터널 열기를 한 흐름으로 처리합니다.

## 필요한 것

- AWS CLI v2
- [Session Manager 플러그인](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html)

## 설치

[Releases](https://github.com/kintoun-secops/kintoun-ssm-tunnel/releases) 에서 운영체제에 맞는 압축 파일을 내려받아 풀고 실행합니다.
같은 릴리스의 `checksums.txt` 로 파일 해시를 확인할 수 있습니다.
저장소가 비공개이므로 GitHub CLI 로 내려받습니다.

```shell
gh release download --repo kintoun-secops/kintoun-ssm-tunnel --pattern "*linux_amd64.tar.gz"
```

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
