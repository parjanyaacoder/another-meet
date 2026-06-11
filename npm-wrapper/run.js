#!/usr/bin/env node
const { spawnSync } = require('child_process');
const path = require('path');

const binPath = path.join(__dirname, 'bin', 'another-meet');
const result = spawnSync(binPath, process.argv.slice(2), { stdio: 'inherit' });

if (result.error) {
  console.error(`Failed to run another-meet: ${result.error.message}`);
  process.exit(1);
}

process.exit(result.status || 0);
