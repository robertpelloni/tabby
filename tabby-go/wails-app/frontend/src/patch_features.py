#!/usr/bin/env python3
"""Patch main.js to add Terminal Toolbar, Command Palette, and SSH Config Import."""
import re

with open('main.js', 'rb') as f:
    data = f.read()

content = data.decode('utf-8')
changes = 0

# 1. Add ImportSSHConfig to imports
old = 'SFTPChmod, SFTPReadlink, SFTPSymlink,'
new = 'SFTPChmod, SFTPReadlink, SFTPSymlink,\nImportSSHConfig,'
if old in content:
    content = content.replace(old, new)
    changes += 1
    print('1. ImportSSHConfig added to imports')

# 2. Add keyboard shortcut for Command Palette (Ctrl+Shift+P)
old = "if (ctrl && shift && e.key === 'S') { e.preventDefault(); openSerialDialog(); return; }"
new = "if (ctrl && shift && e.key === 'P') { e.preventDefault(); toggleCommandPalette(); return; }\n    if (ctrl && shift && e.key === 'S') { e.preventDefault(); openSerialDialog(); return; }"
if old in content:
    content = content.replace(old, new)
    changes += 1
    print('2. Command palette shortcut added')

# 3. Add button bindings for new features
old = "document.getElementById('sftp-close-btn').onclick = () => closeSFTPBrowser();"
new = "document.getElementById('sftp-close-btn').onclick = () => closeSFTPBrowser();\n    document.getElementById('btn-command-palette').onclick = () => toggleCommandPalette();\n    document.getElementById('btn-import-ssh-config').onclick = () => importSSHConfig();\n    document.getElementById('cmd-palette-input').oninput = () => filterCommandPalette();\n    document.getElementById('cmd-palette-input').onkeydown = (e) => handlePaletteKey(e);"
if old in content:
    content = content.replace(old, new)
    changes += 1
    print('3. Button bindings added')

# 4. Add toolbar on tab activate
old = "this.wrapper.classList.add('active');\n    this.tabEl.classList.add('active');\n    activeTabId = this.id;\n    this.term.focus();"
new = "this.wrapper.classList.add('active');\n    this.tabEl.classList.add('active');\n    activeTabId = this.id;\n    updateToolbar(this);\n    this.term.focus();"
if old in content:
    content = content.replace(old, new)
    changes += 1
    print('4. updateToolbar added to activate()')

# 5. Add toolbar mouse events on wrapper
old = "this.wrapper.className = 'terminal-wrapper';"
new = "this.wrapper.className = 'terminal-wrapper';\n    this.wrapper.onmouseenter = () => showToolbarForTab(this);\n    this.wrapper.onmouseleave = () => hideToolbarForTab(this);\n    this.pinToolbar = false;"
if old in content:
    content = content.replace(old, new)
    changes += 1
    print('5. Toolbar mouse events added')

# 6. Add command palette and SSH config import buttons to toolbar
m = re.search(r'id="btn-settings"\s+title="Settings', content)
if m:
    insert_point = m.start()
    new_btns = 'id="btn-command-palette" title="Command Palette (Ctrl+Shift+P)">&#9881;</button><button class="btn-icon" id="btn-import-ssh-config" title="Import SSH Config">&#128272;</button><button class="btn-icon" '
    content = content[:insert_point] + new_btns + content[insert_point:]
    changes += 1
    print('6. Command palette and SSH config buttons added to toolbar')
else:
    print('6. WARNING: btn-settings not found in buildUI')

# 7. Inject all new function code before PROFILES section
marker = '// ===== PROFILES ====='
if marker in content:
    new_funcs = """

// ===== TERMINAL TOOLBAR =====
function buildToolbar(tab) {
  let html = '<div class="terminal-toolbar" id="toolbar-' + tab.id + '">';
  if (tab.isSSH) {
    html += '<span class="toolbar-badge ssh">SSH</span>';
    html += '<span class="toolbar-info">' + escHtml(tab.title || 'ssh') + '</span>';
    html += '<button class="toolbar-btn" onclick="openSFTPBrowser(getActiveTab().sshConnectionId)" title="SFTP">\\U0001f4c2</button>';
    html += '<button class="toolbar-btn" onclick="openForwardDialog(getActiveTab().sshConnectionId)" title="Forward">\\U0001f504</button>';
  } else if (tab.isSerial) {
    html += '<span class="toolbar-badge serial">SER</span>';
    html += '<span class="toolbar-info">' + escHtml(tab.title || 'serial') + '</span>';
  } else if (tab.isTelnet) {
    html += '<span class="toolbar-badge telnet">TEL</span>';
    html += '<span class="toolbar-info">' + escHtml(tab.title || 'telnet') + '</span>';
  } else {
    html += '<span class="toolbar-badge local">LOCAL</span>';
    html += '<span class="toolbar-info">' + escHtml(tab.title || 'shell') + '</span>';
  }
  html += '<div class="toolbar-spacer"></div>';
  html += '<button class="toolbar-btn" onclick="getActiveTab().copySelection()" title="Copy">C</button>';
  html += '<button class="toolbar-btn" onclick="getActiveTab().pasteFromClipboard()" title="Paste">P</button>';
  html += '<button class="toolbar-btn" onclick="toggleFind()" title="Find">F</button>';
  html += '<button class="toolbar-btn" onclick="clearTerminal()" title="Clear">X</button>';
  html += '<button class="toolbar-btn toolbar-pin" onclick="toggleToolbarPin(\\''" + tab.id + "'\\')" title="Pin">Pin</button>';
  html += '</div>';
  return html;
}

function escHtml(s) {
  return s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;');
}

function updateToolbar(tab) {
  if (!tab || !tab.wrapper) return;
  let toolbar = tab.wrapper.querySelector('.terminal-toolbar');
  const div = document.createElement('div');
  div.innerHTML = buildToolbar(tab);
  const newToolbar = div.firstElementChild;
  if (toolbar) toolbar.replaceWith(newToolbar);
  else tab.wrapper.insertBefore(newToolbar, tab.wrapper.firstChild);
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
  if (tab) { tab.term.clear(); tab.term.focus(); }
}

function showToolbarForTab(tab) {
  if (!tab || !tab.wrapper || tab.pinToolbar) return;
  const toolbar = tab.wrapper.querySelector('.terminal-toolbar');
  if (toolbar) toolbar.classList.add('visible');
}

function hideToolbarForTab(tab) {
  if (!tab || !tab.wrapper || tab.pinToolbar) return;
  const toolbar = tab.wrapper.querySelector('.terminal-toolbar');
  if (toolbar) toolbar.classList.remove('visible');
}

// ===== COMMAND PALETTE =====
const PALETTE_COMMANDS = [
  { id: 'new-tab', label: 'New Tab', icon: '+', action: function() { newTab(); } },
  { id: 'close-tab', label: 'Close Tab', icon: 'x', action: function() { var t = getActiveTab(); if (t) t.close(); } },
  { id: 'ssh-connect', label: 'SSH Connect...', icon: '#', action: function() { openSSHDialog(); } },
  { id: 'serial-connect', label: 'Serial Port...', icon: '~', action: function() { openSerialDialog(); } },
  { id: 'telnet-connect', label: 'Telnet Connect...', icon: '=', action: function() { openTelnetDialog(); } },
  { id: 'import-ssh-config', label: 'Import SSH Config', icon: '@', action: function() { importSSHConfig(); } },
  { id: 'toggle-settings', label: 'Settings', icon: '*', action: function() { toggleSettings(); } },
  { id: 'split-horizontal', label: 'Split Horizontal', icon: '-', action: function() { splitPane('horizontal'); } },
  { id: 'split-vertical', label: 'Split Vertical', icon: '|', action: function() { splitPane('vertical'); } },
  { id: 'find', label: 'Find in Terminal', icon: '?', action: function() { toggleFind(); } },
  { id: 'clear-terminal', label: 'Clear Terminal', icon: '!', action: function() { clearTerminal(); } },
  { id: 'copy', label: 'Copy Selection', icon: 'C', action: function() { var t = getActiveTab(); if (t) t.copySelection(); } },
  { id: 'paste', label: 'Paste from Clipboard', icon: 'V', action: function() { var t = getActiveTab(); if (t) t.pasteFromClipboard(); } },
  { id: 'zoom-in', label: 'Increase Font Size', icon: '+', action: function() { fontSize = Math.min(32, fontSize + 1); tabs.forEach(function(t) { t.term.options.fontSize = fontSize; }); showToast('Font: ' + fontSize, 'info'); } },
  { id: 'zoom-out', label: 'Decrease Font Size', icon: '-', action: function() { fontSize = Math.max(8, fontSize - 1); tabs.forEach(function(t) { t.term.options.fontSize = fontSize; }); showToast('Font: ' + fontSize, 'info'); } },
  { id: 'zoom-reset', label: 'Reset Font Size', icon: '0', action: function() { fontSize = 14; tabs.forEach(function(t) { t.term.options.fontSize = fontSize; }); showToast('Font: 14', 'info'); } },
  { id: 'toggle-fullscreen', label: 'Toggle Fullscreen', icon: 'F', action: function() { if (document.fullscreenElement) document.exitFullscreen(); else document.documentElement.requestFullscreen(); } },
  { id: 'sftp-browser', label: 'SFTP File Browser', icon: 'F', action: function() { var t = getActiveTab(); if (t && t.isSSH) openSFTPBrowser(t.sshConnectionId); else showToast('Requires active SSH tab', 'info'); } },
  { id: 'port-forward', label: 'Port Forwarding', icon: 'P', action: function() { var t = getActiveTab(); if (t && t.isSSH) openForwardDialog(t.sshConnectionId); else showToast('Requires active SSH tab', 'info'); } },
  { id: 'save-session', label: 'Save Session State', icon: 'S', action: function() { saveSession(); } },
  { id: 'reload-colors', label: 'Reload Color Schemes', icon: 'R', action: function() { reloadColorSchemes(); } },
];

var paletteVisible = false;
var paletteSelectedIdx = 0;

function toggleCommandPalette() {
  paletteVisible = !paletteVisible;
  var el = document.getElementById('command-palette');
  if (paletteVisible) {
    el.classList.add('active');
    var input = document.getElementById('cmd-palette-input');
    input.value = '';
    input.focus();
    paletteSelectedIdx = 0;
    filterCommandPalette();
  } else {
    el.classList.remove('active');
    var tab = getActiveTab();
    if (tab) tab.term.focus();
  }
}

function filterCommandPalette() {
  var query = (document.getElementById('cmd-palette-input').value || '').toLowerCase();
  var container = document.getElementById('cmd-palette-items');
  var filtered = PALETTE_COMMANDS.filter(function(cmd) { return cmd.label.toLowerCase().includes(query); });
  paletteSelectedIdx = 0;
  container.innerHTML = filtered.map(function(cmd, i) {
    return '<div class="palette-item' + (i === 0 ? ' selected' : '') + '" data-id="' + cmd.id + '">' +
      '<span class="palette-icon">' + cmd.icon + '</span>' +
      '<span class="palette-label">' + cmd.label + '</span></div>';
  }).join('');
  container.querySelectorAll('.palette-item').forEach(function(el) {
    el.onclick = function() { executePaletteCommand(el.dataset.id); };
  });
}

function handlePaletteKey(e) {
  var items = document.querySelectorAll('.palette-item');
  if (e.key === 'Escape') { e.preventDefault(); toggleCommandPalette(); }
  else if (e.key === 'ArrowDown') { e.preventDefault(); paletteSelectedIdx = Math.min(paletteSelectedIdx + 1, items.length - 1); items.forEach(function(el, i) { el.classList.toggle('selected', i === paletteSelectedIdx); }); }
  else if (e.key === 'ArrowUp') { e.preventDefault(); paletteSelectedIdx = Math.max(paletteSelectedIdx - 1, 0); items.forEach(function(el, i) { el.classList.toggle('selected', i === paletteSelectedIdx); }); }
  else if (e.key === 'Enter') { e.preventDefault(); var sel = items[paletteSelectedIdx]; if (sel) sel.click(); }
}

function executePaletteCommand(id) {
  var cmd = PALETTE_COMMANDS.find(function(c) { return c.id === id; });
  if (cmd) { toggleCommandPalette(); cmd.action(); }
}

async function reloadColorSchemes() {
  try {
    const schemes = await GetColorSchemes();
    if (schemes && schemes.length) {
      schemes.forEach(function(s) { COLOR_SCHEMES[s.Name] = s; });
      schemeNames = schemes.map(function(s) { return s.Name; });
      showToast('Loaded ' + schemes.length + ' color schemes', 'success');
    }
  } catch (err) { showToast('Failed to reload schemes: ' + err, 'error'); }
}

// ===== SSH CONFIG IMPORT =====
async function importSSHConfig() {
  try {
    const result = await ImportSSHConfig();
    if (result && result.length > 0) {
      let imported = 0;
      let skipped = 0;
      for (var i = 0; i < result.length; i++) {
        var host = result[i];
        var opts = host.options || host.Options || {};
        var h = opts.host || opts.Host || '';
        var p = opts.port || opts.Port || 22;
        var u = opts.user || opts.User || 'root';
        var name = u + '@' + h;
        var exists = savedProfiles.find(function(pr) { return pr.name === name && pr.type === 'ssh'; });
        if (exists) { skipped++; continue; }
        savedProfiles.push({
          id: 'ssh-import-' + Date.now() + '-' + imported,
          type: 'ssh',
          name: name,
          options: { host: h, port: p, user: u, auth: ((opts.privateKeys && opts.privateKeys.length) ? 'publicKey' : 'agent'), privateKeys: opts.privateKeys || opts.PrivateKeys || [], jumpHost: opts.jumpHost || opts.JumpHost || '' },
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

"""
    content = content.replace(marker, new_funcs + '\n' + marker)
    changes += 1
    print('7. Terminal Toolbar, Command Palette, SSH Config Import functions injected')

# Write back in binary to avoid encoding issues
with open('main.js', 'wb') as f:
    f.write(content.encode('utf-8'))
print(f'\nTotal: {changes} changes applied.')
