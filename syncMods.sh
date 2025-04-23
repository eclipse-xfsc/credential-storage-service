#!/usr/bin/env bash

set -euo pipefail

SOURCE_MOD="$1"
TARGET_MOD="$2"

echo "🔍 Vergleiche $SOURCE_MOD ↔ $TARGET_MOD"

cd "$(dirname "$SOURCE_MOD")"
SOURCE_DIR=$(pwd)
cd - > /dev/null

cd "$(dirname "$TARGET_MOD")"
TARGET_DIR=$(pwd)
cd - > /dev/null

# Extract all dependencies and versions from source
echo "📦 Lese Dependencies aus $SOURCE_MOD ..."
cd "$SOURCE_DIR"
SOURCE_DEPS=$(go list -m all)

# Switch to target
cd "$TARGET_DIR"

echo "🛠️  Vergleiche mit $TARGET_MOD ..."
while read -r dep version; do
  # Skip the main module line
  [[ "$dep" == "=>" || "$dep" == "mod" || "$dep" == "" ]] && continue

  TARGET_VERSION=$(go list -m -f '{{.Version}}' "$dep" 2>/dev/null || echo "")

  if [[ "$TARGET_VERSION" != "$version" ]]; then
    echo "📌 Update $dep: $TARGET_VERSION → $version"
    go get "$dep@$version"
  fi
done <<< "$(echo "$SOURCE_DEPS" | tail -n +2)"  # Skip first line (main module)

echo "✅ Abgleich abgeschlossen. Führe jetzt 'go mod tidy' aus ..."
go mod tidy
