"""Exercise the MSI embedded in a jpackage EXE on a Windows CI runner."""
import argparse
from contextlib import contextmanager
import ctypes
from ctypes import wintypes
from pathlib import Path
import subprocess
import tempfile
import winreg

from windows_installer import DISPLAY_NAME, LEGACY_PRODUCTS, UNINSTALL_KEY


def check_shortcuts(target):
    command = ("[Environment]::GetFolderPath('CommonDesktopDirectory'); "
               "[Environment]::GetFolderPath('CommonPrograms')")
    folders = subprocess.check_output(['pwsh.exe', '-NoProfile', '-NonInteractive',
                                       '-Command', command], text=True, encoding='utf-8').splitlines()
    if len(folders) != 2 or not all(folders):
        raise RuntimeError(f'Cannot resolve Windows shortcut folders: {folders}')
    desktop = Path(folders[0]) / f'{DISPLAY_NAME}.lnk'
    menu = Path(folders[1]) / 'appleJuiceNETZ' / f'{DISPLAY_NAME}.lnk'
    for shortcut in (desktop, menu):
        if not shortcut.is_file():
            raise RuntimeError(f'Shortcut missing: {shortcut}')
        escaped = str(shortcut).replace("'", "''")
        command = f"(New-Object -ComObject WScript.Shell).CreateShortcut('{escaped}').TargetPath"
        destination = subprocess.check_output(['pwsh.exe', '-NoProfile', '-NonInteractive',
                                               '-Command', command], text=True, encoding='utf-8').strip()
        if not Path(destination).samefile(target / f'{DISPLAY_NAME}.exe'):
            raise RuntimeError(f'Shortcut target incorrect: {shortcut}: {destination}')
    return desktop, menu


@contextmanager
def legacy_entry(product, view):
    path = f'{UNINSTALL_KEY}\\{product}'
    try:
        existing = winreg.OpenKey(winreg.HKEY_LOCAL_MACHINE, path, 0, winreg.KEY_READ | view)
    except FileNotFoundError:
        pass
    else:
        existing.Close()
        raise RuntimeError(f'Refusing to overwrite existing uninstall entry: {path}')
    with winreg.CreateKeyEx(winreg.HKEY_LOCAL_MACHINE, path, 0, winreg.KEY_WRITE | view) as entry:
        winreg.SetValueEx(entry, 'DisplayName', 0, winreg.REG_SZ, product)
        winreg.SetValueEx(entry, 'UninstallString', 0, winreg.REG_SZ,
                         r'"C:\Removed NSIS Installation\uninstaller.exe"')
    try:
        yield
    finally:
        winreg.DeleteKeyEx(winreg.HKEY_LOCAL_MACHINE, path, view)


def extract_msi(installer, destination):
    kernel = ctypes.WinDLL('kernel32', use_last_error=True)
    kernel.LoadLibraryExW.argtypes = [wintypes.LPCWSTR, wintypes.HANDLE, wintypes.DWORD]
    kernel.LoadLibraryExW.restype = wintypes.HMODULE
    kernel.FindResourceW.argtypes = [wintypes.HMODULE, wintypes.LPCWSTR, ctypes.c_void_p]
    kernel.FindResourceW.restype = wintypes.HANDLE
    kernel.SizeofResource.argtypes = [wintypes.HMODULE, wintypes.HANDLE]
    kernel.SizeofResource.restype = wintypes.DWORD
    kernel.LoadResource.argtypes = [wintypes.HMODULE, wintypes.HANDLE]
    kernel.LoadResource.restype = wintypes.HANDLE
    kernel.LockResource.argtypes = [wintypes.HANDLE]
    kernel.LockResource.restype = ctypes.c_void_p
    kernel.FreeLibrary.argtypes = [wintypes.HMODULE]
    module = kernel.LoadLibraryExW(str(installer), None, 2)
    if not module:
        raise ctypes.WinError(ctypes.get_last_error())
    try:
        resource = kernel.FindResourceW(module, 'MSI', ctypes.c_void_p(10))
        if not resource:
            raise ctypes.WinError(ctypes.get_last_error())
        size = kernel.SizeofResource(module, resource)
        pointer = kernel.LockResource(kernel.LoadResource(module, resource))
        if not size or not pointer:
            raise ctypes.WinError(ctypes.get_last_error())
        destination.write_bytes(ctypes.string_at(pointer, size))
    finally:
        kernel.FreeLibrary(module)


def check_install(installer, target, log, expected, *properties):
    try:
        command = subprocess.list2cmdline([
            'msiexec.exe', '/i', str(installer), '/qn', '/norestart', '/L*v', str(log),
        ]) + f' INSTALLDIR="{target}"'
        if properties:
            command += ' ' + ' '.join(properties)
        result = subprocess.run(command, timeout=180, check=False)
    except subprocess.TimeoutExpired:
        if log.exists():
            print(log.read_text(encoding='utf-16', errors='replace')[-16000:], flush=True)
        raise
    text = log.read_text(encoding='utf-16', errors='replace') if log.exists() else ''
    if result.returncode not in expected:
        raise RuntimeError(f'Installer returned {result.returncode}:\n{text[-16000:]}')
    if 1603 in expected and 'Bitte zuerst das alte Setup deinstallieren' not in text:
        raise RuntimeError(f'Legacy guard did not report the intended error:\n{text[-16000:]}')
    if 1603 in expected and f'Alte Installation von {DISPLAY_NAME} gefunden.' not in text:
        raise RuntimeError(f'Legacy guard did not report the generic product name:\n{text[-16000:]}')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('installer', type=Path)
    args = parser.parse_args()
    installer = args.installer.resolve()
    if not installer.is_file():
        raise RuntimeError(f'Installer missing: {installer}')
    with tempfile.TemporaryDirectory(prefix='aj-installer-test-') as temporary:
        parent = Path(temporary)
        msi = parent / 'installer.msi'
        extract_msi(installer, msi)
        target = parent / 'Install Target'
        marker = target / 'Java'
        marker.mkdir(parents=True)
        products = LEGACY_PRODUCTS
        for index, product in enumerate(products):
            for view_name, view in (('32', winreg.KEY_WOW64_32KEY), ('64', winreg.KEY_WOW64_64KEY)):
                with legacy_entry(product, view):
                    check_install(msi, parent / 'Clean Target',
                                  parent / f'legacy-{index}-{view_name}.log', (1603,))
                print(f'Legacy registry rejection passed: {product} ({view_name}-bit)', flush=True)
        try:
            with legacy_entry(products[0] + ' unrelated', winreg.KEY_WOW64_32KEY):
                check_install(msi, target, parent / 'install.log', (0, 3010))
            if not (target / 'runtime').is_dir():
                raise RuntimeError(f'Installer ignored INSTALLDIR: {target}')
            if not marker.is_dir():
                raise RuntimeError('Installer removed the unrelated Java directory')
            shortcuts = check_shortcuts(target)
            with legacy_entry(products[-1], winreg.KEY_WOW64_64KEY):
                check_install(msi, target, parent / 'repair.log', (0, 3010),
                              'REINSTALL=ALL', 'REINSTALLMODE=vomus')
        finally:
            with legacy_entry(products[-1], winreg.KEY_WOW64_64KEY):
                result = subprocess.run(['msiexec.exe', '/x', str(msi), '/qn', '/norestart'],
                                        timeout=180, check=False)
            if result.returncode not in (0, 3010):
                raise RuntimeError(f'Uninstall returned {result.returncode}')
        if any(shortcut.exists() for shortcut in shortcuts):
            raise RuntimeError('Uninstall left product shortcuts behind')
    print('Windows installer: registry rejection, generic error, shortcut names, install, repair and uninstall passed')


if __name__ == '__main__':
    main()
