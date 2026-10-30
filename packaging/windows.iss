#ifndef Version
#define Version "0.0.0"
#endif
#ifndef Arch
#define Arch "amd64"
#endif
#ifndef Asset
#define Asset "AJCollector-windows-" + Arch
#endif
#ifndef PackageDir
#define PackageDir "..\dist\ajcollector-" + Version + "-windows-" + Arch
#endif

[Setup]
AppId={{6119FFB5-F01B-4162-AB46-A62D5DD16F37}
AppName=appleJuice Collector
AppVersion={#Version}
AppPublisher=appleJuiceNETZ
AppPublisherURL=https://applejuicenetz.github.io
AppSupportURL=https://applejuicenetz.github.io/faq/
DefaultDirName={autopf}\appleJuiceNETZ\Collector
DefaultGroupName=appleJuiceNETZ
PrivilegesRequired=admin
#if Arch == "arm64"
ArchitecturesAllowed=arm64
ArchitecturesInstallIn64BitMode=arm64
#else
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
#endif
OutputDir=..\dist\assets
OutputBaseFilename={#Asset}
SetupIconFile=..\assets\windows\AJCollector.ico
UninstallDisplayName=appleJuice Collector
UninstallDisplayIcon={app}\ajcollector.exe
Compression=lzma2
SolidCompression=yes
WizardStyle=modern
CloseApplications=yes
AppMutex=AJCollectorSingleInstance

[Files]
Source: "{#PackageDir}\ajcollector.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "{#PackageDir}\README.md"; DestDir: "{app}"; Flags: ignoreversion
Source: "{#PackageDir}\LICENSE"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\appleJuice Collector"; Filename: "{app}\ajcollector.exe"; WorkingDir: "{app}"
Name: "{autodesktop}\appleJuice Collector"; Filename: "{app}\ajcollector.exe"; WorkingDir: "{app}"

[Run]
Filename: "{app}\ajcollector.exe"; Description: "Start appleJuice Collector"; Flags: postinstall nowait skipifsilent

[Code]
// An installation of the former Java collector (jpackage or NSIS setup) registers an uninstall
// entry named "appleJuice Collector" under a different key than this setup (<AppId>_is1).
function JavaCollector(Root: Integer; var Found: String): Boolean;
var
  Keys: TArrayOfString;
  I: Integer;
  Name: String;
begin
  Result := False;
  if RegGetSubkeyNames(Root, 'Software\Microsoft\Windows\CurrentVersion\Uninstall', Keys) then
    for I := 0 to GetArrayLength(Keys) - 1 do begin
      Name := '';
      RegQueryStringValue(Root, 'Software\Microsoft\Windows\CurrentVersion\Uninstall\' + Keys[I], 'DisplayName', Name);
      if (Name = 'appleJuice Collector') and (Pos('_is1', Keys[I]) = 0) then begin
        Found := Keys[I];
        Result := True;
        Exit;
      end;
    end;
end;

function PrepareToInstall(var NeedsRestart: Boolean): String;
var
  Found: String;
begin
  Result := '';
  Found := '';
  if not JavaCollector(HKCU, Found) then
    if not JavaCollector(HKLM32, Found) then
      if IsWin64 then JavaCollector(HKLM64, Found);
  if Found <> '' then
    Result := 'An old Java appleJuice Collector installation was found.' + #13#10 +
      'Uninstall it first, then run this setup again. The configuration is kept.';
end;
