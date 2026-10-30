#!/usr/bin/env bash
# Builds ajcollector for GOOS/GOARCH and packages it into dist/assets.
set -euo pipefail
: "${GOOS:?}" "${GOARCH:?}" "${VERSION:?}"
name="ajcollector-${VERSION}-${GOOS}-${GOARCH}"
platform="$GOOS"
if [ "$GOOS" = darwin ]; then platform=macos; fi
asset_arch="$GOARCH"
if [ "$GOARCH" = arm64 ]; then asset_arch=aarch64; fi
asset="AJCollector-${platform}-${asset_arch}"
mkdir -p "dist/$name" dist/assets
ext=""
flags="-s -w"
if [ "$GOOS" = windows ]; then
  ext=.exe
  flags="$flags -H windowsgui"
fi
pkg=github.com/applejuicenetz/core-information-collector/internal/buildinfo
flags="$flags -X $pkg.Version=$VERSION"
# The window needs cgo (Fyne), so every target is built natively on its own platform.
CGO_ENABLED=1 go build -trimpath -ldflags "$flags" -o "dist/$name/ajcollector$ext" ./cmd/collector
cp LICENSE README.md "dist/$name/"
case "$GOOS" in
  windows)
    iscc="${ISCC:-/c/Program Files (x86)/Inno Setup 6/ISCC.exe}"
    # Git Bash rewrites arguments that start with a slash; disable that for ISCC options.
    MSYS_NO_PATHCONV=1 MSYS2_ARG_CONV_EXCL='*' "$iscc" "-DVersion=$VERSION" "-DArch=$GOARCH" "-DAsset=$asset" packaging/windows.iss
    ;;
  linux)
    # The Flatpak manifest takes dist/ajcollector (one architecture per build); there is no
    # separate Linux archive.
    cp "dist/$name/ajcollector" dist/ajcollector
    ;;
  darwin)
    app="dist/dmg/AJCollector.app"
    rm -rf dist/dmg
    mkdir -p "$app/Contents/MacOS" "$app/Contents/Resources"
    cp "dist/$name/ajcollector" "$app/Contents/MacOS/AJCollector"
    cp assets/mac/AJCollector.icns "$app/Contents/Resources/AJCollector.icns"
    sed "s/@VERSION@/${VERSION%%-*}/g" packaging/macos/Info.plist > "$app/Contents/Info.plist"
    ln -s /Applications dist/dmg/Applications
    hdiutil create -volname "appleJuice Collector $VERSION" -srcfolder dist/dmg -ov -format UDZO "dist/assets/$asset.dmg"
    ;;
esac
