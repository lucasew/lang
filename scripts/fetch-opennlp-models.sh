#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DEST="$ROOT/third_party/opennlp-models"
mkdir -p "$DEST"
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
for art in opennlp-chunk-models opennlp-postag-models opennlp-tokenize-models; do
  url="https://repo1.maven.org/maven2/edu/washington/cs/knowitall/${art}/1.5/${art}-1.5.jar"
  echo "Downloading $url"
  curl -fsSL -o "$TMP/$art.jar" "$url"
  unzip -jo "$TMP/$art.jar" '*.bin' -d "$DEST"
done
ls -la "$DEST"
