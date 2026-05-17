#!/usr/bin/env python3
"""Patch main.js to add Terminal Toolbar, Command Palette, and SSH Config Import."""

with open('main.js', 'r', encoding='utf-8') as f:
    content = f.read()

changes = 0

# 1. Add keyboard shortcut for Command Palette (Ctrl+Shift+P)
old = "if (ctrl && shift && e.key === 'S') { e.preventDefault(); openSerialDialog(); return; }"
new = """if (ctrl && shift && e.key === 'P') { e.preventDefault(); toggleCommandPalette(); return; }
    if (ctrl && shift && e.key === 'S') { e.preventDefault(); openSerialDialog(); return; }"""
if old in content:
    content = content.replace(old, new)
    changes += 1

# 2. Add toolbar button and command palette button bindings
old = "document.getElementById('sftp-close-btn').onclick = () => closeSFTPBrowser();"
new = """document.getElementById('sftp-close-btn').onclick = () => closeSFTPBrowser();
    document.getElementById('btn-command-palette').onclick = () => toggleCommandPalette();
    document.getElementById('btn-import-ssh-config').onclick = () => importSSHConfig();
    document.getElementById('cmd-palette-input').oninput = () => filterCommandPalette();
    document.getElementById('cmd-palette-input').onkeydown = (e) => handlePaletteKey(e);"""
if old in content:
    content = content.replace(old, new)
    changes += 1

# 3. Add Toolbar and Command Palette functions before PROFILES section
marker = '// ===== PROFILES ====='
if marker in content:
    new_funcs = '''

// ===== TERMINAL TOOLBAR =====
function buildToolbar(tab) {
  let html = '<div class="terminal-toolbar" id="toolbar-' + tab.id + '">';
  // Connection type indicator
  if (tab.isSSH) {
    html += '<span class="toolbar-badge ssh" title="SSH Connection">SSH</span>';
    html += '<span class="toolbar-info">' + (tab.title || 'ssh') + '</span>';
    html += '<button class="toolbar-btn" onclick="openSFTPBrowser(getActiveTab().sshConnectionId)" title="SFTP Browser">\\ud83d\\udcc2</button>';
    html += '<button class="toolbar-btn" onclick="openForwardDialog(getActiveTab().sshConnectionId)" title="Port Forwarding">\\ud83d\\udd00</button>';
  } else if (tab.isSerial) {
    html += '<span class="toolbar-badge serial" title="Serial Port">SER</span>';
    html += '<span class="toolbar-info">' + (tab.title || 'serial') + '</span>';
  } else if (tab.isTelnet) {
    html += '<span class="toolbar-badge telnet" title="Telnet">TEL</span>';
    html += '<span class="toolbar-info">' + (tab.title || 'telnet') + '</span>';
  } else {
    html += '<span class="toolbar-badge local" title="Local Shell">LOCAL</span>';
    html += '<span class="toolbar-info">' + (tab.title || 'shell') + '</span>';
  }
  // Right-side actions
  html += '<div class="toolbar-spacer"></div>';
  html += '<button class="toolbar-btn" onclick="getActiveTab().copySelection()" title="Copy Selection">\\ud83d\\udccb</button>';
  html += '<button class="toolbar-btn" onclick="getActiveTab().pasteFromClipboard()" title="Paste">\\ud83d\\udcc4</button>';
  html += '<button class="toolbar-btn" onclick="toggleFind()" title="Find (Ctrl+F)">\\ud83d\\udd0d</button>';
  html += '<button class="toolbar-btn" onclick="clearTerminal()" title="Clear Terminal">\\ud83d\\udd0c</button>';
  html += '<button class="toolbar-btn toolbar-pin" onclick="toggleToolbarPin(\\'' + tab.id + '\\')" title="Pin Toolbar">\\ud83d\\udccd</button>';
  html += '</div>';
  return html;
}

function updateToolbar(tab) {
  if (!tab || !tab.wrapper) return;
  let toolbar = tab.wrapper.querySelector('.terminal-toolbar');
  if (!toolbar) {
    // Insert toolbar at top of terminal wrapper
    const div = document.createElement('div');
    div.innerHTML = buildToolbar(tab);
    toolbar = div.firstElementChild;
    tab.wrapper.insertBefore(toolbar, tab.wrapper.firstChild);
  } else {
    const div = document.createElement('div');
    div.innerHTML = buildToolbar(tab);
    const newToolbar = div.firstElementChild;
    toolbar.replaceWith(newToolbar);
  }
}

function toggleToolbarPin(tabId) {
  const tab = tabs.find(t => t.id === tabId);
  if (!tab) return;
  tab.pinToolbar = !tab.pinToolbar;
  const toolbar = tab.wrapper.querySelector('.terminal-toolbar');
  if (toolbar) {
    toolbar.classList.toggle('pinned', tab.pinToolbar);
    toolbar.classList.toggle('unpinned', !tab.pinToolbar);
  }
}

function clearTerminal() {
  const tab = getActiveTab();
  if (tab) {
    tab.term.clear();
    tab.term.focus();
  }
}

function showToolbarForTab(tab) {
  if (!tab || !tab.wrapper) return;
  const toolbar = tab.wrapper.querySelector('.terminal-toolbar');
  if (toolbar && !tab.pinToolbar) {
    toolbar.classList.add('visible');
  }
}

function hideToolbarForTab(tab) {
  if (!tab || !tab.wrapper) return;
  const toolbar = tab.wrapper.querySelector('.terminal-toolbar');
  if (toolbar && !tab.pinToolbar) {
    toolbar.classList.remove('visible');
  }
}

// ===== COMMAND PALETTE =====
const PALETTE_COMMANDS = [
  { id: 'new-tab', label: 'New Tab', icon: '\\u2795', action: () => newTab() },
  { id: 'close-tab', label: 'Close Tab', icon: '\\u274c', action: () => { const t = getActiveTab(); if (t) t.close(); } },
  { id: 'ssh-connect', label: 'SSH Connect...', icon: '\\ud83d\\udd10', action: () => openSSHDialog() },
  { id: 'serial-connect', label: 'Serial Port...', icon: '\\ud83d\\udce1', action: () => openSerialDialog() },
  { id: 'telnet-connect', label: 'Telnet Connect...', icon: '\\ud83c\\udf10', action: () => openTelnetDialog() },
  { id: 'import-ssh-config', label: 'Import SSH Config', icon: '\\ud83d\\udcc2', action: () => importSSHConfig() },
  { id: 'toggle-settings', label: 'Settings', icon: '\\u2699', action: () => toggleSettings() },
  { id: 'split-horizontal', label: 'Split Horizontal', icon: '\\u2194', action: () => splitPane('horizontal') },
  { id: 'split-vertical', label: 'Split Vertical', icon: '\\u2195', action: () => splitPane('vertical') },
  { id: 'find', label: 'Find in Terminal', icon: '\\ud83d\\udd0d', action: () => toggleFind() },
  { id: 'clear-terminal', label: 'Clear Terminal', icon: '\\ud83d\\udd0c', action: () => clearTerminal() },
  { id: 'copy', label: 'Copy Selection', icon: '\\ud83d\\udccb', action: () => { const t = getActiveTab(); if (t) t.copySelection(); } },
  { id: 'paste', label: 'Paste from Clipboard', icon: '\\ud83d\\udcc4', action: () => { const t = getActiveTab(); if (t) t.pasteFromClipboard(); } },
  { id: 'zoom-in', label: 'Increase Font Size', icon: '\\ud83d\\udd0e+', action: () => { fontSize = Math.min(32, fontSize + 1); tabs.forEach(t => t.term.options.fontSize = fontSize); showToast('Font: ' + fontSize, 'info'); } },
  { id: 'zoom-out', label: 'Decrease Font Size', icon: '\\ud83d\\udd0e-', action: () => { fontSize = Math.max(8, fontSize - 1); tabs.forEach(t => t.term.options.fontSize = fontSize); showToast('Font: ' + fontSize, 'info'); } },
  { id: 'zoom-reset', label: 'Reset Font Size', icon: '\\ud83d\\udd0e0', action: () => { fontSize = 14; tabs.forEach(t => t.term.options.fontSize = fontSize); showToast('Font: 14', 'info'); } },
  { id: 'toggle-fullscreen', label: 'Toggle Fullscreen', icon: '\\u26f6', action: () => { document.documentElement.requestFullscreen?.() || document.exitFullscreen?.(); } },
  { id: 'sftp-browser', label: 'SFTP File Browser', icon: '\\ud83d\\udcc1', action: () => { const t = getActiveTab(); if (t && t.isSSH) openSFTPBrowser(t.sshConnectionId); else showToast('Requires active SSH tab', 'info'); } },
  { id: 'port-forward', label: 'Port Forwarding', icon: '\\ud83d\\udd00', action: () => { const t = getActiveTab(); if (t && t.isSSH) openForwardDialog(t.sshConnectionId); else showToast('Requires active SSH tab', 'info'); } },
  { id: 'save-session', label: 'Save Session State', icon: '\\ud83d\\udcbe', action: () => saveSession() },
  { id: 'reload-colors', label: 'Reload Color Schemes', icon: '\\ud83c\\udfa8', action: () => reloadColorSchemes() },
];

let paletteVisible = false;
let paletteSelectedIdx = 0;

function toggleCommandPalette() {
  paletteVisible = !paletteVisible;
  const el = document.getElementById('command-palette');
  if (paletteVisible) {
    el.classList.add('active');
    const input = document.getElementById('cmd-palette-input');
    input.value = '';
    input.focus();
    paletteSelectedIdx = 0;
    filterCommandPalette();
  } else {
    el.classList.remove('active');
    const tab = getActiveTab();
    if (tab) tab.term.focus();
  }
}

function filterCommandPalette() {
  const query = (document.getElementById('cmd-palette-input').value || '').toLowerCase();
  const container = document.getElementById('cmd-palette-items');
  const filtered = PALETTE_COMMANDS.filter(cmd => cmd.label.toLowerCase().includes(query));
  paletteSelectedIdx = 0;
  container.innerHTML = filtered.map((cmd, i) =>
    '<div class="palette-item' + (i === 0 ? ' selected' : '') + '" data-idx="' + i + '" onclick="executePaletteCommand(\\'' + cmd.id + '\\')">' +
    '<span class="palette-icon">' + cmd.icon + '</span>' +
    '<span class="palette-label">' + cmd.label + '</span></div>'
  ).join('');
}

function handlePaletteKey(e) {
  const items = document.querySelectorAll('.palette-item');
  if (e.key === 'Escape') {
    e.preventDefault();
    toggleCommandPalette();
  } else if (e.key === 'ArrowDown') {
    e.preventDefault();
    paletteSelectedIdx = Math.min(paletteSelectedIdx + 1, items.length - 1);
    items.forEach((el, i) => el.classList.toggle('selected', i === paletteSelectedIdx));
  } else if (e.key === 'ArrowUp') {
    e.preventDefault();
    paletteSelectedIdx = Math.max(paletteSelectedIdx - 1, 0);
    items.forEach((el, i) => el.classList.toggle('selected', i === paletteSelectedIdx));
  } else if (e.key === 'Enter') {
    e.preventDefault();
    const selected = items[paletteSelectedIdx];
    if (selected) selected.click();
  }
}

function executePaletteCommand(id) {
  const cmd = PALETTE_COMMANDS.find(c => c.id === id);
  if (cmd) {
    toggleCommandPalette();
    cmd.action();
  }
}

async function reloadColorSchemes() {
  try {
    const schemes = await GetColorSchemes();
    if (schemes && schemes.length) {
      schemes.forEach(s => { COLOR_SCHEMES[s.Name] = s; });
      schemeNames = schemes.map(s => s.Name);
      showToast('Loaded ' + schemes.length + ' color schemes', 'success');
    }
  } catch (err) {
    showToast('Failed to reload schemes: ' + err, 'error');
  }
}

// ===== SSH CONFIG IMPORT =====
async function importSSHConfig() {
  try {
    const result = await ImportSSHConfig();
    if (result && result.length > 0) {
      let imported = 0;
      let skipped = 0;
      for (const host of result) {
        const name = (host.user || 'user') + '@' + (host.host || 'unknown');
        const exists = savedProfiles.find(p => p.name === name && p.type === 'ssh');
        if (exists) { skipped++; continue; }
        savedProfiles.push({
          id: 'ssh-import-' + Date.now() + '-' + imported,
          type: 'ssh',
          name: name,
          options: {
            host: host.host,
            port: host.port || 22,
            user: host.user || 'root',
            auth: host.identityFile ? 'publicKey' : 'agent',
            privateKeys: host.identityFile ? [host.identityFile] : [],
            jumpHost: host.proxyJump || '',
          },
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        });
        imported++;
      }
      await SaveProfiles(savedProfiles);
      renderProfiles();
      showToast('Imported ' + imported + ' hosts' + (skipped ? ', skipped ' + skipped + ' duplicates' : ''), 'success');
    } else {
      showToast('No SSH hosts found in config', 'info');
    }
  } catch (err) {
    showToast('SSH config import failed: ' + err, 'error');
  }
}

'''
    content = content.replace(marker, new_funcs + '\n' + marker)
    changes += 1

# 4. Update Tab constructor to add toolbar on creation
old_activate = """this.wrapper.classList.add('active');
    this.tabEl.classList.add('active');
    activeTabId = this.id;
    this.term.focus();"""
new_activate = """this.wrapper.classList.add('active');
    this.tabEl.classList.add('active');
    activeTabId = this.id;
    updateToolbar(this);
    this.term.focus();"""
if old_activate in content:
    content = content.replace(old_activate, new_activate)
    changes += 1

# 5. Add toolbar mouse enter/leave on wrapper
old_wrapper = "this.wrapper.className = 'terminal-wrapper';"
new_wrapper = """this.wrapper.className = 'terminal-wrapper';
    this.wrapper.onmouseenter = () => showToolbarForTab(this);
    this.wrapper.onmouseleave = () => hideToolbarForTab(this);
    this.pinToolbar = false;"""
if old_wrapper in content:
    content = content.replace(old_wrapper, new_wrapper)
    changes += 1

with open('main.js', 'w', encoding='utf-8') as f:
    f.write(content)

print(f'main.js patched! {changes} changes applied.')
