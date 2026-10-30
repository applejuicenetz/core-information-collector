# appleJuice Collector

![](https://img.shields.io/github/v/release/applejuicenetz/collector.svg)
![](https://img.shields.io/github/downloads/applejuicenetz/collector/total)
![](https://img.shields.io/github/license/applejuicenetz/collector.svg)

Dieses kleine Tool holt die Informationen von deinem Core (siehe unten `Platzhalter`) und leitet diese aufbereitet an
eine definierte URL weiter.

Diese Informationen sind außerdem als Tooltip mittels des `infoLine`-Elements konfigurierbar und werden ebenfalls
als `stdOut` ausgegeben.

## Installation

| Plattform | Pakete |
|-----------|--------|
| Windows | [amd64 und aarch64 (`.exe`)](https://github.com/applejuicenetz/collector/releases) |
| macOS | [amd64 und aarch64 (`.dmg`)](https://github.com/applejuicenetz/collector/releases) |
| Linux | [amd64 und aarch64 (`.flatpak`)](https://github.com/applejuicenetz/collector/releases); alternativ über das [Flatpak-Repository](https://github.com/applejuicenetz/flatpak) |

Der Collector ist ein einzelnes Programm ohne Java-Laufzeit. Die Pakete enthalten es fertig; Linux-Nutzer können es alternativ über das Flatpak-Repository installieren. Eine vorhandene Java-Installation des Collectors muss vorher deinstalliert werden (Windows verlangt das ausdrücklich). Die Konfiguration `collector.xml` bleibt erhalten und wird weiterverwendet.

Start ohne Fenster und Symbol: `ajcollector -notray` (oder `AJ_NOTRAY=1`). Weitere Optionen: `-config <Ordner>` (oder `AJ_COLLECTOR_CONFIG_DIR`) wählt einen anderen Konfigurationsordner, `-version` zeigt die Version.

Der Collector zeigt ein Fenster mit allen aktuellen Werten und ein Symbol im Infobereich. Das Schließen-Kreuz (oder Esc) blendet das Fenster nur aus, ein Klick auf das Symbol zeigt es wieder. Das Menü des Symbols bietet `Status`, `Execute`, `Config`, `About` und `Quit`. Gibt es eine neuere Version, erscheint zusätzlich ein Eintrag zum Download. Ist `trayIcon` auf `false` gesetzt, beendet das Schließen des Fensters den Collector.

Ohne Desktop (zum Beispiel auf einem Server) oder mit `-notray` läuft der Collector ohne Fenster und ohne Symbol; die Werte stehen dann im Log.

## Changelog

Ein aktuelles Changelog befindet sich [hier](CHANGELOG.md)

## Konfiguration

Die Konfiguration erfolgt mittels XML Datei `collector.xml`. Eine vorhandene Datei mit dem bisherigen Namen wird beim ersten Start automatisch umbenannt.

Diese kannst du öffnen, in dem du den Collector startest und per rechtsklick den Menüpunkt `config` auswählst.

Alternativ findest du die Datei hier:

- Windows: `C:\Benutzer\%USERNAME%\appleJuice\collector\`
- Linux und macOS: `~/appleJuice/collector/`

Sofern der Core auf dem gleichen Gerät läuft und kein Passwort hat, funktioniert der Collector ohne weiteres zutun.

Hat der Core ein Passwort und/oder läuft auf einem anderen Gerät, muss die `.xml` Datei entsprechend angepasst werden.

| Konfiguration    | Wert         | Erklärung                  | Beispiel                                                                          |
|------------------|--------------|----------------------------|-----------------------------------------------------------------------------------|
| `infoLine`       | `Text`       | Text mit Platzhaltern      | `Credits %coreCredits% - Uploaded %coreSessionUpload% - Upload %coreUploadSpeed%` |
| `intervall`      | `60000`      | Millisekunden              | sollte nicht niedriger als `5000` (5 Sekunden) sein (Core überlastung möglich)    |
| `trayIcon`       | `true`       | zeige das TrayIcon         | `true` oder `false`, steuert das TrayIcon                                         |
| `taskbarIcon`    | `true`       | nur aus Kompatibilität     | wird nicht mehr ausgewertet, vorhandene Dateien bleiben gültig                    |
| `core > host`    | `valid host` | IP des Core mit Protokoll  | Bei den meisten `http://127.0.0.1`                                                |
| `core > port `   | `9851`       | Core XML Port              | Der XML API Port des Core                                                         |
| `core > timeout` | `5000`       | Millisekunden              | optional, Zeitüberschreitung für Anfragen an den Core                             |
| `core > password` | `md5sum`     | MD5 Passwort vom Core      | `de305845b091d971732a123977e2d816` kann aus der `settings.xml` entnommen werden   |
| `target > url`   | `valid url`  | Ziel URL                   | `https://discord.applejuicenet.cc/api/core-collector/`                            |
| `target > token` | `Text`       | Auth Token für die API URL | `d9c1f872-5f48-42af-bd0d-601f2f05352a`                                            |
| `target > line`  | `Text`       | Text mit Platzhaltern      | `Credits %coreCredits% - Uploaded %coreSessionUpload% - Upload %coreUploadSpeed%` |

im Block `<targets>` können mehrere `<target> </target>` Einträge existieren um die gesammelten Daten an mehrere
Endpunkte weiterzuleiten.

## Beispiel XML

Inhalt der `collector.xml` Datei

```xml
<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<collector intervall="60000" trayIcon="true" taskbarIcon="true">
    <infoLine>Core %coreVersion% - System %coreSystem% - Credits %coreCredits% - Uploaded %coreSessionUpload% -
        Downloaded %coreSessionDownload% - Upload %coreUploadSpeed% - Download
        %coreDownloadSpeed%
    </infoLine>
    <core host="http://127.0.0.1" password="" port="9851"/>
    <targets>
        <target>
            <url>https://discord.applejuicenet.cc/api/core-collector/</url>
            <token>_MEIN_TOKEN_</token>
            <line>Core `%coreVersion%` - Credits `%coreCredits%` - Uploaded `%coreSessionUpload%` - Downloaded
                `%coreSessionDownload%` - Upload `%coreUploadSpeed%` - Download `%coreDownloadSpeed%`
            </line>
        </target>
        <target>
            <url>https://www.irgendwo-anders.tld/api/core-collector/</url>
            <token>_MEIN_TOKEN_</token>
            <line>Core `%coreVersion%` - Credits `%coreCredits%` - Uploaded `%coreSessionUpload%` - Downloaded
                `%coreSessionDownload%` - Upload `%coreUploadSpeed%` - Download `%coreDownloadSpeed%`
            </line>
        </target>
    </targets>
</collector>
```

## Platzhalter

Es sind folgende Platzhalter in `infoLine` und `target > line` möglich:

| Platzhalter             | Beispiel     |
|-------------------------|--------------|
| `%coreVersion%`         | 0.31.149.112 |
| `%coreSystem%`          | Windows      |
| `%coreCredits%`         | 15,5GB       |
| `%coreConnections%`     | 21           |
| `%coreSessionUpload%`   | 31GB         |
| `%coreSessionDownload%` | 2GB          |
| `%coreUploadSpeed%`     | 1,2MB/s      |
| `%coreDownloadSpeed%`   | 60kb/s       |
| `%coreUploads%`         | 12           |
| `%coreDownloads%`       | 8            |
| `%coreDownloadsReady%`  | 3            |
| `%shareFiles%`          | 36           |
| `%shareSize%`           | 4,6GB        |
| `%networkUser%`         | 700          |
| `%networkFiles%`        | 3.182.468    |
| `%networkFileSize%`     | 798TB        |

## Discord Beispiele

Für `!aj` im Discord

### Nur Upload mit Discord Emojis

```plain
:green_apple: `%coreVersion%` :moneybag: `%coreCredits%` :arrow_upper_right: `%coreSessionUpload%` :arrow_up: `%coreUploadSpeed%` :handshake: `%coreConnections%`
```

### Upload und Download mit Discord Emojis

```plain
:green_apple: `%coreVersion%` :moneybag: `%coreCredits%` :handshake: `%coreConnections%` :arrow_up: `%coreUploadSpeed%` :arrow_down: `%coreDownloadSpeed%` :arrow_lower_right: `%coreSessionDownload%` :arrow_upper_right: `%coreSessionUpload%` :card_box: `%shareFiles%` (`%shareSize%`)
```

## als Docker Container

```yaml
services:
  applejuice_collector:
    container_name: applejuice_collector
    image: ghcr.io/applejuicenetz/core-information-collector:latest
    network_mode: bridge
    restart: always
    mem_limit: 32MB
    volumes:
      - ~/applejuice/collector.xml:/app/appleJuice/collector/collector.xml
```

Im Container läuft der Collector ohne Fenster und Symbol. Die Ausgabe steht im Container-Log.
