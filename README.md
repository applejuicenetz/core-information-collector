# appleJuice Core Information Collector

Dieses kleine Tool holt die Informationen von deinem Core (siehe unten `Platzhalter`) und leitet diese aufbereitet an eine definierte URL weiter.

Diese Informationen sind außerdem als Tooltip mittels des `info_line` Parameters konfigurierbar und werden ebenfalls als `stdOut` ausgegeben.

## Konfiguration

Die Datei `core-information-collector.properties` wird beim ersten Start automatisch im gleichen Ordner angelegt.

Sofern der Core auf dem gleichen Gerät läuft und kein Passwort hat, funktioniert das Tool ohne weiteres zutun.

Hat der Core ein Passwort und/oder läuft auf einem anderen Gerät, muss die ini Datei entsprechend angepasst werden.

| Konfiguration   | Wert         | Erklärung                  | Beispiel                                                                          |
|-----------------|--------------|----------------------------|-----------------------------------------------------------------------------------|
| `info_line `    | `Text`       | Text mit Platzhaltern      | `Credits %coreCredits% - Uploaded %coreSessionUpload% - Upload %coreUploadSpeed%` |
| `forward_url`   | `valid url`  | API URL für `forward_line` | `https://discord-bot.knastbruder.applejuicent.de/api/core-collector` oder `off`   |
| `forward_line`  | `Text`       | Text mit Platzhaltern      | `Credits %coreCredits% - Uploaded %coreSessionUpload% - Upload %coreUploadSpeed%` |
| `forward_token` | `Text`       | Auth Token für die API URL | `d9c1f872-5f48-42af-bd0d-601f2f05352a` (bekommst du vom Discord Bot)              |
| `interval`      | `60000`      | Millisekunden              | sollte nicht niedriger als `5000` (5 Sekunden) sein (Core überlastung möglich)    |
| `core_port`     | `9851`       | Core XML Port              | Der XML API Port des Core                                                         |
| `core_host`     | `valid host` | IP des Core mit Protokoll  | Bei den meisten `http://127.0.0.1`                                                |
| `core_passwd`   | `md5sum`     | MD5 Passwort vom Core      | `de305845b091d971732a123977e2d816` kann aus der `settings.xml` entnommen werden   |


## Platzhalter

Es sind folgende Platzhalter in `info_line` und `forward_line` möglich:

| Platzhalter             | Beispiel     | Wo        |
|-------------------------|--------------|-----------|
| `%coreVersion%`         | 0.31.149.111 | überall   |
| `%coreSystem%`          | Windows      | überall   |
| `%coreCredits%`         | 15,5GB       | überall   |
| `%coreSessionUpload%`   | 31GB         | überall   |
| `%coreSessionDownload%` | 2GB          | überall   |
| `%coreUploadSpeed%`     | 1,2MB/s      | überall   |
| `%coreDownloadSpeed%`   | 60kb/s       | überall   |
| `%coreUploads%`         | 12           | überall   |
| `%coreDownloads%`       | 8            | überall   |
| `%coreDownloadsReady%`  | 3            | überall   |
| `%networkUser%`         | 700          | info_line |
| `%networkFiles%`        | 3.182.468    | info_line |
| `%networkFileSize%`     | 798TB        | info_line |
