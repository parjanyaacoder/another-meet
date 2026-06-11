const os = require('os');
const fs = require('fs');
const path = require('path');
const https = require('https');
const tar = require('tar');
const unzipper = require('unzipper');

const VERSION = '1.0.3';
const REPO = 'parjanyaacoder/another-meet';

const platformMap = {
  darwin: 'darwin',
  linux: 'linux',
  win32: 'windows'
};

const archMap = {
  x64: 'amd64',
  arm64: 'arm64'
};

const platform = platformMap[os.platform()];
const arch = archMap[os.arch()];

if (!platform || !arch) {
  console.error(`Unsupported platform or architecture: ${os.platform()} ${os.arch()}`);
  process.exit(1);
}

const ext = platform === 'windows' ? 'zip' : 'tar.gz';
const filename = `another-meet_${VERSION}_${platform}_${arch}.${ext}`;
const url = `https://github.com/${REPO}/releases/download/v${VERSION}/${filename}`;

const binDir = path.join(__dirname, 'bin');
if (!fs.existsSync(binDir)) {
  fs.mkdirSync(binDir);
}

const binPath = path.join(binDir, platform === 'windows' ? 'another-meet.exe' : 'another-meet');

console.log(`Downloading ${url}...`);

function download(url) {
  https.get(url, (res) => {
    if (res.statusCode === 301 || res.statusCode === 302) {
      download(res.headers.location);
    } else if (res.statusCode === 200) {
      handleResponse(res);
    } else {
      console.error(`Failed to download: HTTP ${res.statusCode}`);
      process.exit(1);
    }
  }).on('error', (err) => {
    console.error(`Error: ${err.message}`);
    process.exit(1);
  });
}

function handleResponse(res) {
  if (ext === 'zip') {
    res.pipe(unzipper.Extract({ path: binDir })).on('close', finish);
  } else {
    res.pipe(tar.x({ C: binDir })).on('close', finish);
  }
}

function finish() {
  if (platform !== 'windows') {
    fs.chmodSync(binPath, 0o755);
  }
  console.log('Installation complete!');
}

download(url);
