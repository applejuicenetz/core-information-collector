# appleJuice Core Information Collector

Dieses kleine Tool holt die Informationen von deinem Core (Credits, Upload, Download usw.) und leitet diese an eine definierte URL weiter.

Außerdem werden ein Teil der Informationen als `stdOut` ausgegeben.

## Konfiguration

Die Datei `core-information-collector.properties` wird beim ersten Start automatisch im gleichen Ordner angelegt.

Sofern der Core auf dem gleichen Gerät läuft und kein Passwort hat, funktioniert das Tool ohne weiteres zutun.

Hat der Core ein Passwort und/oder läuft auf einem anderen Gerät, muss die ini Datei entsprechend angepasst werden.

Achtung: der Wert `core_passwd` muss die `md5sum` sein, das kann z.B. auf https://www.md5hashgenerator.com umgewandelt werden.

Alternativ kopiert man sich das MD5-Passwort aus der `settings.xml` vom Core.

```ini
collector_url=https://discord-bot.knastbruder.applejuicent.de/api/core-collector
collector_token=TOKEN_VOM_DISCORD_BOT
core_port=9851
core_host=http://192.168.155.10
core_passwd=
interval=60000
```

### Collector URL

Als standard ist unser [Discord Bot](https://github.com/applejuicenet/discord-bot) hinterlegt.

Wird der Wert geleert, werden keine Daten übermittelt und das Programm zeigt lediglich im SystemTray die aktuellen Core Informationen als Tooltip ;) 

### Collector Token

Das token bekommst du vom `appleJuiceNET` Bot als private Nachricht.

Dafür musst du lediglich `!aj` in den Channel Tippen wo der Bot anwesend ist.

Bei Ersteinrichtung hat das Token lediglich eine gültigkeit von 15 Minuten.
Sobald die ersten Daten für das Token empfangen wurden, bleibt das Token dauerhaft gültig.
