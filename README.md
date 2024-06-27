# appleJuice Collector

![](https://img.shields.io/github/v/release/applejuicenetz/core-information-collector.svg)
![](https://img.shields.io/github/downloads/applejuicenetz/core-information-collector/total)
![](https://img.shields.io/github/license/applejuicenetz/core-information-collector.svg)

![](https://github.com/applejuicenetz/core-information-collector/actions/workflows/container.yml/badge.svg)
![](https://img.shields.io/docker/pulls/applejuicenetz/core-information-collector)
![](https://img.shields.io/docker/image-size/applejuicenetz/core-information-collector)

![](https://github.com/applejuicenetz/core-information-collector/actions/workflows/snapcraft.yml/badge.svg)
![](https://snapcraft.io/applejuice-collector/badge.svg)

Dieses kleine Tool holt die Informationen von deinem Core (siehe unten `Platzhalter`) und leitet diese aufbereitet an eine definierte URL weiter.

Diese Informationen sind außerdem als Tooltip mittels des `info_line` Parameters konfigurierbar und werden ebenfalls als `stdOut` ausgegeben.


## Installation

| Platform 	| Link          	                                                                       |
|----------	|------------------------------------------------------------------------------------------|
| Windows  	| [setup.exe](https://github.com/applejuicenetz/core-information-collector/releases)   	   |
| macOS    	| [AJCollector.dmg](https://github.com/applejuicenetz/core-information-collector/releases) |
| Linux    	| [Snap Package](https://snapcraft.io/applejuice-collector)	                               |

## Changelog

Ein aktuelles Changelog befindet sich [hier](CHANGELOG.md)

## Konfiguration

Die Konfiguration erfolgt mittels XML Datei `core-information-collector.xml`

Diese kannst du öffnen, in dem du den Collector startest und per rechtsklick den Menüpunkt `config` auswählst.

Alternativ findest du die Datei hier:
- Windows: `C:\Benutzer\%USERNAME%\appleJuice\collector\`
- Linux und macOS: `~/appleJuice/collector/`


Sofern der Core auf dem gleichen Gerät läuft und kein Passwort hat, funktioniert der Collector ohne weiteres zutun.

Hat der Core ein Passwort und/oder läuft auf einem anderen Gerät, muss die `.xml` Datei entsprechend angepasst werden.

| Konfiguration    | Wert         | Erklärung                  | Beispiel                                                                          |
|------------------|--------------|----------------------------|-----------------------------------------------------------------------------------|
| `info_line`      | `Text`       | Text mit Platzhaltern      | `Credits %coreCredits% - Uploaded %coreSessionUpload% - Upload %coreUploadSpeed%` |
| `interval`       | `60000`      | Millisekunden              | sollte nicht niedriger als `5000` (5 Sekunden) sein (Core überlastung möglich)    |
| `trayIcon`       | `true`       | zeige das TrayIcon         | `true` oder `false`, steuert das TrayIcon                                         |
| `taskbarIcon`    | `true`       | zeige Icon in der Taskbar  | `true` oder `false`, steuert das Icon in der Taskbar                              |
| `core > host`    | `valid host` | IP des Core mit Protokoll  | Bei den meisten `http://127.0.0.1`                                                |
| `core > port `   | `9851`       | Core XML Port              | Der XML API Port des Core                                                         |
| `core > passwd`  | `md5sum`     | MD5 Passwort vom Core      | `de305845b091d971732a123977e2d816` kann aus der `settings.xml` entnommen werden   |
| `target > url`   | `valid url`  | Ziel URL                   | `https://discord.applejuicenet.cc/api/core-collector/`                        |
| `target > token` | `Text`       | Auth Token für die API URL | `d9c1f872-5f48-42af-bd0d-601f2f05352a`                                            |
| `target > line`  | `Text`       | Text mit Platzhaltern      | `Credits %coreCredits% - Uploaded %coreSessionUpload% - Upload %coreUploadSpeed%` |

im Block `<targets>` können mehrere `<target> </target>` Einträge existieren um die gesammelten Daten an mehrere Endpunkte weiterzuleiten.

## Beispiel XML

Inhalt der `core-information-collector.xml` Datei

```xml
<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<collector intervall="60000" trayIcon="true" taskbarIcon="true">
    <infoLine>Core %coreVersion% - System %coreSystem% - Credits %coreCredits% - Uploaded %coreSessionUpload% - Downloaded %coreSessionDownload% - Upload %coreUploadSpeed% - Download
        %coreDownloadSpeed%
    </infoLine>
    <core host="http://127.0.0.1" password="" port="9851"/>
    <targets>
        <target>
            <url>https://discord.applejuicenet.cc/api/core-collector/</url>
            <token>_MEIN_TOKEN_</token>
            <line>Core `%coreVersion%` - Credits `%coreCredits%` - Uploaded `%coreSessionUpload%` - Downloaded `%coreSessionDownload%` - Upload `%coreUploadSpeed%` - Download `%coreDownloadSpeed%`
            </line>
        </target>
        <target>
            <url>https://www.irgendwo-anders.tld/api/core-collector/</url>
            <token>_MEIN_TOKEN_</token>
            <line>Core `%coreVersion%` - Credits `%coreCredits%` - Uploaded `%coreSessionUpload%` - Downloaded `%coreSessionDownload%` - Upload `%coreUploadSpeed%` - Download `%coreDownloadSpeed%`
            </line>
        </target>
    </targets>
</collector>
```

## Platzhalter

Es sind folgende Platzhalter in `info_line` und `target > line` möglich:

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
:green_apple: `%coreVersion%` :moneybag: `%coreCredits%` :handshake: `%coreConnections%` :arrow_up: `%coreUploadSpeed%` :arrow_down: `%coreDownloadSpeed%` :arrow_lower_right: `%coreSessionDownload%` :arrow_upper_right: `%coreSessionUpload%`
```

## als Docker Container

```yaml
version: '2.4'

services:
    applejuice_collector:
        container_name: applejuice_collector
        image: ghcr.io/applejuicenetz/core-information-collector:latest
        network_mode: bridge
        restart: always
        mem_limit: 64MB
        volumes:
            - ~/applejuice/core-information-collector.xml:/app/appleJuice/collector/core-information-collector.xml
```
