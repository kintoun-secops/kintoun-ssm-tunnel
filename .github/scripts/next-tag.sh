#!/bin/sh
# 가장 높은 vX.Y.Z 태그에서 다음 버전을 계산해 출력한다. 사용법: next-tag.sh [patch|minor|major]
set -eu

bump="${1:-patch}"
latest=$(git tag --list 'v[0-9]*.[0-9]*.[0-9]*' --sort=-v:refname | head -n 1)

if [ -z "$latest" ]; then
  echo "v0.1.0"
  exit 0
fi

IFS=. read -r major minor patch <<EOF
${latest#v}
EOF

case "$bump" in
  major) major=$((major + 1)); minor=0; patch=0 ;;
  minor) minor=$((minor + 1)); patch=0 ;;
  patch) patch=$((patch + 1)) ;;
  *) echo "알 수 없는 bump: $bump" >&2; exit 1 ;;
esac

echo "v${major}.${minor}.${patch}"
