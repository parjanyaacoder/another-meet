import os
import platform
import urllib.request
import tarfile
import zipfile
import shutil
import stat
import tempfile
from setuptools import setup
from setuptools.command.install import install

VERSION = "1.0.3"
REPO = "parjanyaacoder/another-meet"

class CustomInstall(install):
    def run(self):
        install.run(self)
        
        system = platform.system().lower()
        machine = platform.machine().lower()

        if system == "darwin":
            os_name = "darwin"
        elif system == "linux":
            os_name = "linux"
        elif system == "windows":
            os_name = "windows"
        else:
            raise Exception(f"Unsupported OS: {system}")

        if machine in ["x86_64", "amd64"]:
            arch_name = "amd64"
        elif machine in ["arm64", "aarch64"]:
            arch_name = "arm64"
        else:
            raise Exception(f"Unsupported architecture: {machine}")

        ext = "zip" if os_name == "windows" else "tar.gz"
        filename = f"another-meet_{VERSION}_{os_name}_{arch_name}.{ext}"
        url = f"https://github.com/{REPO}/releases/download/v{VERSION}/{filename}"

        print(f"Downloading {url}...")
        
        with tempfile.TemporaryDirectory() as tmpdirname:
            tmp_file = os.path.join(tmpdirname, filename)
            urllib.request.urlretrieve(url, tmp_file)

            bin_dir = self.install_scripts
            os.makedirs(bin_dir, exist_ok=True)

            extract_dir = os.path.join(tmpdirname, "extract")
            os.makedirs(extract_dir, exist_ok=True)

            if ext == "zip":
                with zipfile.ZipFile(tmp_file, 'r') as zip_ref:
                    zip_ref.extractall(extract_dir)
            else:
                with tarfile.open(tmp_file, "r:gz") as tar:
                    tar.extractall(path=extract_dir)

            bin_name = "another-meet.exe" if os_name == "windows" else "another-meet"
            src_bin = os.path.join(extract_dir, bin_name)
            dst_bin = os.path.join(bin_dir, bin_name)

            shutil.copy(src_bin, dst_bin)

            if os_name != "windows":
                st = os.stat(dst_bin)
                os.chmod(dst_bin, st.st_mode | stat.S_IEXEC)

        print("Installation complete!")

with open("README.md", "r", encoding="utf-8") as fh:
    long_description = fh.read()

setup(
    name="another-meet",
    version=VERSION,
    description="Manage Google Meet meetings from your terminal",
    long_description=long_description,
    long_description_content_type="text/markdown",
    author="parjanyaacoder",
    url="https://github.com/parjanyaacoder/another-meet",
    cmdclass={
        'install': CustomInstall,
    },
)
