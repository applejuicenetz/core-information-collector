# Agent Instructions core-information-collector

Go program (binary `ajcollector`) that reads statistics from an appleJuice Core XML API and forwards them as JSON to configured targets. Project-wide rules are in `../project/AGENTS.md`.

## Rules

- Standard library only, except `fyne.io/fyne/v2` for the status window and the tray icon (with its dependencies, including `fyne.io/systray`, used directly for the tooltip). Only `cmd/collector/gui.go` imports it; `internal/*` stays standard library. Do not add other third-party modules.
- Documentation and comments are English. `README.md` is German and contains user-facing information only. This file holds all technical documentation and states the current state only.
- Never reference `core-src` in this repository. Describe observable Core API behavior instead.
- No emojis. The product is written appleJuice, the organization appleJuiceNETZ.
- Compatibility contracts: `collector.xml` (element and attribute names, `intervall` with double l, `trayIcon`, `taskbarIcon`, `core host|port|password|timeout`, `targets/target/url|token|line`), the placeholder names, the forward JSON field names (`forward.Payload`) and the forward headers.

## Toolchain

Go 1.27 (`go.mod`). Not installed on the host; use Docker. Fyne needs CGO and the X11, OpenGL, Wayland and xkbcommon headers, so build and test in a `golang:1.27` image with these packages (`libgl1-mesa-dev xorg-dev libwayland-dev libxkbcommon-dev`, plus `xvfb` for window checks):

```sh
docker run --rm -v "$PWD":/src -w /src -e GOFLAGS=-buildvcs=false -e CGO_ENABLED=1 <image> sh -c \
  'test -z "$(gofmt -l .)" && go vet ./... && go test -race -timeout 300s ./...'
```

Without CGO or with `-tags nogui` the window and tray are compiled out (`gui_other.go`); the container image uses this (`CGO_ENABLED=0 go build -tags nogui`). Container image: `docker build -t applejuicenetz/collector:dev .`. Files written by a container into the repository belong to root; `chown` them afterwards.

## Layout

| Path | Purpose |
| --- | --- |
| `cmd/collector` | flags, wiring, update check, single instance check (`instance_*.go`); `gui.go` status window and tray menu (Fyne, build tag `cgo && !nogui`); `gui_other.go` headless stand-in; `tray.png` icon; `resource_windows_*.syso` Windows icon resource |
| `internal/config` | `collector.xml`: path resolution, legacy file rename, defaults file, parsing |
| `internal/core` | Core client: `information.xml`, `share.xml` (streamed), `modified.xml`; password masking; idle timeout |
| `internal/format` | placeholder number formats (binary sizes, German separators) |
| `internal/collector` | placeholders, replacement, status rows, one run (query, publish, forward), loop |
| `internal/forward` | JSON POST to a target |
| `internal/update` | latest release lookup and version comparison |
| `internal/status` | rows of the status window (sorted, `Status` first, change notification) |
| `internal/logx` | stdout log and rotating error log file |
| `internal/buildinfo` | version set by `-ldflags` |
| `packaging` | `package.sh`, Inno Setup script, macOS `Info.plist` |
| `flatpak` | Flatpak manifest and desktop files |
| `testdata` | mock Core documents, Java reference values |

## Behavior

- Configuration directory: `-config`, else `AJ_COLLECTOR_CONFIG_DIR`, else `~/appleJuice/collector`. `collector.xml` is created with defaults when missing. A writable `core-information-collector.xml` is renamed to it; a read-only one is used in place. Invalid or non-positive numbers (`intervall`, `core timeout`) log an error and use 60000 and 5000 ms. A target without `url` is a startup error. A URL containing `5f297e.online-server.cloud` is replaced by the default forward URL. An empty password is sent as the MD5 of the empty string. `taskbarIcon` is parsed but has no effect.
- One run requests `information.xml`, `share.xml` and `modified.xml` with `?password=<md5>`. Any failure, malformed value or missing element fails the whole run: nothing is published or forwarded, and the status rows reduce to `Status = Core-Abfrage fehlgeschlagen`. Runs never overlap. The first run starts immediately, then every `intervall`.
- The timeout applies to connecting and to the wait for the next data, not to the whole transfer, so a large `share.xml` that keeps flowing completes. `share.xml` is parsed as a stream (count and sum of `<share size>`). DTDs and entities are never resolved. Passwords never appear in errors or logs (`core.MaskPassword`); the wrong-password text of the Core is reported as `wrong password. access denied.`.
- `downloads` and `uploads` count every `<download>` and `<upload>` element; `coreDownloadsReady` counts top-level downloads with `ready="0"`. `networkinfo filesize` is megabytes, a comma decimal mark is accepted.
- Formats: `format.FileSize` uses binary units with at most two decimals and German separators regardless of host locale; `format.NetworkShareSize` cuts (not rounds) to two decimals. Both were compared with the former Java output (`testdata/java-reference.txt`, `TestAgainstJavaReference`); values below 2 MB and from 2^40 MB on are not compared because the Java output was wrong there.
- Forward: POST JSON with the fields of `forward.Payload`, `Authorization: Token <token>` only for a non-empty token, `User-Agent: Collector/<version>; Go/<go version>; (<os>/<arch>)`, 5 second timeout, status 400 and above is an error. A failing target is logged and does not stop the others or fail the run.
- Logging: `info:` and `error:` lines on stdout; errors also in `collector.log` in the configuration directory (rotated to `.old` at 1 MiB).
- Update check: once at start against the GitHub releases API, skipped for `dev` and `SNAPSHOT` builds and with `AJ_COLLECTOR_DISABLE_UPDATE_CHECK=yes` (the Flatpak sets it). A newer release adds a tray menu entry that opens the releases page.
- Window: a native Fyne window with a compact two-column grid (12 pt text) of all placeholders; it has a fixed size, is resized to the content after every row change (minimum width 300) and refreshes after every run (about 330x390 on Linux). With a tray (`trayIcon` not `false`) closing the window or Escape only hides it, a click on the tray icon shows it. Without a tray closing the window ends the program. The window needs a display: on Linux without `DISPLAY` and `WAYLAND_DISPLAY`, with `-notray`/`AJ_NOTRAY`, or in a build without CGO or with `-tags nogui` the collector runs headless without window and tray. `taskbarIcon` has no effect.
- Single instance: Windows uses a named mutex `AJCollectorSingleInstance` (`cmd/collector/instance_windows.go`) per user session; the Inno Setup `AppMutex` uses the same name, so setup and uninstall ask to close a running collector. Linux and macOS take an exclusive non-blocking `flock` on `collector.lock` in the configuration directory (`instance_other.go`); the kernel releases it when the process ends. The Flatpak shares `~/appleJuice` between starts, so a second Flatpak start is covered. A second start prints `AJCollector läuft bereits.` (Windows: message box) and exits with status 0. Two instances with different `-config` directories are allowed. A directory that cannot be locked does not block the start.
- Tray menu: Status shows the window, About opens a small fixed-size window with the name and version, Execute runs a query, Config opens `collector.xml` with the platform handler, Quit ends the program; an update entry appears for a newer release. The tooltip is the info line. Without a tray host (no session bus) there is no icon.
- The version comes from `-ldflags -X .../internal/buildinfo.Version=<tag>`; unset it is `dev`.

## Packaging and release

`.github/workflows/release.yml` runs on every tag, only for commits on `main`: tests, then four native packages (Windows `.exe` Inno Setup installers for amd64 and arm64, macOS `.dmg` for amd64 and arm64) and two Flatpak bundles, then the GitHub release with all six files, the release metadata commit on `main` and the dispatch to the flatpak repository. Every package is built with CGO on a runner of its own platform (`packaging/package.sh`); there is no cross compilation. The Windows arm64 runner has no ARM64 C compiler, so the workflow installs `llvm-mingw` (UCRT, aarch64 host) and sets `CC`/`CXX` for that package. A tag containing `-` (for example `5.0.0-beta.1`) is published as pre-release, not as latest; the release metadata commit and the dispatch to the flatpak repository are skipped for it. `container.yml` builds the multi-arch headless image (`VERSION` build argument) for tags. `ci.yml` checks formatting, vet, tests with `-race` and the headless build. The Flatpak manifest grants X11, Wayland, IPC and DRI access for the window.

`packaging/package.sh` needs `GOOS`, `GOARCH` and `VERSION`. The Windows setup refuses to install over a Java collector installation (uninstall entry named appleJuice Collector); its own uninstall entry is named `appleJuice Collector` without version (`UninstallDisplayName`), the version is shown only in the version column. The Windows icon resource is regenerated after icon changes with `go run github.com/akavel/rsrc@v0.10.2 -arch <amd64|arm64> -ico assets/windows/AJCollector.ico -o cmd/collector/resource_windows_<arch>.syso`. `cmd/collector/tray.png` is the tray icon.

## Tests

- Unit tests cover every package except the window and tray (`gui.go`) and `main`. `cmd/collector` tests run `run` end to end against fake Core and target servers.
- Documents in `testdata/mock-*.xml` and `mock-share.xml.gz` come from the local `core-mock`; the expected values were extracted with the Java implementation from the same documents.
- Live check: start the image against `core-mock` (`http://127.0.0.1:19851`, password `test`, MD5 `098f6bcd4621d373cade4e832627b4f6`) with `--network host` and a local receiver as target; remove the container afterwards.
- Window check: build the binary in the image above, start it under `Xvfb` with `XDG_RUNTIME_DIR` set against `core-mock`, find the window with `xwininfo` and capture it with ImageMagick `import`.
- Not covered by automated tests: the window contents and the tray on every platform (only the Linux window was started, under Xvfb), the Windows and macOS packages (built only in CI), the update check against GitHub, and runs against a real Core.
