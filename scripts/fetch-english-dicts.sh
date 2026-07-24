#!/usr/bin/env bash
# Fetch LanguageTool english-pos-dict (binary morfologik dicts not in the LT git tree).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
VERSION="${1:-0.6}"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT
URL="https://repo1.maven.org/maven2/org/languagetool/english-pos-dict/${VERSION}/english-pos-dict-${VERSION}.jar"
echo "Downloading $URL"
curl -fsSL -o "$TMP/dict.jar" "$URL"
DEST_TP="$ROOT/third_party/english-pos-dict"
mkdir -p "$DEST_TP"
unzip -o "$TMP/dict.jar" -d "$TMP/out"
rm -rf "$DEST_TP/org"
cp -a "$TMP/out/org" "$DEST_TP/"
# optional: also place into LT submodule layout for tools that expect it
DEST_MOD="$ROOT/inspiration/languagetool/languagetool-language-modules/en/src/main/resources/org/languagetool/resource/en"
if [[ -d "$(dirname "$DEST_MOD")" ]]; then
  mkdir -p "$DEST_MOD/hunspell"
  cp -a "$TMP/out/org/languagetool/resource/en/." "$DEST_MOD/"
  echo "Also copied into submodule resource tree (local only, not for commit)."
fi
echo "Installed English dicts into $DEST_TP"
