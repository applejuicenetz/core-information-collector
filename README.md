# appleJuice Core Information Collector

![](https://img.shields.io/github/v/release/applejuicenet/core-information-collector.svg)
![](https://img.shields.io/github/license/applejuicenet/core-information-collector.svg)
![](https://img.shields.io/docker/pulls/applejuicenet/core-information-collector.svg)
![](https://img.shields.io/docker/image-size/applejuicenet/core-information-collector)
![](https://github.com/applejuicenet/core-information-collector/workflows/docker/badge.svg)

Dieses kleine Tool holt die Informationen von deinem Core (siehe unten `Platzhalter`) und leitet diese aufbereitet an eine definierte URL weiter.

Diese Informationen sind außerdem als Tooltip mittels des `info_line` Parameters konfigurierbar und werden ebenfalls als `stdOut` ausgegeben.

## Download

rechts auf `Releases` klicken :wink: 

## Konfiguration

Die Datei `core-information-collector.properties` wird beim ersten Start automatisch im gleichen Ordner angelegt.

Sofern der Core auf dem gleichen Gerät läuft und kein Passwort hat, funktioniert das Tool ohne weiteres zutun.

Hat der Core ein Passwort und/oder läuft auf einem anderen Gerät, muss die `.properties` Datei entsprechend angepasst werden.

| Konfiguration   | Wert         | Erklärung                  | Beispiel                                                                          |
|-----------------|--------------|----------------------------|-----------------------------------------------------------------------------------|
| `info_line`     | `Text`       | Text mit Platzhaltern      | `Credits %coreCredits% - Uploaded %coreSessionUpload% - Upload %coreUploadSpeed%` |
| `forward_url`   | `valid url`  | API URL für `forward_line` | `https://discord-bot.knastbruder.applejuicent.de/api/core-collector` oder `off`   |
| `forward_line`  | `Text`       | Text mit Platzhaltern      | `Credits %coreCredits% - Uploaded %coreSessionUpload% - Upload %coreUploadSpeed%` |
| `forward_token` | `Text`       | Auth Token für die API URL | `d9c1f872-5f48-42af-bd0d-601f2f05352a` (bekommst du vom Discord Bot)              |
| `interval`      | `60000`      | Millisekunden              | sollte nicht niedriger als `5000` (5 Sekunden) sein (Core überlastung möglich)    |
| `core_port`     | `9851`       | Core XML Port              | Der XML API Port des Core                                                         |
| `core_host`     | `valid host` | IP des Core mit Protokoll  | Bei den meisten `http://127.0.0.1`                                                |
| `core_passwd`   | `md5sum`     | MD5 Passwort vom Core      | `de305845b091d971732a123977e2d816` kann aus der `settings.xml` entnommen werden   |

alle Konfigurationswerte können auch als `Environment` Umgebung definiert werden, **müssen** dann aber als `Caps` geschrieben werden (siehe unten Docker Beispiel).

## Platzhalter

Es sind folgende Platzhalter in `info_line` und `forward_line` möglich:

| Platzhalter             | Beispiel     |
|-------------------------|--------------|
| `%coreVersion%`         | 0.31.149.111 |
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

## als Docker Container

```yaml
version: '2.4'

services:
    applejuice_core_collector:
        container_name: applejuice_core_collector
        image: applejuicenet/core-information-collector:latest
        network_mode: bridge
        restart: always
        mem_limit: 32MB
        environment:
            TZ: Europe/Berlin
            CORE_HOST: http://192.168.155.10
            CORE_PORT: 9851
            CORE_PASSWD: de305845b091d971732a123977e2d816
            FORWARD_TOKEN: bbd04788-0000-0000-0000-687bdf011a7a
```
