#!/bin/sh
# ktun 설치 스크립트 (Linux, macOS)
#   curl -fsSL https://github.com/kintoun-secops/kintoun-ssm-tunnel/releases/latest/download/install.sh | sh
# VERSION 으로 버전을, INSTALL_DIR 로 설치 위치를 바꿀 수 있다.
set -eu

REPO="kintoun-secops/kintoun-ssm-tunnel"
BIN="ktun"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

fail() {
  echo "오류: $*" >&2
  exit 1
}

case "$(uname -s)" in
  Linux) os="linux" ;;
  Darwin) os="darwin" ;;
  *) fail "지원하지 않는 운영체제입니다. Windows 는 install.ps1 을 사용하세요." ;;
esac

case "$(uname -m)" in
  x86_64 | amd64) arch="amd64" ;;
  aarch64 | arm64) arch="arm64" ;;
  *) fail "지원하지 않는 아키텍처입니다: $(uname -m)" ;;
esac

version="${VERSION:-}"
if [ -z "$version" ]; then
  latest=$(curl -fsSIL -o /dev/null -w '%{url_effective}' "https://github.com/$REPO/releases/latest") ||
    fail "최신 릴리스를 확인하지 못했습니다."
  version="${latest##*/}"
fi

name="${BIN}_${version}_${os}_${arch}"
base="https://github.com/$REPO/releases/download/$version"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo "$version ($os/$arch) 을 내려받는 중..."
curl -fsSL -o "$tmp/$name.tar.gz" "$base/$name.tar.gz" || fail "$name.tar.gz 를 내려받지 못했습니다."
curl -fsSL -o "$tmp/checksums.txt" "$base/checksums.txt" || fail "checksums.txt 를 내려받지 못했습니다."

expected=$(grep " $name.tar.gz\$" "$tmp/checksums.txt" | cut -d' ' -f1)
[ -n "$expected" ] || fail "checksums.txt 에 $name.tar.gz 항목이 없습니다."

if command -v sha256sum >/dev/null 2>&1; then
  actual=$(sha256sum "$tmp/$name.tar.gz" | cut -d' ' -f1)
elif command -v shasum >/dev/null 2>&1; then
  actual=$(shasum -a 256 "$tmp/$name.tar.gz" | cut -d' ' -f1)
else
  fail "sha256sum 또는 shasum 이 필요합니다."
fi
[ "$expected" = "$actual" ] || fail "체크섬이 일치하지 않습니다."

tar -xzf "$tmp/$name.tar.gz" -C "$tmp"
mkdir -p "$INSTALL_DIR"
install -m 755 "$tmp/$name/$BIN" "$INSTALL_DIR/$BIN"

echo "$INSTALL_DIR/$BIN 에 설치했습니다."
case ":$PATH:" in
  *":$INSTALL_DIR:"*)
    echo "실행: $BIN"
    ;;
  *)
    case "$(basename "${SHELL:-}")" in
      zsh) rc="$HOME/.zshrc" ;;
      bash)
        rc="$HOME/.bashrc"
        [ "$os" = "darwin" ] && rc="$HOME/.bash_profile"
        ;;
      *) rc="" ;;
    esac

    echo
    echo "$INSTALL_DIR 가 PATH 에 없어 아직 $BIN 명령으로 실행할 수 없습니다."
    if [ -n "$rc" ]; then
      echo "아래 명령을 한 번 실행한 뒤 터미널을 새로 여세요."
      echo
      echo "  echo 'export PATH=\"$INSTALL_DIR:\$PATH\"' >> $rc"
    else
      echo "사용하는 셸의 설정 파일에 아래 줄을 추가한 뒤 터미널을 새로 여세요."
      echo
      echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
    fi
    echo
    echo "지금 바로 실행하려면: $INSTALL_DIR/$BIN"
    ;;
esac
