"""Bundle Collector with a verified Java 21 32-bit runtime."""
import argparse
from pathlib import Path
import shutil
import tarfile
import tempfile
import zipfile


def package_archive(jar, runtime, destination, platform, arch):
    destination.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory() as temporary:
        root = Path(temporary) / 'AJCollector'
        root.mkdir()
        shutil.copy2(jar, root / 'AJCollector.jar')
        shutil.copytree(runtime, root / 'jre')
        options = '--enable-preview --enable-native-access=ALL-UNNAMED -Xmx384m -XX:+UseSerialGC -Djava.net.preferIPv4Stack=true'
        if platform == 'windows':
            (root / 'ajcollector.cmd').write_text(
                '@echo off\r\ncd /d "%~dp0"\r\n'
                f'"%~dp0jre\\bin\\java.exe" {options} -jar "%~dp0AJCollector.jar" %*\r\n', encoding='utf-8')
            output = destination / f'AJCollector-windows-{arch}.zip'
            with zipfile.ZipFile(output, 'w', zipfile.ZIP_DEFLATED) as archive:
                for file in sorted(root.rglob('*')):
                    if file.is_file():
                        archive.write(file, file.relative_to(root.parent))
        else:
            launcher = root / 'ajcollector'
            launcher.write_text(
                '#!/bin/sh\nset -eu\nROOT=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)\ncd "$ROOT"\n'
                f'exec "$ROOT/jre/bin/java" {options} -Dsun.java2d.xrender=false -jar "$ROOT/AJCollector.jar" "$@"\n', encoding='utf-8')
            launcher.chmod(0o755)
            output = destination / f'AJCollector-linux-{arch}.tar.gz'
            with tarfile.open(output, 'w:gz') as archive:
                archive.add(root, arcname=root.name)
    return output


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--jar', type=Path, required=True)
    parser.add_argument('--runtime', type=Path, required=True)
    parser.add_argument('--destination', type=Path, default=Path('target'))
    parser.add_argument('--platform', choices=['linux', 'windows'], required=True)
    parser.add_argument('--arch', choices=['x86', 'armhf'], required=True)
    args = parser.parse_args()
    print(package_archive(args.jar, args.runtime, args.destination, args.platform, args.arch))
