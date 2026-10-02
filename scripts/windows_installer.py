"""Add the legacy NSIS Java-directory guard to a JDK 25 WiX template."""
import argparse
import os
import platform
from pathlib import Path
import subprocess
import tempfile
import zipfile
from xml.sax.saxutils import escape


def add_legacy_guard(template, dll, folder, product):
    for marker in ('</Product>', '<InstallExecuteSequence>', '<InstallUISequence>'):
        if template.count(marker) != 1:
            raise RuntimeError(f'Unsupported jpackage WiX template: {marker}')
    if 'AjCheckLegacyInstall' in template:
        raise RuntimeError('Legacy guard already present')
    message = ('Eine alte NSIS-Installation wurde gefunden (Unterordner Java in [AJ_LEGACY_PATH]). '
               'Bitte zuerst das alte Setup deinstallieren und danach dieses Setup erneut starten. '
               'Die Installation wird abgebrochen.')
    declarations = f'''
    <Property Id="AJ_LEGACY_FOLDER" Value="{escape(folder, {'"': '&quot;'})}"/>
    <Property Id="AJ_LEGACY_PRODUCT" Value="{escape(product, {'"': '&quot;'})}"/>
    <Binary Id="AjLegacyGuard" SourceFile="{escape(str(dll), {'"': '&quot;'})}"/>
    <CustomAction Id="AjCheckLegacyInstall" BinaryKey="AjLegacyGuard"
                  DllEntry="CheckLegacyInstallation" Execute="immediate" Return="check"/>
    <CustomAction Id="AjCheckLegacyTarget" BinaryKey="AjLegacyGuard"
                  DllEntry="CheckLegacyInstallation" Execute="immediate" Return="check"/>
    <CustomAction Id="AjBlockLegacyInstall" Error="{escape(message, {'"': '&quot;'})}"/>
    <CustomAction Id="AjBlockLegacyTarget" Error="{escape(message, {'"': '&quot;'})}"/>
'''
    template = template.replace('</Product>', declarations + '\n  </Product>')
    template = template.replace('<InstallExecuteSequence>', '''<InstallExecuteSequence>
      <Custom Action="AjCheckLegacyInstall" Before="RemoveExistingProducts">NOT (REMOVE~="ALL")</Custom>
      <Custom Action="AjBlockLegacyInstall" After="AjCheckLegacyInstall">AJ_LEGACY_PATH AND NOT (REMOVE~="ALL")</Custom>
      <Custom Action="AjCheckLegacyTarget" Before="InstallValidate">NOT (REMOVE~="ALL")</Custom>
      <Custom Action="AjBlockLegacyTarget" After="AjCheckLegacyTarget">AJ_LEGACY_PATH AND NOT (REMOVE~="ALL")</Custom>''')
    template = template.replace('<InstallUISequence>', '''<InstallUISequence>
      <Custom Action="AjCheckLegacyInstall" After="CostFinalize">NOT (REMOVE~="ALL")</Custom>
      <Custom Action="AjBlockLegacyInstall" After="AjCheckLegacyInstall">AJ_LEGACY_PATH AND NOT (REMOVE~="ALL")</Custom>
      <Custom Action="AjCheckLegacyTarget" Before="ExecuteAction">NOT (REMOVE~="ALL")</Custom>
      <Custom Action="AjBlockLegacyTarget" After="AjCheckLegacyTarget">AJ_LEGACY_PATH AND NOT (REMOVE~="ALL")</Custom>''')
    return template


def compile_guard(directory):
    vswhere = Path(os.environ['ProgramFiles(x86)']) / 'Microsoft Visual Studio/Installer/vswhere.exe'
    machines = (platform.machine(), os.environ.get('PROCESSOR_ARCHITECTURE', ''),
                os.environ.get('PROCESSOR_ARCHITEW6432', ''))
    native = any(machine.upper() in ('ARM64', 'AARCH64') for machine in machines)
    architecture = 'arm64' if native else 'x64'
    component = ('Microsoft.VisualStudio.Component.VC.Tools.ARM64' if native
                 else 'Microsoft.VisualStudio.Component.VC.Tools.x86.x64')
    installation = subprocess.check_output([
        str(vswhere), '-latest', '-products', '*', '-requires', component,
        '-property', 'installationPath'
    ], text=True).strip()
    if not installation:
        raise RuntimeError(f'MSVC {architecture} build tools required for the Windows MSI custom action')
    setup = Path(installation) / 'VC/Auxiliary/Build/vcvarsall.bat'
    source = Path(__file__).with_name('legacy_install_guard.c')
    dll = directory / 'legacy-install-guard.dll'
    probe = directory / 'legacy-install-guard-test.exe'
    script = directory / 'build-legacy-guard.bat'
    script.write_text('\r\n'.join([
        '@echo off',
        f'call "{setup}" {architecture} || exit /b 1',
        f'cl /nologo /W4 /WX /LD /MT "{source}" /link msi.lib advapi32.lib /OUT:"{dll}" || exit /b 1',
        f'cl /nologo /W4 /WX /wd4191 /MT /DAJ_GUARD_TEST "{source}" /link msi.lib advapi32.lib /STACK:262144 /OUT:"{probe}" || exit /b 1',
        '',
    ]), encoding='utf-8')
    subprocess.run([str(script)], cwd=directory, check=True)
    subprocess.run([str(probe), '--dll', str(dll)], check=True)
    with tempfile.TemporaryDirectory() as temporary:
        parent = Path(temporary)
        cases = [(parent / 'missing', 0), (parent, 0)]
        for path, expected in cases:
            result = subprocess.run([str(probe), str(path)], check=False)
            if result.returncode != expected:
                raise RuntimeError(f'Legacy guard test failed for {path}')
        marker = parent / 'Java'
        marker.write_text('not a directory', encoding='utf-8')
        if subprocess.run([str(probe), str(parent)], check=False).returncode != 0:
            raise RuntimeError('Java file incorrectly detected as legacy directory')
        marker.unlink()
        marker.mkdir()
        if subprocess.run([str(probe), str(parent)], check=False).returncode != 1:
            raise RuntimeError('Empty legacy Java directory not detected')
        (marker / 'java.exe').write_bytes(b'legacy')
        if subprocess.run([str(probe), str(parent)], check=False).returncode != 1:
            raise RuntimeError('Populated legacy Java directory not detected')
    return dll


def windows_resources(java_home, directory, folder, product, template=None):
    directory = Path(directory).resolve()
    directory.mkdir(parents=True, exist_ok=True)
    if template is None:
        with zipfile.ZipFile(Path(java_home) / 'jmods/jdk.jpackage.jmod') as archive:
            template = archive.read('classes/jdk/jpackage/internal/resources/main.wxs').decode('utf-8')
    dll = compile_guard(directory)
    template = add_legacy_guard(template, dll, folder, product)
    (directory / 'main.wxs').write_text(template, encoding='utf-8')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--folder', required=True)
    parser.add_argument('--product', required=True)
    parser.add_argument('--resource-dir', required=True, type=Path)
    args = parser.parse_args()
    windows_resources(Path(os.environ['JAVA_HOME']), args.resource_dir, args.folder, args.product)


if __name__ == '__main__':
    main()
