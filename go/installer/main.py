import json
import urllib.request
from pathlib import Path

REPO = "layer8co/VOMic"

# The binary name which resides in the client's system.
BINARY = "vomic"

release_url = f"https://api.github.com/repos/{REPO}/releases/tags/nightly"

print("Going to install vomic nightly.")
with open("/dev/tty") as tty:
    print("Proceed? Press Enter to continue")
    tty.readline()

with urllib.request.urlopen(release_url) as response:
    release = json.load(response)

asset = next(
    a for a in release["assets"]
    if a["name"] == "vomic-linux-amd64"
)

install_dir = Path.home() / ".local/bin"
install_dir.mkdir(parents=True, exist_ok=True)

destination = install_dir / BINARY

print(f"Downloading asset: {asset["browser_download_url"]}")
urllib.request.urlretrieve(asset["browser_download_url"], destination)
destination.chmod(0o755)

print(f"Installed {BINARY} to {destination}")
