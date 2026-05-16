import re, sys

with open('tabby-go/wails-app/frontend/src/main.js', 'r', encoding='utf-8') as f:
    c = f.read()

# 1. Add SSH button
c = c.replace(
    '<button class="btn-icon" id="btn-new-tab" title="New Tab (Ctrl+Shift+T)">+</button>\n            <button class="btn-icon" id="btn-settings"',
    '<button class="btn-icon" id="btn-new-tab" title="New Tab (Ctrl+Shift+T)">+</button>\n            <button class="btn-icon" id="btn-ssh" title="New SSH Connection">\U0001f510</button>\n            <button class="btn-icon" id="btn-settings"'
)

# 2. Add color scheme dropdown
c = c.replace(
    '<div class="settings-page active" id="settings-appearance">\n                <h3>Font</h3>',
    '<div class="settings-page active" id="settings-appearance">\n                <h3>Color Scheme</h3>\n                <div class="setting-group"><label>Terminal Color Scheme</label><select id="s-color-scheme"></select></div>\n                <div id="color-scheme-preview" style="display:flex;flex-wrap:wrap;gap:3px;padding:8px 0;"></div>\n                <h3>Font</h3>'
)

# 3. Add profiles tab
c = c.replace(
    '<button class="settings-tab" data-tab="serial">\U0001f4e1 Serial</button>',
    '<button class="settings-tab" data-tab="serial">\U0001f4e1 Serial</button>\n            <button class="settings-tab" data-tab="profiles">\U0001f4c1 Profiles</button>'
)

# 4. Add profiles page
c = c.replace(
    '        <div id="settings-actions">',
    '            <div class="settings-page" id="settings-profiles">\n                <h3>Saved Profiles</h3>\n                <div id="profiles-editor-list" style="margin-bottom:12px;"></div>\n                <button class="btn-primary" id="btn-add-profile" style="width:100%;margin-bottom:8px;">+ Add SSH Profile</button>\n            </div>\n        </div>\n        <div id="settings-actions">'
)

# 5. Add SSH dialog modal
c = c.replace(
    '    </div>\n\n    // Populate shell dropdown',
    '''    </div>
    <!-- SSH Connection Dialog -->
    <div id="ssh-dialog" class="modal-overlay">
        <div class="modal-box">
            <h3>\U0001f510 New SSH Connection</h3>
            <div class="setting-group"><label>Host</label><input type="text" id="ssh-host" placeholder="example.com"></div>
            <div class="setting-group"><label>Port</label><input type="number" id="ssh-port" value="22" min="1" max="65535"></div>
            <div class="setting-group"><label>Username</label><input type="text" id="ssh-user" placeholder="root"></div>
            <div class="setting-group"><label>Authentication</label>
                <select id="ssh-auth"><option value="agent">SSH Agent</option><option value="publicKey">Public Key</option><option value="password">Password</option><option value="keyboardInteractive">Keyboard Interactive</option></select>
            </div>
            <div class="setting-group" id="ssh-password-group" style="display:none;"><label>Password</label><input type="password" id="ssh-password" placeholder="Password"></div>
            <div class="setting-group" id="ssh-key-group" style="display:none;"><label>Private Key Path</label><input type="text" id="ssh-key-path" placeholder="~/.ssh/id_ed25519"></div>
            <div class="setting-group"><label>Save as Profile</label><div class="toggle-container"><input type="checkbox" id="ssh-save-profile"><label for="ssh-save-profile" class="toggle-label"></label></div></div>
            <div class="modal-actions">
                <button class="btn-secondary" id="ssh-cancel">Cancel</button>
                <button class="btn-primary" id="ssh-connect">Connect</button>
            </div>
        </div>
    </div>

    // Populate shell dropdown'''
)

# 6. Add profiles section to sidebar
c = c.replace(
    '        <div id="sidebar-footer">',
    '        <div id="profiles-section" style="border-top:1px solid #2b2b2b;padding:8px 0;max-height:200px;overflow-y:auto;display:none;">\n            <div style="padding:0 12px 4px;font-size:10px;color:#666;text-transform:uppercase;letter-spacing:0.5px;">Profiles</div>\n            <div id="profiles-list"></div>\n        </div>\n        <div id="sidebar-footer">'
)

# 7. Add SSH button binding
c = c.replace(
    "document.getElementById('btn-settings').onclick = () => toggleSettings();",
    "document.getElementById('btn-ssh').onclick = () => openSSHDialog();\n    document.getElementById('btn-settings').onclick = () => toggleSettings();"
)

# 8. Add SSH auth toggle
c = c.replace(
    '    // Button bindings\n    document.getElementById',
    "    // SSH auth toggle\n    document.getElementById('ssh-auth').onchange = (e) => { document.getElementById('ssh-password-group').style.display = e.target.value === 'password' ? 'block' : 'none'; document.getElementById('ssh-key-group').style.display = e.target.value === 'publicKey' ? 'block' : 'none'; };\n\n    // Button bindings\n    document.getElementById"
)

# 9. Add SSH and profiles button bindings
c = c.replace(
    "document.getElementById('btn-reset').onclick = () => doResetSettings();",
    "document.getElementById('ssh-cancel').onclick = () => closeSSHDialog();\n    document.getElementById('ssh-connect').onclick = () => doSSHConnect();\n    document.getElementById('btn-add-profile').onclick = () => addProfile();\n    document.getElementById('btn-reset').onclick = () => doResetSettings();"
)

# 10. Add color scheme dropdown population
c = c.replace(
    '    // Populate shell dropdown',
    "    // Populate color scheme dropdown\n    const schemeSelect = document.getElementById('s-color-scheme');\n    schemeNames.forEach(name => { const opt = document.createElement('option'); opt.value = name; opt.textContent = name; schemeSelect.appendChild(opt); });\n    schemeSelect.onchange = () => applyColorScheme(schemeSelect.value);\n\n    // Populate shell dropdown"
)

# 11. Add renderProfiles call
c = c.replace(
    '    // Settings tab navigation',
    '    // Render profiles\n    renderProfiles();\n\n    // Settings tab navigation'
)

# 12. Color scheme in applySettingsToUI
c = c.replace(
    "    // Appearance\n    set('s-font-family', s.FontFamily);",
    "    // Appearance\n    set('s-color-scheme', s.ColorScheme || 'Tabby Default');\n    if (s.ColorScheme) applyColorScheme(s.ColorScheme);\n    set('s-font-family', s.FontFamily);"
)

# 13. ColorScheme in saveSettingsFromUI
c = c.replace(
    "    const s = {\n        FontFamily:",
    "    const s = {\n        ColorScheme: document.getElementById('s-color-scheme').value || 'Tabby Default',\n        FontFamily:"
)

# 14. Replace static themes with fallback
c = c.replace(
    "const DARK_THEME = {\n    background: '#1e1e1e', foreground: '#cccccc', cursor: '#aeafad', selectionBackground: '#264f78',\n    black: '#1e1e1e', red: '#f44747', green: '#6a9955', yellow: '#d7ba7d',\n    blue: '#569cd6', magenta: '#c586c0', cyan: '#4ec9b0', white: '#cccccc',\n    brightBlack: '#666666', brightRed: '#f44747', brightGreen: '#6a9955',\n    brightYellow: '#d7ba7d', brightBlue: '#569cd6', brightMagenta: '#c586c0',\n    brightCyan: '#4ec9b0', brightWhite: '#e0e0e0',\n};\nconst LIGHT_THEME = {\n    background: '#ffffff', foreground: '#333333', cursor: '#333333', selectionBackground: '#add6ff',\n    black: '#000000', red: '#cd3131', green: '#00bc00', yellow: '#949800',\n    blue: '#0451a5', magenta: '#bc05bc', cyan: '#0598bc', white: '#555555',\n    brightBlack: '#666666', brightRed: '#cd3131', brightGreen: '#00bc00',\n    brightYellow: '#949800', brightBlue: '#0451a5', brightMagenta: '#bc05bc',\n    brightCyan: '#0598bc', brightWhite: '#a0a0a0',\n};",
    "const FALLBACK_DARK = { background: '#1e1e1e', foreground: '#cccccc', cursor: '#aeafad', selectionBackground: '#264f78', black: '#1e1e1e', red: '#f44747', green: '#6a9955', yellow: '#d7ba7d', blue: '#569cd6', magenta: '#c586c0', cyan: '#4ec9b0', white: '#cccccc', brightBlack: '#666666', brightRed: '#f44747', brightGreen: '#6a9955', brightYellow: '#d7ba7d', brightBlue: '#569cd6', brightMagenta: '#c586c0', brightCyan: '#4ec9b0', brightWhite: '#e0e0e0' };"
)

# 15. Update theme references in Tab constructor
c = c.replace(
    "const theme = settings.Theme || 'dark';",
    "const colorScheme = settings.ColorScheme || 'Tabby Default'; const theme = getColorSchemeTheme(colorScheme);"
)
c = c.replace(
    "theme: theme === 'light' ? LIGHT_THEME : DARK_THEME,",
    "theme: theme || FALLBACK_DARK,"
)

# 16. Update applyTheme
c = c.replace(
    "const isDark = theme !== 'light';",
    "const isDark = !isSchemeLight(settings.ColorScheme || 'Tabby Default');"
)
c = c.replace(
    "tabs.forEach(t => {\n        if (t.term) t.term.options.theme = isDark ? DARK_THEME : LIGHT_THEME;\n    });",
    "const schemeTheme = getColorSchemeTheme(settings.ColorScheme || 'Tabby Default'); tabs.forEach(t => { if (t.term && schemeTheme) t.term.options.theme = schemeTheme; });"
)

# 17. Add SSH event listeners
c = c.replace(
    "EventsOn('pty.data', (params) => { (window.__ptyDataHandlers || []).forEach(h => h(params)); });\nEventsOn('pty.exit', (params) => { (window.__ptyExitHandlers || []).forEach(h => h(params)); });",
    "EventsOn('pty.data', (params) => { (window.__ptyDataHandlers || []).forEach(h => h(params)); });\nEventsOn('pty.exit', (params) => { (window.__ptyExitHandlers || []).forEach(h => h(params)); });\nEventsOn('ssh.data', (params) => { (window.__ptyDataHandlers || []).forEach(h => h(params)); });\nEventsOn('ssh.exit', (params) => { (window.__ptyExitHandlers || []).forEach(h => h(params)); });"
)

# 18. Add Tab.isSSH support
c = c.replace(
    "this.ptyId = null; this.title = 'Shell'; this.shell = shell || defaultShell; this.exited = false;",
    "this.ptyId = null; this.title = 'Shell'; this.shell = shell || defaultShell; this.exited = false; this.isSSH = false; this.sshConnectionId = null; this.sshSessionId = null;"
)
c = c.replace(
    "if ((params.ptyId ?? params.PTYID) === this.ptyId) this.term.write(atob(params.data));",
    "if ((params.ptyId ?? params.PTYID) === this.ptyId) this.term.write(atob(params.data)); if (this.isSSH && (params.sessionId ?? params.SessionID) === this.sshSessionId) this.term.write(atob(params.data));"
)
c = c.replace(
    "if (this.ptyId) PTYKill(this.ptyId, '').catch(() => {});",
    "if (this.ptyId) PTYKill(this.ptyId, '').catch(() => {}); if (this.isSSH && this.sshConnectionId) SSHClose({ connectionId: this.sshConnectionId }).catch(() => {});"
)
c = c.replace(
    "if (this.ptyId && !this.exited) PTYResize(this.ptyId, this.term.cols, this.term.rows);\n        });",
    "if (this.ptyId && !this.exited) PTYResize(this.ptyId, this.term.cols, this.term.rows); if (this.isSSH && this.sshConnectionId && this.sshSessionId) SSHResize({ connectionId: this.sshConnectionId, sessionId: this.sshSessionId, columns: this.term.cols, rows: this.term.rows });\n        });"
)
c = c.replace(
    "async pasteFromClipboard() { try { const text = await navigator.clipboard.readText(); if (text && this.ptyId && !this.exited) PTYWrite(this.ptyId, btoa(text)); } catch (_) { showToast('Clipboard access denied', 'error'); } }",
    "async pasteFromClipboard() { try { const text = await navigator.clipboard.readText(); if (text) { if (this.isSSH && this.sshConnectionId && this.sshSessionId) SSHWrite({ connectionId: this.sshConnectionId, sessionId: this.sshSessionId, data: btoa(text) }); else if (this.ptyId && !this.exited) PTYWrite(this.ptyId, btoa(text)); } } catch (_) { showToast('Clipboard access denied', 'error'); } }"
)

# 19. Add helper functions before BUILD UI
c = c.replace(
    '// ===== BUILD UI =====',
    '''// ===== COLOR SCHEME HELPERS =====
function getColorSchemeTheme(name) { const scheme = COLOR_SCHEMES[name]; if (!scheme) return null; const c = scheme.Colors || []; return { background: scheme.Background, foreground: scheme.Foreground, cursor: scheme.Cursor, cursorAccent: scheme.CursorAccent || undefined, selectionBackground: scheme.Selection || undefined, selectionForeground: scheme.SelectionForeground || undefined, black: c[0], red: c[1], green: c[2], yellow: c[3], blue: c[4], magenta: c[5], cyan: c[6], white: c[7], brightBlack: c[8], brightRed: c[9], brightGreen: c[10], brightYellow: c[11], brightBlue: c[12], brightMagenta: c[13], brightCyan: c[14], brightWhite: c[15] }; }
function isSchemeLight(name) { const bg = (COLOR_SCHEMES[name] && COLOR_SCHEMES[name].Background) || '#171717'; return isLightColor(bg); }
function isLightColor(hex) { if (!hex || hex.length < 7 || hex[0] !== '#') return false; const r = parseInt(hex.slice(1,3),16), g = parseInt(hex.slice(3,5),16), b = parseInt(hex.slice(5,7),16); return (0.299*r + 0.587*g + 0.114*b) / 255 > 0.5; }
function applyColorScheme(name) { const theme = getColorSchemeTheme(name); if (!theme) return; tabs.forEach(t => { if (t.term) t.term.options.theme = theme; }); if (isSchemeLight(name)) { document.body.classList.add('light-theme'); document.body.classList.remove('dark-theme'); } else { document.body.classList.add('dark-theme'); document.body.classList.remove('light-theme'); } renderColorSchemePreview(name); }
function renderColorSchemePreview(name) { const container = document.getElementById('color-scheme-preview'); if (!container) return; const scheme = COLOR_SCHEMES[name]; if (!scheme) { container.innerHTML = ''; return; } const c = scheme.Colors || []; const all = [scheme.Background, scheme.Foreground, scheme.Cursor, ...c]; container.innerHTML = all.map(color => `<div style="width:16px;height:16px;border-radius:3px;background:${color};border:1px solid #3a3a3a;" title="${color}"></div>`).join(''); }
// ===== SSH DIALOG =====
function openSSHDialog() { document.getElementById('ssh-dialog').classList.add('active'); document.getElementById('ssh-host').focus(); }
function closeSSHDialog() { document.getElementById('ssh-dialog').classList.remove('active'); const t = getActiveTab(); if (t) t.term.focus(); }
async function doSSHConnect() { const host = document.getElementById('ssh-host').value.trim(); const port = parseInt(document.getElementById('ssh-port').value) || 22; const user = document.getElementById('ssh-user').value.trim(); const auth = document.getElementById('ssh-auth').value; if (!host) { showToast('Host is required', 'error'); return; } closeSSHDialog(); showStatus('Connecting to ' + host + '...'); const authParams = { type: auth }; if (auth === 'password') authParams.password = document.getElementById('ssh-password').value; if (auth === 'publicKey') authParams.privateKeyPaths = [document.getElementById('ssh-key-path').value || '~/.ssh/id_ed25519']; try { const result = await SSHConnect({ host, port, user, auth: authParams, keepaliveInterval: 30, keepaliveCountMax: 3, readyTimeout: 15000 }); showToast('Connected to ' + host, 'success'); const tab = new Tab(defaultShell, 'ssh://' + user + '@' + host); tabs.push(tab); tab.activate(); tab.ptyId = null; tab.sshConnectionId = result.connectionId; tab.setTitle(user + '@' + host); tab.tabEl.querySelector('.tab-icon').textContent = '\\U0001f510'; const shellResult = await SSHStartShell({ connectionId: result.connectionId, columns: tab.term.cols, rows: tab.term.rows, terminal: 'xterm-256color' }); tab.sshSessionId = shellResult.sessionId; tab.isSSH = true; tab.term.onData((data) => { if (tab.sshConnectionId && tab.sshSessionId) SSHWrite({ connectionId: tab.sshConnectionId, sessionId: tab.sshSessionId, data: btoa(data) }); }); showStatus('SSH - ' + user + '@' + host); if (document.getElementById('ssh-save-profile') && document.getElementById('ssh-save-profile').checked) { savedProfiles.push({ id: 'ssh-' + Date.now(), type: 'ssh', name: user + '@' + host, options: { host, port, user, auth, privateKeys: authParams.privateKeyPaths || [] }, createdAt: new Date().toISOString(), updatedAt: new Date().toISOString() }); SaveProfiles(savedProfiles).catch(() => {}); renderProfiles(); } } catch (err) { showToast('SSH failed: ' + err, 'error'); showStatus('SSH failed - ' + host); } }
// ===== PROFILES =====
function renderProfiles() { const section = document.getElementById('profiles-section'); const list = document.getElementById('profiles-list'); const editor = document.getElementById('profiles-editor-list'); if (!savedProfiles || savedProfiles.length === 0) { if (section) section.style.display = 'none'; if (editor) editor.innerHTML = '<div style="color:#666;font-size:12px;">No saved profiles yet.</div>'; return; } if (section) section.style.display = 'block'; if (list) { list.innerHTML = savedProfiles.map(p => { const icon = p.type === 'ssh' ? '\\U0001f510' : p.type === 'serial' ? '\\U0001f4e1' : '\\u2318'; return `<div class="profile-item" data-profile-id="${p.id}" title="${p.name}"><span class="profile-icon">${icon}</span><span class="profile-name">${p.name}</span></div>`; }).join(''); list.querySelectorAll('.profile-item').forEach(el => { el.onclick = () => { const profile = savedProfiles.find(p => p.id === el.dataset.profileId); if (profile) connectProfile(profile); }; }); } if (editor) { editor.innerHTML = savedProfiles.map(p => { const icon = p.type === 'ssh' ? '\\U0001f510' : p.type === 'serial' ? '\\U0001f4e1' : '\\u2318'; return `<div class="profile-editor-item"><span>${icon} ${p.name}</span><button class="btn-icon profile-delete" data-id="${p.id}" title="Delete">\\u00d7</button></div>`; }).join(''); editor.querySelectorAll('.profile-delete').forEach(btn => { btn.onclick = (e) => { e.stopPropagation(); const id = btn.dataset.id; savedProfiles = savedProfiles.filter(p => p.id !== id); SaveProfiles(savedProfiles).catch(() => {}); renderProfiles(); }; }); } }
async function connectProfile(profile) { if (profile.type === 'ssh') { const opts = profile.options; openSSHDialog(); document.getElementById('ssh-host').value = opts.host || ''; document.getElementById('ssh-port').value = opts.port || 22; document.getElementById('ssh-user').value = opts.user || ''; document.getElementById('ssh-auth').value = opts.auth || 'agent'; document.getElementById('ssh-auth').dispatchEvent(new Event('change')); if (opts.auth === 'password') document.getElementById('ssh-password').value = opts.password || ''; if (opts.auth === 'publicKey' && opts.privateKeys && opts.privateKeys.length) document.getElementById('ssh-key-path').value = opts.privateKeys[0]; } else if (profile.type === 'local') { newTab(profile.options && profile.options.shell || profile.options && profile.options.command); } }
function addProfile() { const id = 'profile-' + Date.now(); savedProfiles.push({ id, type: 'ssh', name: 'New SSH Profile', group: '', options: { host: '', port: 22, user: '', auth: 'agent' }, createdAt: new Date().toISOString(), updatedAt: new Date().toISOString() }); SaveProfiles(savedProfiles).catch(() => {}); renderProfiles(); showToast('Profile added', 'info'); }
// ===== BUILD UI ====='''
)

# 20. Add SSH username default
c = c.replace(
    "hideSettings();\n}",
    "hideSettings();\n}\nGetUsername().then(u => { const el = document.getElementById('ssh-user'); if (el && !el.value) el.value = u; }).catch(() => {});"
)

with open('tabby-go/wails-app/frontend/src/main.js', 'w', encoding='utf-8') as f:
    f.write(c)

print('Patched successfully!')
