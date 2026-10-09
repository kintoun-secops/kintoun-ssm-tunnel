#!/bin/sh
# 다음 vX.Y.Z 태그를 계산해 출력한다. 사용법: next-tag.sh [auto|patch|minor|major]
#
# auto 는 마지막 태그 이후 커밋 메시지에서 가장 높은 단계를 고른다.
#   BREAKING CHANGE 또는 type! : major (major 가 0 이면 minor)
#   feat                       : minor
#   그 밖의 모든 커밋          : patch
set -eu

mode="${1:-auto}"
latest=$(git tag --list 'v[0-9]*.[0-9]*.[0-9]*' --sort=-v:refname | head -n 1)

if [ -z "$latest" ]; then
  echo "v0.1.0"
  exit 0
fi

IFS=. read -r major minor patch <<EOF
${latest#v}
EOF

bump="$mode"
if [ "$mode" = "auto" ]; then
  bump="patch"
  for commit in $(git rev-list "$latest"..HEAD); do
    subject=$(git log -1 --format=%s "$commit")
    if printf '%s\n' "$subject" | grep -Eq '^[a-z]+(\([^)]*\))?!:' ||
      git log -1 --format=%b "$commit" | grep -Eq '^BREAKING[ -]CHANGE:'; then
      bump="major"
      break
    fi
    if printf '%s\n' "$subject" | grep -Eq '^feat(\([^)]*\))?:'; then
      bump="minor"
    fi
  done
  if [ "$bump" = "major" ] && [ "$major" -eq 0 ]; then
    bump="minor"
  fi
  echo "bump=$bump ($latest 이후 커밋 기준)" >&2
fi

case "$bump" in
  major) major=$((major + 1)); minor=0; patch=0 ;;
  minor) minor=$((minor + 1)); patch=0 ;;
  patch) patch=$((patch + 1)) ;;
  *) echo "알 수 없는 bump: $bump" >&2; exit 1 ;;
esac

echo "v${major}.${minor}.${patch}"
