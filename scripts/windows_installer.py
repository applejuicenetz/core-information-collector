"""Add a native MSI registry guard for known legacy NSIS uninstall entries."""
import argparse
import os
from pathlib import Path
import zipfile
from xml.sax.saxutils import escape

LEGACY_PRODUCTS = {
    'Core': ('appleJuice Core (x86)', 'appleJuice Core (x64)', 'appleJuice Core (Beta)'),
    'Collector': ('appleJuice Collector',),
    'JavaGUI': ('appleJuice JavaGUI',),
}
UNINSTALL_KEY = r'Software\Microsoft\Windows\CurrentVersion\Uninstall'


def add_legacy_guard(template, folder):
    for marker in ('</Product>', '<InstallExecuteSequence>', '<InstallUISequence>'):
        if template.count(marker) != 1:
            raise RuntimeError(f'Unsupported jpackage WiX template: {marker}')
    if 'AjBlockLegacyInstall' in template:
        raise RuntimeError('Legacy guard already present')
    products = LEGACY_PRODUCTS[folder]
    declarations = []
    properties = []
    for index, product in enumerate(products):
        key = escape(f'{UNINSTALL_KEY}\\{product}', {'"': '&quot;'})
        for view in ('32', '64'):
            identifier = f'AJ_LEGACY_{index}_{view}'
            properties.append(identifier)
            declarations.append(f'''
    <Property Id="{identifier}">
      <RegistrySearch Id="AjLegacySearch{index}_{view}" Root="HKLM" Key="{key}"
                      Name="UninstallString" Type="raw" Win64="{'yes' if view == '64' else 'no'}"/>
    </Property>''')
    message = ('Eine alte NSIS-Installation wurde gefunden: ' + ', '.join(products) + '. '
               'Bitte zuerst das alte Setup deinstallieren und danach dieses Setup erneut starten. '
               'Die Installation wird abgebrochen.')
    declarations.append(f'''
    <CustomAction Id="AjBlockLegacyInstall" Error="{escape(message, {'"': '&quot;'})}"/>''')
    condition = 'NOT Installed AND NOT (REMOVE~="ALL") AND (' + ' OR '.join(properties) + ')'
    template = template.replace('</Product>', '\n'.join(declarations) + '\n  </Product>')
    template = template.replace('<InstallExecuteSequence>', f'''<InstallExecuteSequence>
      <Custom Action="AjBlockLegacyInstall" Before="RemoveExistingProducts">{condition}</Custom>''')
    template = template.replace('<InstallUISequence>', f'''<InstallUISequence>
      <Custom Action="AjBlockLegacyInstall" After="AppSearch">{condition}</Custom>''')
    return template


def windows_resources(java_home, directory, folder, template=None):
    directory = Path(directory).resolve()
    directory.mkdir(parents=True, exist_ok=True)
    if template is None:
        with zipfile.ZipFile(Path(java_home) / 'jmods/jdk.jpackage.jmod') as archive:
            template = archive.read('classes/jdk/jpackage/internal/resources/main.wxs').decode('utf-8')
    template = add_legacy_guard(template, folder)
    (directory / 'main.wxs').write_text(template, encoding='utf-8')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--folder', required=True, choices=LEGACY_PRODUCTS)
    parser.add_argument('--resource-dir', required=True, type=Path)
    args = parser.parse_args()
    windows_resources(Path(os.environ['JAVA_HOME']), args.resource_dir, args.folder)


if __name__ == '__main__':
    main()
