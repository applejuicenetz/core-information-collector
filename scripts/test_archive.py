import tarfile
import tempfile
import unittest
import zipfile
from pathlib import Path

from archive import package_archive


class ArchiveTest(unittest.TestCase):
    def setUp(self):
        self.directory = tempfile.TemporaryDirectory()
        root = Path(self.directory.name)
        self.jar = root / 'AJCollector.jar'
        self.jar.write_bytes(b'collector')
        self.runtime = root / 'runtime'
        (self.runtime / 'bin').mkdir(parents=True)
        (self.runtime / 'bin/java').write_text('java')
        (self.runtime / 'bin/java.exe').write_text('java')
        self.output = root / 'out'

    def tearDown(self):
        self.directory.cleanup()

    def test_linux_launcher_uses_bundled_runtime_and_preview(self):
        output = package_archive(self.jar, self.runtime, self.output, 'linux', 'x86')
        self.assertEqual('AJCollector-linux-x86.tar.gz', output.name)
        with tarfile.open(output) as archive:
            launcher = archive.extractfile('AJCollector/ajcollector').read().decode()
            self.assertIn('jre/bin/java', launcher)
            self.assertIn('--enable-preview', launcher)
            self.assertIn('"$@"', launcher)
            self.assertTrue(archive.getmember('AJCollector/ajcollector').mode & 0o111)
            self.assertIn('AJCollector/AJCollector.jar', archive.getnames())

    def test_windows_zip_contains_cmd_launcher(self):
        output = package_archive(self.jar, self.runtime, self.output, 'windows', 'x86')
        self.assertEqual('AJCollector-windows-x86.zip', output.name)
        with zipfile.ZipFile(output) as archive:
            launcher = archive.read('AJCollector/ajcollector.cmd').decode()
            self.assertIn('jre\\bin\\java.exe', launcher)
            self.assertIn('--enable-preview', launcher)
            self.assertIn('AJCollector/AJCollector.jar', archive.namelist())


if __name__ == '__main__':
    unittest.main()
