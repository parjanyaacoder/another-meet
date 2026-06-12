import os
import sys
import platform
import urllib.request
import tarfile
import zipfile
import stat
import tempfile
import subprocess

VERSION = "1.1.0"
REPO = "parjanyaacoder/another-meet"

def main():
    cache_dir = os.path.expanduser("~/.another-meet/bin")
    os.makedirs(cache_dir, exist_ok=True)
    
    system = platform.system().lower()
    bin_name = "another-meet.exe" if system == "windows" else "another-meet"
    bin_path = os.path.join(cache_dir, f"another-meet-{VERSION}-{bin_name}")
    
    if not os.path.exists(bin_path):
        print(f"First run: Downloading another-meet v{VERSION} for your system...")
        machine = platform.machine().lower()
        
        if system == "darwin":
            os_name = "darwin"
        elif system == "linux":
            os_name = "linux"
        elif system == "windows":
            os_name = "windows"
        else:
            print(f"Unsupported OS: {system}")
            sys.exit(1)
            
        if machine in ["x86_64", "amd64"]:
            arch_name = "x86_64"
        elif machine in ["arm64", "aarch64"]:
            arch_name = "arm64"
        elif machine in ["i386", "i686", "x86"]:
            arch_name = "i386"
        else:
            print(f"Unsupported architecture: {machine}")
            sys.exit(1)
            
        ext = "zip" if os_name == "windows" else "tar.gz"
        filename = f"another-meet_{VERSION}_{os_name}_{arch_name}.{ext}"
        url = f"https://github.com/{REPO}/releases/download/v{VERSION}/{filename}"
        
        try:
            with tempfile.TemporaryDirectory() as tmpdirname:
                tmp_file = os.path.join(tmpdirname, filename)
                urllib.request.urlretrieve(url, tmp_file)
                
                extract_dir = os.path.join(tmpdirname, "extract")
                os.makedirs(extract_dir, exist_ok=True)
                
                if ext == "zip":
                    with zipfile.ZipFile(tmp_file, 'r') as zip_ref:
                        zip_ref.extractall(extract_dir)
                else:
                    with tarfile.open(tmp_file, "r:gz") as tar:
                        tar.extractall(path=extract_dir)
                        
                src_bin = os.path.join(extract_dir, bin_name)
                with open(src_bin, 'rb') as f_src, open(bin_path, 'wb') as f_dst:
                    f_dst.write(f_src.read())
                    
                if system != "windows":
                    st = os.stat(bin_path)
                    os.chmod(bin_path, st.st_mode | stat.S_IEXEC)
                    
        except Exception as e:
            print(f"Failed to download another-meet from {url}")
            print(f"Error: {e}")
            sys.exit(1)
            
    # Execute the binary
    sys.exit(subprocess.call([bin_path] + sys.argv[1:]))

if __name__ == "__main__":
    main()
