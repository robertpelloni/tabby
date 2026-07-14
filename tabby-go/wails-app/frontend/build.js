#!/usr/bin/env node
// Manual frontend build script - bypasses hanging vite 3 on Node 24
const esbuild = require('esbuild');
const fs = require('fs');
const path = require('path');

const distDir = path.join(__dirname, 'dist');

async function build() {
  // Clean dist
  if (fs.existsSync(distDir)) {
    fs.rmSync(distDir, { recursive: true });
  }
  fs.mkdirSync(distDir, { recursive: true });
  fs.mkdirSync(path.join(distDir, 'assets'), { recursive: true });

  // Bundle JS - resolve all imports including wailsjs
  await esbuild.build({
    entryPoints: [path.join(__dirname, 'src/main.js')],
    bundle: true,
    outfile: path.join(distDir, 'assets/index.js'),
    format: 'esm',
    platform: 'browser',
    target: 'es2020',
    minify: false,
    sourcemap: false,
    loader: {
      '.css': 'copy',
      '.woff': 'file',
      '.woff2': 'file',
      '.ttf': 'file',
      '.png': 'file',
      '.svg': 'file',
    },
    external: [],
    alias: {
      // Alias the wailsjs runtime to its actual location
    },
  });

  // Copy CSS
  const cssSrc = path.join(__dirname, 'src/app.css');
  if (fs.existsSync(cssSrc)) {
    fs.copyFileSync(cssSrc, path.join(distDir, 'assets/index.css'));
  }

  // Copy and patch index.html
  let html = fs.readFileSync(path.join(__dirname, 'index.html'), 'utf8');
  // Replace the module script reference to use the bundled output
  html = html.replace(
    '<script src="./src/main.js" type="module"></script>',
    '<link rel="stylesheet" href="./assets/index.css"><script src="./assets/index.js" type="module"></script>'
  );
  fs.writeFileSync(path.join(distDir, 'index.html'), html);

  // Copy wailsjs runtime if needed (for any external references)
  const wailsjsSrc = path.join(__dirname, 'wailsjs');
  if (fs.existsSync(wailsjsSrc)) {
    fs.cpSync(wailsjsSrc, path.join(distDir, 'wailsjs'), { recursive: true });
  }

  console.log('Frontend build complete!');
  console.log('Files:', fs.readdirSync(distDir));
}

build().catch(e => {
  console.error('Build failed:', e);
  process.exit(1);
});
