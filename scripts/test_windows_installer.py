"""Exercise the packaged MSI through the jpackage EXE on a Windows CI runner."""
import argparse
from pathlib import Path
import subprocess
import tempfile


def check_install(installer, target, log, expected, *properties):
    result = subprocess.run([
        str(installer), '/qn', '/norestart', '/L*v', str(log),
        f'INSTALLDIR={target}', *properties,
    ], timeout=180, check=False)
    text = log.read_text(encoding='utf-16', errors='replace') if log.exists() else ''
    if result.returncode not in expected:
        raise RuntimeError(f'Installer returned {result.returncode}:\n{text[-16000:]}')
    if 1603 in expected and 'Bitte zuerst das alte Setup deinstallieren' not in text:
        raise RuntimeError(f'Legacy guard did not report the intended error:\n{text[-16000:]}')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('installer', type=Path)
    args = parser.parse_args()
    installer = args.installer.resolve()
    if not installer.is_file():
        raise RuntimeError(f'Installer missing: {installer}')
    with tempfile.TemporaryDirectory(prefix='aj-installer-test-') as temporary:
        parent = Path(temporary)
        target = parent / 'Install Target'
        marker = target / 'Java'
        marker.mkdir(parents=True)
        check_install(installer, target, parent / 'legacy.log', (1603,))
        marker.rmdir()
        try:
            check_install(installer, target, parent / 'install.log', (0, 3010))
            if not (target / 'runtime').is_dir():
                raise RuntimeError(f'Installer ignored INSTALLDIR: {target}')
            check_install(installer, target, parent / 'repair.log', (0, 3010),
                          'REINSTALL=ALL', 'REINSTALLMODE=vomus')
        finally:
            result = subprocess.run([str(installer), 'uninstall'], timeout=180, check=False)
            if result.returncode not in (0, 3010):
                raise RuntimeError(f'Uninstall returned {result.returncode}')
    print('Windows installer: legacy rejection, install, repair and uninstall passed')


if __name__ == '__main__':
    main()
