# Changelog

## 4.0.0

- Mit Java 25 gebaut.
- Native Pakete für Windows und macOS sowie Flatpak-Bundles für Linux werden jeweils für `amd64` und `aarch64` auf der passenden Architektur erstellt. Alle Pakete enthalten eine Java-25-Laufzeit; eine separate Java-Installation ist nicht erforderlich.
- Die Konfigurationsdatei wurde von `core-information-collector.xml` zu `collector.xml` umbenannt. Eine vorhandene Datei mit altem Namen wird beim Start automatisch übernommen.
- Zeitüberschreitung für HTTP-Anfragen an den Core auf fünf Sekunden gesetzt.

## 3.1.0

- Die Werte `shareFiles` und `shareSize` stehen nun als Platzhalter und in der JSON-Nutzlast zur Verfügung.

## 3.0.4

- Die veraltete Collector-URL `5f297e.online-server.cloud` wird nun in allen Fällen durch `discord.applejuicenet.cc` ersetzt. Dabei wird eine Logmeldung ausgegeben.

## 3.0.3

- Die veraltete Collector-URL `5f297e.online-server.cloud` wurde durch `discord.applejuicenet.cc` ersetzt.

## 3.0.2

- Menüpunkt `config` zum Taskleisten-Symbol hinzugefügt; er öffnet die XML-Datei direkt.

## 3.0.1

- `WM_CLASS` unter Linux korrigiert.
- Die Umgebungsvariable `AJ_COLLECTOR_DISABLE_UPDATE_CHECK` mit dem Wert `yes` deaktiviert die Versionsprüfung.

## 3.0.0

- Mindestens Java 11 als Laufzeitumgebung (JRE) erforderlich.
- Taskleisten-Unterstützung für alle Betriebssysteme hinzugefügt.
- Die Einstellung `intervall` wird nun berücksichtigt.
- Das Statusfenster wird gemäß dem konfigurierten Intervall aktualisiert.
- Snap-Paket für Linux verfügbar.
- Der Konfigurationsordner wird beim ersten Start angelegt.
- Änderungsprotokoll hinzugefügt.

## 2.1.4

- Neues Icon.
- Die Windows-EXE verwendet den `Java`-Ordner.
- Die Konfigurationsdatei wird beim ersten Start angelegt.

## 2.1.3

- Die Anwendung wird bei schwerwiegenden Konfigurationsfehlern beendet.

## 2.1.2

- Konfiguration auf XML umgestellt; mehrere Ziele sind möglich.
- Native macOS-App mit `javapackager` erstellt.

## 2.0.3

- Die Version wird beim Start auf der Standardausgabe (`stdout`) angezeigt.
- Statusfenster hinzugefügt.
- Nutzlast im JSON-Format.

## 1.X

- Erste Veröffentlichung.
