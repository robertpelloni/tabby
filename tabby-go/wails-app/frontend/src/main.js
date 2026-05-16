import './style.css';
import './app.css';
import '@xterm/xterm/css/xterm.css';

import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import { WebLinksAddon } from '@xterm/addon-web-links';
import { SearchAddon } from '@xterm/addon-search';
import {
    PTYSpawn, PTYWrite, PTYResize, PTYKill,
    SSHConnect, SSHStartShell, SSHWrite, SSHResize, SSHClose,
    GetDefaultShell, GetAvailableShells,
    GetColorSchemes,
    SetWindowTitle, GetSettings, SaveSettings, ResetSettings,
    SaveSessionState, LoadSessionState, ClearSessionState,
    GetProfiles, SaveProfiles,
    GetUsername,
} from '../wailsjs/go/main/App';
import { EventsOn } from '../wailsjs/runtime/runtime';

// ===== GLOBALS =====
const COLOR_SCHEMES = {};
let schemeNames = [];
const tabs = [];
let activeTabId = null;
let tabCounter = 0;
let defaultShell = '';
let availableShells = [];
let findVisible = false;
let fontSize = 14;
let activeSplitPane = null;
let settings = {};
let savedProfiles = [];

// ===== INIT =====
async function init() {
    defaultShell = await GetDefaultShell();
    try { availableShells = await GetAvailableShells(); } catch (_) { availableShells = []; }
    try { settings = await GetSettings(); } catch (_) { settings = {}; }
    if (settings.FontSize) fontSize = settings.FontSize;

    // Load color schemes from Go backend
    try {
        const schemes = await GetColorSchemes();
        if (schemes && schemes.length) {
            schemes.forEach(s => { COLOR_SCHEMES[s.Name] = s; });
            schemeNames = schemes.map(s => s.Name);
        }
    } catch (_) {
        schemeNames = ['Tabby Default','Tabby Default Light','Dracula','Solarized Dark','Solarized Light','Monokai','Nord','One Half Dark','One Half Light','Gruvbox Dark','Tokyo Night','Catppuccin Mocha','Catppuccin Latte','Ayu Dark','Atom One Light','Batman'];
    }

    // Load connection profiles
    try { savedProfiles = await GetProfiles() || []; } catch (_) { savedProfiles = []; }

    buildUI();
    bindGlobalKeys();
    applySettingsToUI();
    const restored = await restoreSession();
    if (!restored) newTab();
}

// ===== COLOR SCHEME HELPERS =====
function getColorSchemeTheme(name) { const scheme = COLOR_SCHEMES[name]; if (!scheme) return null; const c = scheme.Colors || []; return { background: scheme.Background, foreground: scheme.Foreground, cursor: scheme.Cursor, cursorAccent: scheme.CursorAccent || undefined, selectionBackground: scheme.Selection || undefined, selectionForeground: scheme.SelectionForeground || undefined, black: c[0], red: c[1], green: c[2], yellow: c[3], blue: c[4], magenta: c[5], cyan: c[6], white: c[7], brightBlack: c[8], brightRed: c[9], brightGreen: c[10], brightYellow: c[11], brightBlue: c[12], brightMagenta: c[13], brightCyan: c[14], brightWhite: c[15] }; }
function isSchemeLight(name) { const bg = (COLOR_SCHEMES[name] && COLOR_SCHEMES[name].Background) || '#171717'; return isLightColor(bg); }
function isLightColor(hex) { if (!hex || hex.length < 7 || hex[0] !== '#') return false; const r = parseInt(hex.slice(1,3),16), g = parseInt(hex.slice(3,5),16), b = parseInt(hex.slice(5,7),16); return (0.299*r + 0.587*g + 0.114*b) / 255 > 0.5; }
function applyColorScheme(name) { const theme = getColorSchemeTheme(name); if (!theme) return; tabs.forEach(t => { if (t.term) t.term.options.theme = theme; }); if (isSchemeLight(name)) { document.body.classList.add('light-theme'); document.body.classList.remove('dark-theme'); } else { document.body.classList.add('dark-theme'); document.body.classList.remove('light-theme'); } renderColorSchemePreview(name); }
function renderColorSchemePreview(name) { const container = document.getElementById('color-scheme-preview'); if (!container) return; const scheme = COLOR_SCHEMES[name]; if (!scheme) { container.innerHTML = ''; return; } const c = scheme.Colors || []; const all = [scheme.Background, scheme.Foreground, scheme.Cursor, ...c]; container.innerHTML = all.map(color => `<div style="width:16px;height:16px;border-radius:3px;background:${color};border:1px solid #3a3a3a;" title="${color}"></div>`).join(''); }
// ===== SSH DIALOG =====
function openSSHDialog() { document.getElementById('ssh-dialog').classList.add('active'); document.getElementById('ssh-host').focus(); }
function closeSSHDialog() { document.getElementById('ssh-dialog').classList.remove('active'); const t = getActiveTab(); if (t) t.term.focus(); }
async function doSSHConnect() { const host = document.getElementById('ssh-host').value.trim(); const port = parseInt(document.getElementById('ssh-port').value) || 22; const user = document.getElementById('ssh-user').value.trim(); const auth = document.getElementById('ssh-auth').value; if (!host) { showToast('Host is required', 'error'); return; } closeSSHDialog(); showStatus('Connecting to ' + host + '...'); const authParams = { type: auth }; if (auth === 'password') authParams.password = document.getElementById('ssh-password').value; if (auth === 'publicKey') authParams.privateKeyPaths = [document.getElementById('ssh-key-path').value || '~/.ssh/id_ed25519']; try { const result = await SSHConnect({ host, port, user, auth: authParams, keepaliveInterval: 30, keepaliveCountMax: 3, readyTimeout: 15000 }); showToast('Connected to ' + host, 'success'); const tab = new Tab(defaultShell, 'ssh://' + user + '@' + host); tabs.push(tab); tab.activate(); tab.ptyId = null; tab.sshConnectionId = result.connectionId; tab.setTitle(user + '@' + host); tab.tabEl.querySelector('.tab-icon').textContent = '\U0001f510'; const shellResult = await SSHStartShell({ connectionId: result.connectionId, columns: tab.term.cols, rows: tab.term.rows, terminal: 'xterm-256color' }); tab.sshSessionId = shellResult.sessionId; tab.isSSH = true; tab.term.onData((data) => { if (tab.sshConnectionId && tab.sshSessionId) SSHWrite({ connectionId: tab.sshConnectionId, sessionId: tab.sshSessionId, data: btoa(data) }); }); showStatus('SSH - ' + user + '@' + host); if (document.getElementById('ssh-save-profile') && document.getElementById('ssh-save-profile').checked) { savedProfiles.push({ id: 'ssh-' + Date.now(), type: 'ssh', name: user + '@' + host, options: { host, port, user, auth, privateKeys: authParams.privateKeyPaths || [] }, createdAt: new Date().toISOString(), updatedAt: new Date().toISOString() }); SaveProfiles(savedProfiles).catch(() => {}); renderProfiles(); } } catch (err) { showToast('SSH failed: ' + err, 'error'); showStatus('SSH failed - ' + host); } }
// ===== PROFILES =====
function renderProfiles() { const section = document.getElementById('profiles-section'); const list = document.getElementById('profiles-list'); const editor = document.getElementById('profiles-editor-list'); if (!savedProfiles || savedProfiles.length === 0) { if (section) section.style.display = 'none'; if (editor) editor.innerHTML = '<div style="color:#666;font-size:12px;">No saved profiles yet.</div>'; return; } if (section) section.style.display = 'block'; if (list) { list.innerHTML = savedProfiles.map(p => { const icon = p.type === 'ssh' ? '\U0001f510' : p.type === 'serial' ? '\U0001f4e1' : '\u2318'; return `<div class="profile-item" data-profile-id="${p.id}" title="${p.name}"><span class="profile-icon">${icon}</span><span class="profile-name">${p.name}</span></div>`; }).join(''); list.querySelectorAll('.profile-item').forEach(el => { el.onclick = () => { const profile = savedProfiles.find(p => p.id === el.dataset.profileId); if (profile) connectProfile(profile); }; }); } if (editor) { editor.innerHTML = savedProfiles.map(p => { const icon = p.type === 'ssh' ? '\U0001f510' : p.type === 'serial' ? '\U0001f4e1' : '\u2318'; return `<div class="profile-editor-item"><span>${icon} ${p.name}</span><button class="btn-icon profile-delete" data-id="${p.id}" title="Delete">\u00d7</button></div>`; }).join(''); editor.querySelectorAll('.profile-delete').forEach(btn => { btn.onclick = (e) => { e.stopPropagation(); const id = btn.dataset.id; savedProfiles = savedProfiles.filter(p => p.id !== id); SaveProfiles(savedProfiles).catch(() => {}); renderProfiles(); }; }); } }
async function connectProfile(profile) { if (profile.type === 'ssh') { const opts = profile.options; openSSHDialog(); document.getElementById('ssh-host').value = opts.host || ''; document.getElementById('ssh-port').value = opts.port || 22; document.getElementById('ssh-user').value = opts.user || ''; document.getElementById('ssh-auth').value = opts.auth || 'agent'; document.getElementById('ssh-auth').dispatchEvent(new Event('change')); if (opts.auth === 'password') document.getElementById('ssh-password').value = opts.password || ''; if (opts.auth === 'publicKey' && opts.privateKeys && opts.privateKeys.length) document.getElementById('ssh-key-path').value = opts.privateKeys[0]; } else if (profile.type === 'local') { newTab(profile.options && profile.options.shell || profile.options && profile.options.command); } }
function addProfile() { const id = 'profile-' + Date.now(); savedProfiles.push({ id, type: 'ssh', name: 'New SSH Profile', group: '', options: { host: '', port: 22, user: '', auth: 'agent' }, createdAt: new Date().toISOString(), updatedAt: new Date().toISOString() }); SaveProfiles(savedProfiles).catch(() => {}); renderProfiles(); showToast('Profile added', 'info'); }
// ===== BUILD UI =====
function buildUI() {
    document.querySelector('#app').innerHTML = `
    <div id="sidebar">
        <div id="sidebar-header">
            <div class="logo"><span>⌘</span> Tabby</div>
            <div style="display:flex;gap:4px;">
                <button class="btn-icon" id="btn-new-tab" title="New Tab (Ctrl+Shift+T)">+</button>
                <button class="btn-icon" id="btn-settings" title="Settings (Ctrl+,)">⚙</button>
            </div>
        </div>
        <div id="tab-list"></div>
        <div id="profiles-section" style="border-top:1px solid #2b2b2b;padding:8px 0;max-height:200px;overflow-y:auto;display:none;">
            <div style="padding:0 12px 4px;font-size:10px;color:#666;text-transform:uppercase;letter-spacing:0.5px;">Profiles</div>
            <div id="profiles-list"></div>
        </div>
        <div id="sidebar-footer">
            <div class="status-dot" id="status-dot"></div>
            <div class="status-text" id="status-text">Ready</div>
        </div>
    </div>
    <div id="main-content">
        <div id="welcome">
            <div class="title"><span>Tabby</span> Go</div>
            <div class="shortcuts">
                <div class="shortcut"><kbd>Ctrl+Shift+T</kbd> New Tab</div>
                <div class="shortcut"><kbd>Ctrl+W</kbd> Close Tab</div>
                <div class="shortcut"><kbd>Ctrl+Tab</kbd> Next Tab</div>
                <div class="shortcut"><kbd>Ctrl+Shift+Tab</kbd> Prev Tab</div>
                <div class="shortcut"><kbd>Ctrl+\\</kbd> Split Vertical</div>
                <div class="shortcut"><kbd>Ctrl+Shift+\\</kbd> Split Horizontal</div>
                <div class="shortcut"><kbd>Ctrl+Shift+F</kbd> Find</div>
                <div class="shortcut"><kbd>Alt+1-9</kbd> Switch Tab</div>
                <div class="shortcut"><kbd>Ctrl+Shift+C</kbd> Copy</div>
                <div class="shortcut"><kbd>Ctrl+Shift+V</kbd> Paste</div>
                <div class="shortcut"><kbd>Ctrl+,</kbd> Settings</div>
                <div class="shortcut"><kbd>Ctrl++/-</kbd> Font Size</div>
            </div>
        </div>
    </div>
    <div id="settings-panel">
        <div id="settings-header">
            <h2>⚙ Settings</h2>
            <button id="settings-close" class="btn-icon">×</button>
        </div>
        <div id="settings-tabs">
            <button class="settings-tab active" data-tab="appearance">🎨 Appearance</button>
            <button class="settings-tab" data-tab="terminal">⌨ Terminal</button>
            <button class="settings-tab" data-tab="clipboard">📋 Clipboard</button>
            <button class="settings-tab" data-tab="mouse">🖱 Mouse</button>
            <button class="settings-tab" data-tab="tabs">📑 Tabs</button>
            <button class="settings-tab" data-tab="startup">🚀 Startup</button>
            <button class="settings-tab" data-tab="ssh">🔐 SSH</button>
            <button class="settings-tab" data-tab="serial">📡 Serial</button>
            <button class="settings-tab" data-tab="profiles">📁 Profiles</button>
        </div>
        <div id="settings-content">
            <!-- Appearance -->
            <div class="settings-page active" id="settings-appearance">
                <h3>Color Scheme</h3>
                <div class="setting-group"><label>Terminal Color Scheme</label><select id="s-color-scheme"></select></div>
                <div id="color-scheme-preview" style="display:flex;flex-wrap:wrap;gap:3px;padding:8px 0;"></div>
                <h3>Font</h3>
                <div class="setting-group"><label>Font Family</label><input type="text" id="s-font-family" placeholder="Cascadia Code, Fira Code, Consolas..."></div>
                <div class="setting-group"><label>Font Size</label>
                    <div class="slider-container"><input type="range" id="s-font-size" min="8" max="48" value="14"><span class="slider-label" id="s-font-size-val">14</span>px</div>
                </div>
                <div class="setting-group"><label>Fallback Font</label><input type="text" id="s-fallback-font" placeholder="e.g. Nerd Font"></div>
                <div class="setting-group"><label>Normal Font Weight</label><input type="number" id="s-font-weight" min="100" max="900" step="100" value="400"></div>
                <div class="setting-group"><label>Bold Font Weight</label><input type="number" id="s-font-weight-bold" min="100" max="900" step="100" value="700"></div>
                <div class="setting-group"><label>Line Height</label>
                    <div class="slider-container"><input type="range" id="s-line-height" min="1.0" max="2.0" step="0.05" value="1.2"><span class="slider-label" id="s-line-height-val">1.2</span>×</div>
                </div>
                <div class="setting-group"><label>Line Padding</label><input type="number" id="s-line-padding" min="0" max="10" value="0"></div>
                <div class="setting-group"><label>Enable Font Ligatures</label><div class="toggle-container"><input type="checkbox" id="s-ligatures"><label for="s-ligatures" class="toggle-label"></label></div></div>
                <h3>Theme & Colors</h3>
                <div class="setting-group"><label>Color Scheme Mode</label>
                    <select id="s-theme"><option value="auto">Auto (Follow System)</option><option value="dark">Always Dark</option><option value="light">Always Light</option></select>
                </div>
                <div class="setting-group"><label>Opacity</label>
                    <div class="slider-container"><input type="range" id="s-opacity" min="0.3" max="1.0" step="0.05" value="1.0"><span class="slider-label" id="s-opacity-val">1.0</span></div>
                </div>
                <div class="setting-group"><label>UI Spacing</label>
                    <select id="s-spaciness"><option value="1">Compact</option><option value="2">Normal</option><option value="3">Spacious</option></select>
                </div>
                <div class="setting-group"><label>Enable Animations</label><div class="toggle-container"><input type="checkbox" id="s-animations" checked><label for="s-animations" class="toggle-label"></label></div></div>
                <h3>Cursor</h3>
                <div class="setting-group"><label>Cursor Shape</label>
                    <select id="s-cursor-style"><option value="bar">Bar (|)</option><option value="block">Block (█)</option><option value="underline">Underline (▁)</option></select>
                </div>
                <div class="setting-group"><label>Blink Cursor</label><div class="toggle-container"><input type="checkbox" id="s-cursor-blink" checked><label for="s-cursor-blink" class="toggle-label"></label></div></div>
                <h3>Rendering</h3>
                <div class="setting-group"><label>Terminal Frontend</label>
                    <select id="s-frontend"><option value="xterm-webgl">xterm (WebGL)</option><option value="xterm">xterm (Canvas)</option><option value="block">Block Frontend (Experimental)</option></select>
                </div>
                <div class="setting-group"><label>Draw Bold Text in Bright Colors</label><div class="toggle-container"><input type="checkbox" id="s-draw-bold-bright" checked><label for="s-draw-bold-bright" class="toggle-label"></label></div></div>
                <div class="setting-group"><label>Minimum Contrast Ratio</label><input type="number" id="s-min-contrast" min="1" max="21" step="0.5" value="4"></div>
                <h3>Custom CSS</h3>
                <div class="setting-group"><textarea id="s-css" rows="4" placeholder="/* Custom CSS */"></textarea></div>
            </div>
            <!-- Terminal -->
            <div class="settings-page" id="settings-terminal">
                <h3>Shell</h3>
                <div class="setting-group"><label>Default Shell</label>
                    <select id="s-shell"><option value="">Auto-detect</option></select>
                </div>
                <h3>Scrollback</h3>
                <div class="setting-group"><label>Scrollback Lines</label><input type="number" id="s-scrollback" min="100" max="1000000" step="1000" value="25000"></div>
                <h3>Bell</h3>
                <div class="setting-group"><label>Terminal Bell</label>
                    <select id="s-bell"><option value="off">Off</option><option value="visual">Visual</option><option value="audible">Audible</option></select>
                </div>
                <h3>Keyboard</h3>
                <div class="setting-group"><label>Use Alt as Meta key</label><div class="toggle-container"><input type="checkbox" id="s-alt-is-meta"><label for="s-alt-is-meta" class="toggle-label"></label></div></div>
                <div class="setting-group"><label>Scroll on Input</label><div class="toggle-container"><input type="checkbox" id="s-scroll-on-input" checked><label for="s-scroll-on-input" class="toggle-label"></label></div></div>
                <h3>Windows</h3>
                <div class="setting-group"><label>Use ConPTY</label><div class="toggle-container"><input type="checkbox" id="s-use-conpty" checked><label for="s-use-conpty" class="toggle-label"></label></div></div>
                <div class="setting-group"><label>Set Tabby as %COMSPEC%</label><div class="toggle-container"><input type="checkbox" id="s-set-comspec"><label for="s-set-comspec" class="toggle-label"></label></div></div>
            </div>
            <!-- Clipboard -->
            <div class="settings-page" id="settings-clipboard">
                <h3>Clipboard</h3>
                <div class="setting-group"><label>Copy on Select</label><div class="toggle-container"><input type="checkbox" id="s-copy-on-select"><label for="s-copy-on-select" class="toggle-label"></label></div></div>
                <div class="setting-group"><label>Copy with Formatting (HTML)</label><div class="toggle-container"><input type="checkbox" id="s-copy-as-html" checked><label for="s-copy-as-html" class="toggle-label"></label></div></div>
                <div class="setting-group"><label>Bracketed Paste</label><div class="toggle-container"><input type="checkbox" id="s-bracketed-paste" checked><label for="s-bracketed-paste" class="toggle-label"></label></div></div>
                <div class="setting-group"><label>Warn on Multi-line Paste</label><div class="toggle-container"><input type="checkbox" id="s-warn-multiline" checked><label for="s-warn-multiline" class="toggle-label"></label></div></div>
                <div class="setting-group"><label>Replace Line Breaks with Spaces on Paste</label><div class="toggle-container"><input type="checkbox" id="s-replace-newlines"><label for="s-replace-newlines" class="toggle-label"></label></div></div>
                <div class="setting-group"><label>Trim Whitespace and Newlines</label><div class="toggle-container"><input type="checkbox" id="s-trim-whitespace" checked><label for="s-trim-whitespace" class="toggle-label"></label></div></div>
            </div>
            <!-- Mouse -->
            <div class="settings-page" id="settings-mouse">
                <h3>Mouse</h3>
                <div class="setting-group"><label>Right Click</label>
                    <select id="s-right-click"><option value="off">Off</option><option value="menu">Context Menu</option><option value="paste">Paste</option><option value="clipboard">Paste if No Selection, Else Copy</option></select>
                </div>
                <div class="setting-group"><label>Paste on Middle-Click</label><div class="toggle-container"><input type="checkbox" id="s-paste-middle-click" checked><label for="s-paste-middle-click" class="toggle-label"></label></div></div>
                <div class="setting-group"><label>Word Separators</label><input type="text" id="s-word-separator" value=" ()[]{}'&quot;"></div>
            </div>
            <!-- Tabs -->
            <div class="settings-page" id="settings-tabs">
                <h3>Tabs</h3>
                <div class="setting-group"><label>Tab Position</label>
                    <select id="s-tab-position"><option value="left">Left</option><option value="right">Right</option><option value="top">Top</option><option value="bottom">Bottom</option></select>
                </div>
                <div class="setting-group"><label>Last Tab Closes Window</label><div class="toggle-container"><input type="checkbox" id="s-last-tab-closes"><label for="s-last-tab-closes" class="toggle-label"></label></div></div>
                <div class="setting-group"><label>Cycle Tabs (wrap around)</label><div class="toggle-container"><input type="checkbox" id="s-cycle-tabs" checked><label for="s-cycle-tabs" class="toggle-label"></label></div></div>
                <div class="setting-group"><label>Hide Close Button</label><div class="toggle-container"><input type="checkbox" id="s-hide-close-button"><label for="s-hide-close-button" class="toggle-label"></label></div></div>
                <h3>Split Panes</h3>
                <div class="setting-group"><label>Pane Resize Step</label><input type="number" id="s-pane-resize-step" min="0.01" max="1.0" step="0.01" value="0.1"></div>
                <div class="setting-group"><label>Focus Follows Mouse</label><div class="toggle-container"><input type="checkbox" id="s-focus-follows-mouse"><label for="s-focus-follows-mouse" class="toggle-label"></label></div></div>
            </div>
            <!-- Startup -->
            <div class="settings-page" id="settings-startup">
                <h3>Startup</h3>
                <div class="setting-group"><label>Auto-open Terminal on Start</label><div class="toggle-container"><input type="checkbox" id="s-auto-open" checked><label for="s-auto-open" class="toggle-label"></label></div></div>
                <div class="setting-group"><label>Restore Tabs on Restart</label><div class="toggle-container"><input type="checkbox" id="s-recover-tabs" checked><label for="s-recover-tabs" class="toggle-label"></label></div></div>
                <h3>Window</h3>
                <div class="setting-group"><label>Window Frame</label>
                    <select id="s-frame"><option value="thin">Thin</option><option value="none">None (Borderless)</option><option value="native">Native</option></select>
                </div>
                <div class="setting-group"><label>Quake-style Dock</label>
                    <select id="s-dock"><option value="off">Off</option><option value="bottom">Bottom</option><option value="top">Top</option><option value="left">Left</option><option value="right">Right</option></select>
                </div>
                <div class="setting-group"><label>Hide Dock on Blur</label><div class="toggle-container"><input type="checkbox" id="s-dock-hide-blur"><label for="s-dock-hide-blur" class="toggle-label"></label></div></div>
                <div class="setting-group"><label>Dock Always on Top</label><div class="toggle-container"><input type="checkbox" id="s-dock-on-top" checked><label for="s-dock-on-top" class="toggle-label"></label></div></div>
                <div class="setting-group"><label>Hide Tray Icon</label><div class="toggle-container"><input type="checkbox" id="s-hide-tray"><label for="s-hide-tray" class="toggle-label"></label></div></div>
                <h3>Misc</h3>
                <div class="setting-group"><label>Language</label><input type="text" id="s-language" placeholder="Auto"></div>
                <div class="setting-group"><label>Enable Analytics</label><div class="toggle-container"><input type="checkbox" id="s-analytics" checked><label for="s-analytics" class="toggle-label"></label></div></div>
                <div class="setting-group"><label>Automatic Updates</label><div class="toggle-container"><input type="checkbox" id="s-auto-updates" checked><label for="s-auto-updates" class="toggle-label"></label></div></div>
                <div class="setting-group"><label>Experimental Features</label><div class="toggle-container"><input type="checkbox" id="s-experimental"><label for="s-experimental" class="toggle-label"></label></div></div>
            </div>
            <!-- SSH -->
            <div class="settings-page" id="settings-ssh">
                <h3>SSH</h3>
                <div class="setting-group"><label>Warn When Closing Active Connections</label><div class="toggle-container"><input type="checkbox" id="s-ssh-warn-close"><label for="s-ssh-warn-close" class="toggle-label"></label></div></div>
                <div class="setting-group"><label>Verify Host Keys When Connecting</label><div class="toggle-container"><input type="checkbox" id="s-ssh-verify-keys" checked><label for="s-ssh-verify-keys" class="toggle-label"></label></div></div>
                <div class="setting-group"><label>Disable Dynamic Title (SSH)</label><div class="toggle-container"><input type="checkbox" id="s-ssh-disable-title" checked><label for="s-ssh-disable-title" class="toggle-label"></label></div></div>
                <h3>Agent</h3>
                <div class="setting-group"><label>SSH Agent Type</label>
                    <select id="s-ssh-agent-type"><option value="auto">Automatic</option><option value="pageant">Pageant</option><option value="pipe">Named Pipe</option></select>
                </div>
                <div class="setting-group"><label>Agent Pipe Path</label><input type="text" id="s-ssh-agent-path" placeholder="Default: \\\\.\\pipe\\openssh-ssh-agent"></div>
                <h3>X11 Forwarding</h3>
                <div class="setting-group"><label>Override X11 Display</label><input type="text" id="s-ssh-x11" placeholder="Auto-detect"></div>
            </div>
            <!-- Serial -->
            <div class="settings-page" id="settings-serial">
                <h3>Serial Port Defaults</h3>
                <div class="setting-group"><label>Baud Rate</label>
                    <select id="s-serial-baud">
                        <option value="9600">9600</option><option value="19200">19200</option><option value="38400">38400</option>
                        <option value="57600">57600</option><option value="115200" selected>115200</option><option value="230400">230400</option>
                        <option value="460800">460800</option><option value="921600">921600</option><option value="1000000">1000000</option>
                    </select>
                </div>
                <div class="setting-group"><label>Data Bits</label>
                    <select id="s-serial-data-bits"><option value="5">5</option><option value="6">6</option><option value="7">7</option><option value="8" selected>8</option></select>
                </div>
                <div class="setting-group"><label>Stop Bits</label>
                    <select id="s-serial-stop-bits"><option value="1" selected>1</option><option value="2">2</option></select>
                </div>
                <div class="setting-group"><label>Parity</label>
                    <select id="s-serial-parity"><option value="none" selected>None</option><option value="even">Even</option><option value="odd">Odd</option></select>
                </div>
                <div class="setting-group"><label>Flow Control</label>
                    <select id="s-serial-flow"><option value="none" selected>None</option><option value="hardware">Hardware (RTS/CTS)</option><option value="software">Software (XON/XOFF)</option></select>
                </div>
            </div>
        </div>
            <div class="settings-page" id="settings-profiles">
                <h3>Saved Profiles</h3>
                <div id="profiles-editor-list" style="margin-bottom:12px;"></div>
                <button class="btn-primary" id="btn-add-profile" style="width:100%;margin-bottom:8px;">+ Add SSH Profile</button>
            </div>
        </div>
        <div id="settings-actions">
            <button id="btn-reset" class="btn-secondary">Reset to Defaults</button>
            <button id="btn-save" class="btn-primary">Save Settings</button>
        </div>
    </div>`;

    // Populate color scheme dropdown
    const schemeSelect = document.getElementById('s-color-scheme');
    schemeNames.forEach(name => { const opt = document.createElement('option'); opt.value = name; opt.textContent = name; schemeSelect.appendChild(opt); });
    schemeSelect.onchange = () => applyColorScheme(schemeSelect.value);

    // Populate shell dropdown
    const shellSelect = document.getElementById('s-shell');
    availableShells.forEach(s => {
        const opt = document.createElement('option');
        opt.value = s;
        opt.textContent = `${s.split(/[/\\]/).pop().replace('.exe', '')}  (${s})`;
        shellSelect.appendChild(opt);
    });

    // Render profiles
    renderProfiles();

    // Settings tab navigation
    document.querySelectorAll('.settings-tab').forEach(btn => {
        btn.onclick = () => {
            document.querySelectorAll('.settings-tab').forEach(b => b.classList.remove('active'));
            document.querySelectorAll('.settings-page').forEach(p => p.classList.remove('active'));
            btn.classList.add('active');
            document.getElementById(`settings-${btn.dataset.tab}`).classList.add('active');
        };
    });

    // Live-update sliders
    document.getElementById('s-font-size').oninput = (e) => {
        document.getElementById('s-font-size-val').textContent = e.target.value;
        applyFontSize(parseInt(e.target.value));
    };
    document.getElementById('s-line-height').oninput = (e) => {
        document.getElementById('s-line-height-val').textContent = parseFloat(e.target.value).toFixed(2);
        tabs.forEach(t => { if (t.term) { t.term.options.lineHeight = parseFloat(e.target.value); t.fitAddon.fit(); } });
    };
    document.getElementById('s-opacity').oninput = (e) => {
        document.getElementById('s-opacity-val').textContent = parseFloat(e.target.value).toFixed(2);
        document.getElementById('main-content').style.opacity = parseFloat(e.target.value);
    };

    // SSH auth toggle
    document.getElementById('ssh-auth').onchange = (e) => { document.getElementById('ssh-password-group').style.display = e.target.value === 'password' ? 'block' : 'none'; document.getElementById('ssh-key-group').style.display = e.target.value === 'publicKey' ? 'block' : 'none'; };

    // Button bindings
    document.getElementById('btn-new-tab').onclick = (e) => showNewTabDropdown(e);
    document.getElementById('btn-ssh').onclick = () => openSSHDialog();
    document.getElementById('btn-settings').onclick = () => toggleSettings();
    document.getElementById('settings-close').onclick = () => hideSettings();
    document.getElementById('btn-save').onclick = () => saveSettingsFromUI();
    document.getElementById('ssh-cancel').onclick = () => closeSSHDialog();
    document.getElementById('ssh-connect').onclick = () => doSSHConnect();
    document.getElementById('btn-add-profile').onclick = () => addProfile();
    document.getElementById('btn-reset').onclick = () => doResetSettings();
}

// ===== SETTINGS APPLY/SAVE =====
function applySettingsToUI() {
    const s = settings;
    const set = (id, val) => { const el = document.getElementById(id); if (el) el.value = val ?? ''; };
    const check = (id, val) => { const el = document.getElementById(id); if (el) el.checked = !!val; };

    // Appearance
    set('s-color-scheme', s.ColorScheme || 'Tabby Default');
    if (s.ColorScheme) applyColorScheme(s.ColorScheme);
    set('s-font-family', s.FontFamily);
    set('s-font-size', s.FontSize || 14);
    document.getElementById('s-font-size-val').textContent = s.FontSize || 14;
    if (s.FontSize) applyFontSize(s.FontSize);
    set('s-fallback-font', s.FallbackFont);
    set('s-font-weight', s.FontWeight || 400);
    set('s-font-weight-bold', s.FontWeightBold || 700);
    set('s-line-height', s.LineHeight || 1.2);
    document.getElementById('s-line-height-val').textContent = s.LineHeight || 1.2;
    set('s-line-padding', s.LinePadding || 0);
    check('s-ligatures', s.Ligatures);
    set('s-theme', s.Theme || 'dark');
    applyTheme(s.Theme || 'dark');
    set('s-opacity', s.Opacity ?? 1.0);
    document.getElementById('s-opacity-val').textContent = (s.Opacity ?? 1.0).toFixed(2);
    set('s-spaciness', s.Spaciness || 1);
    check('s-animations', s.Animations ?? true);
    set('s-cursor-style', s.CursorStyle || 'bar');
    check('s-cursor-blink', s.CursorBlink ?? true);
    set('s-frontend', s.Frontend || 'xterm-webgl');
    check('s-draw-bold-bright', s.DrawBoldTextInBrightColors ?? true);
    set('s-min-contrast', s.MinimumContrastRatio ?? 4);
    set('s-css', s.CSS || '');

    // Terminal
    set('s-shell', s.Shell || '');
    set('s-scrollback', s.Scrollback || 25000);
    set('s-bell', s.Bell || 'off');
    check('s-alt-is-meta', s.AltIsMeta);
    check('s-scroll-on-input', s.ScrollOnInput ?? true);
    check('s-use-conpty', s.UseConPTY ?? true);
    check('s-set-comspec', s.SetComSpec);

    // Clipboard
    check('s-copy-on-select', s.CopyOnSelect);
    check('s-copy-as-html', s.CopyAsHTML ?? true);
    check('s-bracketed-paste', s.BracketedPaste ?? true);
    check('s-warn-multiline', s.WarnOnMultilinePaste ?? true);
    check('s-replace-newlines', s.ReplaceNewlinesOnPaste);
    check('s-trim-whitespace', s.TrimWhitespaceOnPaste ?? true);

    // Mouse
    set('s-right-click', s.RightClick || 'menu');
    check('s-paste-middle-click', s.PasteOnMiddleClick ?? true);
    set('s-word-separator', s.WordSeparator || " ()[]{}'\"");

    // Tabs
    set('s-tab-position', s.TabPosition || 'left');
    check('s-last-tab-closes', s.LastTabClosesWindow);
    check('s-cycle-tabs', s.CycleTabs ?? true);
    check('s-hide-close-button', s.HideCloseButton);
    set('s-pane-resize-step', s.PaneResizeStep ?? 0.1);
    check('s-focus-follows-mouse', s.FocusFollowsMouse);

    // Startup
    check('s-auto-open', s.AutoOpen ?? true);
    check('s-recover-tabs', s.RecoverTabs ?? true);
    set('s-frame', s.Frame || 'thin');
    set('s-dock', s.Dock || 'off');
    check('s-dock-hide-blur', s.DockHideOnBlur);
    check('s-dock-on-top', s.DockAlwaysOnTop ?? true);
    check('s-hide-tray', s.HideTray);
    set('s-language', s.Language || '');
    check('s-analytics', s.EnableAnalytics ?? true);
    check('s-auto-updates', s.EnableAutomaticUpdates ?? true);
    check('s-experimental', s.EnableExperimentalFeatures);

    // SSH
    check('s-ssh-warn-close', s.SSHWarnOnClose);
    check('s-ssh-verify-keys', s.SSHVerifyHostKeys ?? true);
    check('s-ssh-disable-title', s.SSHDisableDynamicTitle ?? true);
    set('s-ssh-agent-type', s.SSHAgentType || 'auto');
    set('s-ssh-agent-path', s.SSHAgentPath || '');
    set('s-ssh-x11', s.SSHX11Display || '');

    // Serial
    set('s-serial-baud', s.SerialBaudRate || 115200);
    set('s-serial-data-bits', s.SerialDataBits || 8);
    set('s-serial-stop-bits', s.SerialStopBits || 1);
    set('s-serial-parity', s.SerialParity || 'none');
    set('s-serial-flow', s.SerialFlowControl || 'none');
}

function applyFontSize(size) {
    fontSize = size;
    tabs.forEach(t => { if (t.term) { t.term.options.fontSize = size; t.fitAddon.fit(); if (t.ptyId) PTYResize(t.ptyId, t.term.cols, t.term.rows); } });
}

function applyTheme(theme) {
    if (theme === 'light') { document.body.classList.add('light-theme'); document.body.classList.remove('dark-theme'); }
    else if (theme === 'dark') { document.body.classList.add('dark-theme'); document.body.classList.remove('light-theme'); }
    else { document.body.classList.remove('light-theme'); document.body.classList.remove('dark-theme'); }
    // Update xterm themes
    const isDark = !isSchemeLight(settings.ColorScheme || 'Tabby Default');
    const schemeTheme = getColorSchemeTheme(settings.ColorScheme || 'Tabby Default'); tabs.forEach(t => { if (t.term && schemeTheme) t.term.options.theme = schemeTheme; });
}

const FALLBACK_DARK = { background: '#1e1e1e', foreground: '#cccccc', cursor: '#aeafad', selectionBackground: '#264f78', black: '#1e1e1e', red: '#f44747', green: '#6a9955', yellow: '#d7ba7d', blue: '#569cd6', magenta: '#c586c0', cyan: '#4ec9b0', white: '#cccccc', brightBlack: '#666666', brightRed: '#f44747', brightGreen: '#6a9955', brightYellow: '#d7ba7d', brightBlue: '#569cd6', brightMagenta: '#c586c0', brightCyan: '#4ec9b0', brightWhite: '#e0e0e0' };

async function saveSettingsFromUI() {
    const s = {
        ColorScheme: document.getElementById('s-color-scheme').value || 'Tabby Default',
        FontFamily: document.getElementById('s-font-family').value.trim(),
        FontSize: parseInt(document.getElementById('s-font-size').value) || 14,
        FallbackFont: document.getElementById('s-fallback-font').value.trim(),
        FontWeight: parseInt(document.getElementById('s-font-weight').value) || 400,
        FontWeightBold: parseInt(document.getElementById('s-font-weight-bold').value) || 700,
        LineHeight: parseFloat(document.getElementById('s-line-height').value) || 1.2,
        LinePadding: parseInt(document.getElementById('s-line-padding').value) || 0,
        Ligatures: document.getElementById('s-ligatures').checked,
        Theme: document.getElementById('s-theme').value || 'dark',
        Opacity: parseFloat(document.getElementById('s-opacity').value) || 1.0,
        Spaciness: parseInt(document.getElementById('s-spaciness').value) || 1,
        Animations: document.getElementById('s-animations').checked,
        CursorStyle: document.getElementById('s-cursor-style').value || 'bar',
        CursorBlink: document.getElementById('s-cursor-blink').checked,
        Frontend: document.getElementById('s-frontend').value || 'xterm-webgl',
        DrawBoldTextInBrightColors: document.getElementById('s-draw-bold-bright').checked,
        MinimumContrastRatio: parseFloat(document.getElementById('s-min-contrast').value) || 4,
        CSS: document.getElementById('s-css').value,
        Shell: document.getElementById('s-shell').value || '',
        Scrollback: parseInt(document.getElementById('s-scrollback').value) || 25000,
        Bell: document.getElementById('s-bell').value || 'off',
        AltIsMeta: document.getElementById('s-alt-is-meta').checked,
        ScrollOnInput: document.getElementById('s-scroll-on-input').checked,
        UseConPTY: document.getElementById('s-use-conpty').checked,
        SetComSpec: document.getElementById('s-set-comspec').checked,
        CopyOnSelect: document.getElementById('s-copy-on-select').checked,
        CopyAsHTML: document.getElementById('s-copy-as-html').checked,
        BracketedPaste: document.getElementById('s-bracketed-paste').checked,
        WarnOnMultilinePaste: document.getElementById('s-warn-multiline').checked,
        ReplaceNewlinesOnPaste: document.getElementById('s-replace-newlines').checked,
        TrimWhitespaceOnPaste: document.getElementById('s-trim-whitespace').checked,
        RightClick: document.getElementById('s-right-click').value || 'menu',
        PasteOnMiddleClick: document.getElementById('s-paste-middle-click').checked,
        WordSeparator: document.getElementById('s-word-separator').value,
        TabPosition: document.getElementById('s-tab-position').value || 'left',
        LastTabClosesWindow: document.getElementById('s-last-tab-closes').checked,
        CycleTabs: document.getElementById('s-cycle-tabs').checked,
        HideCloseButton: document.getElementById('s-hide-close-button').checked,
        PaneResizeStep: parseFloat(document.getElementById('s-pane-resize-step').value) || 0.1,
        FocusFollowsMouse: document.getElementById('s-focus-follows-mouse').checked,
        AutoOpen: document.getElementById('s-auto-open').checked,
        RecoverTabs: document.getElementById('s-recover-tabs').checked,
        Frame: document.getElementById('s-frame').value || 'thin',
        Dock: document.getElementById('s-dock').value || 'off',
        DockHideOnBlur: document.getElementById('s-dock-hide-blur').checked,
        DockAlwaysOnTop: document.getElementById('s-dock-on-top').checked,
        HideTray: document.getElementById('s-hide-tray').checked,
        Language: document.getElementById('s-language').value.trim(),
        EnableAnalytics: document.getElementById('s-analytics').checked,
        EnableAutomaticUpdates: document.getElementById('s-auto-updates').checked,
        EnableExperimentalFeatures: document.getElementById('s-experimental').checked,
        SSHWarnOnClose: document.getElementById('s-ssh-warn-close').checked,
        SSHVerifyHostKeys: document.getElementById('s-ssh-verify-keys').checked,
        SSHDisableDynamicTitle: document.getElementById('s-ssh-disable-title').checked,
        SSHAgentType: document.getElementById('s-ssh-agent-type').value || 'auto',
        SSHAgentPath: document.getElementById('s-ssh-agent-path').value.trim(),
        SSHX11Display: document.getElementById('s-ssh-x11').value.trim(),
        SerialBaudRate: parseInt(document.getElementById('s-serial-baud').value) || 115200,
        SerialDataBits: parseInt(document.getElementById('s-serial-data-bits').value) || 8,
        SerialStopBits: parseInt(document.getElementById('s-serial-stop-bits').value) || 1,
        SerialParity: document.getElementById('s-serial-parity').value || 'none',
        SerialFlowControl: document.getElementById('s-serial-flow').value || 'none',
    };
    try { await SaveSettings(s); settings = s; applySettingsToUI(); showToast('Settings saved', 'success'); }
    catch (_) { showToast('Failed to save settings', 'error'); }
    hideSettings();
}
GetUsername().then(u => { const el = document.getElementById('ssh-user'); if (el && !el.value) el.value = u; }).catch(() => {});

function doResetSettings() {
    ResetSettings().then(() => { settings = {}; fontSize = 14; applyFontSize(14); applyTheme('dark'); applySettingsToUI(); showToast('Reset to defaults', 'info'); })
    .catch(() => { showToast('Failed to reset', 'error'); });
}

function toggleSettings() { const p = document.getElementById('settings-panel'); if (p.classList.contains('active')) { hideSettings(); return; } applySettingsToUI(); p.classList.add('active'); }
function hideSettings() { document.getElementById('settings-panel').classList.remove('active'); const tab = getActiveTab(); if (tab) tab.term.focus(); }

// ===== TOAST =====
function showToast(message, type = 'info') { const existing = document.querySelector('.toast'); if (existing) existing.remove(); const toast = document.createElement('div'); toast.className = `toast ${type}`; toast.textContent = message; document.body.appendChild(toast); requestAnimationFrame(() => toast.classList.add('visible')); setTimeout(() => { toast.classList.remove('visible'); setTimeout(() => toast.remove(), 300); }, 2500); }

// ===== NEW TAB DROPDOWN =====
function showNewTabDropdown(e) {
    document.querySelectorAll('#new-tab-dropdown').forEach(d => d.remove());
    const dropdown = document.createElement('div'); dropdown.id = 'new-tab-dropdown';
    let html = `<div class="shell-item" data-shell=""><span class="shell-name">Default Shell</span><span class="shell-path">${defaultShell}</span></div>`;
    availableShells.forEach(s => { const name = s.split(/[/\\]/).pop().replace('.exe', ''); html += `<div class="shell-item" data-shell="${s}"><span class="shell-name">${name}</span><span class="shell-path">${s}</span></div>`; });
    dropdown.innerHTML = html;
    const rect = e.currentTarget.getBoundingClientRect();
    dropdown.style.position = 'fixed'; dropdown.style.left = rect.left + 'px'; dropdown.style.top = (rect.bottom + 4) + 'px';
    document.body.appendChild(dropdown);
    dropdown.onclick = (ev) => { const item = ev.target.closest('.shell-item'); if (!item) return; dropdown.remove(); newTab(item.dataset.shell || undefined); };
    setTimeout(() => { document.addEventListener('click', function handler(ev) { if (!dropdown.contains(ev.target)) { dropdown.remove(); document.removeEventListener('click', handler); } }); }, 10);
}

// ===== TAB CLASS =====
class Tab {
    constructor(shell) {
        this.id = `tab-${Date.now()}-${tabCounter++}`;
        this.ptyId = null; this.title = 'Shell'; this.shell = shell || defaultShell; this.exited = false; this.isSSH = false; this.sshConnectionId = null; this.sshSessionId = null;
        const fontFamily = settings.FontFamily || '"Cascadia Code","Fira Code",Consolas,"Courier New",monospace';
        const lineHeight = settings.LineHeight || 1.2;
        const scrollback = settings.Scrollback || 25000;
        const cursorStyle = settings.CursorStyle || 'bar';
        const cursorBlink = settings.CursorBlink ?? true;
        const colorScheme = settings.ColorScheme || 'Tabby Default'; const theme = getColorSchemeTheme(colorScheme);
        const fontWeight = settings.FontWeight || 400;
        const fontWeightBold = settings.FontWeightBold || 700;

        this.term = new Terminal({
            cursorBlink, cursorStyle, fontFamily, fontSize, fontWeight, fontWeightBold,
            lineHeight, allowProposedApi: true, scrollback,
            bellStyle: settings.Bell || 'off',
            theme: theme || FALLBACK_DARK,
        });
        this.fitAddon = new FitAddon(); this.searchAddon = new SearchAddon(); this.webLinksAddon = new WebLinksAddon();
        this.term.loadAddon(this.fitAddon); this.term.loadAddon(this.searchAddon); this.term.loadAddon(this.webLinksAddon);

        this.wrapper = document.createElement('div'); this.wrapper.className = 'terminal-wrapper'; this.wrapper.id = this.id;
        document.getElementById('main-content').appendChild(this.wrapper); this.term.open(this.wrapper);

        this.tabEl = document.createElement('div'); this.tabEl.className = 'tab-item'; this.tabEl.draggable = true; this.tabEl.dataset.tabId = this.id;
        this.tabEl.innerHTML = `<span class="tab-icon">⌘</span><span class="tab-title">${this.title}</span><button class="tab-close">×</button>`;
        document.getElementById('tab-list').appendChild(this.tabEl);

        this.tabEl.onclick = (e) => { if (!e.target.classList.contains('tab-close')) this.activate(); };
        this.tabEl.querySelector('.tab-close').onclick = (e) => { e.stopPropagation(); this.close(); };
        this.tabEl.oncontextmenu = (e) => { e.preventDefault(); showTabContextMenu(e, this); };
        this.tabEl.ondragstart = (e) => { e.dataTransfer.setData('text/plain', this.id); this.tabEl.classList.add('dragging'); };
        this.tabEl.ondragend = () => { this.tabEl.classList.remove('dragging'); };
        this.tabEl.ondragover = (e) => { e.preventDefault(); this.tabEl.classList.add('drag-over'); };
        this.tabEl.ondragleave = () => { this.tabEl.classList.remove('drag-over'); };
        this.tabEl.ondrop = (e) => { e.preventDefault(); this.tabEl.classList.remove('drag-over'); const draggedId = e.dataTransfer.getData('text/plain'); if (draggedId && draggedId !== this.id) reorderTab(draggedId, this.id); };

        this.term.onData((data) => { if (this.ptyId && !this.exited) PTYWrite(this.ptyId, btoa(data)); });
        this.term.onTitleChange((title) => { if (title) this.setTitle(title); });

        this.dataHandler = (params) => { if ((params.ptyId ?? params.PTYID) === this.ptyId) this.term.write(atob(params.data)); if (this.isSSH && (params.sessionId ?? params.SessionID) === this.sshSessionId) this.term.write(atob(params.data)); };
        this.exitHandler = (params) => { if ((params.ptyId ?? params.PTYID) === this.ptyId) { this.exited = true; const code = params.exitCode ?? 0; this.term.writeln(`\r\n\x1b[1;33m[Process exited — code ${code}]\x1b[0m`); this.setTitle(`Exit (${code})`); this.tabEl.querySelector('.tab-icon').textContent = '✕'; this.tabEl.querySelector('.tab-icon').style.color = '#f44747'; } };
        window.__ptyDataHandlers = window.__ptyDataHandlers || []; window.__ptyExitHandlers = window.__ptyExitHandlers || [];
        window.__ptyDataHandlers.push(this.dataHandler); window.__ptyExitHandlers.push(this.exitHandler);
    }

    async spawn() {
        try { const result = await PTYSpawn({ command: this.shell, args: [], env: {}, columns: this.term.cols, rows: this.term.rows }); this.ptyId = result.id; const name = this.shell.split(/[/\\]/).pop().replace('.exe', ''); this.setTitle(name); showStatus(`Connected — ${name}`); }
        catch (err) { this.term.writeln(`\x1b[1;31mFailed to spawn shell: ${err}\x1b[0m`); showToast(`Shell spawn failed: ${err}`, 'error'); }
    }

    activate() {
        const w = document.getElementById('welcome'); if (w) w.style.display = 'none';
        tabs.forEach(t => { t.wrapper.classList.remove('active'); t.tabEl.classList.remove('active'); });
        this.wrapper.classList.add('active'); this.tabEl.classList.add('active'); activeTabId = this.id;
        this.term.focus();
        requestAnimationFrame(() => { this.fitAddon.fit(); if (this.ptyId && !this.exited) PTYResize(this.ptyId, this.term.cols, this.term.rows); });
        saveSession();
    }

    close() {
        window.__ptyDataHandlers = (window.__ptyDataHandlers || []).filter(h => h !== this.dataHandler);
        window.__ptyExitHandlers = (window.__ptyExitHandlers || []).filter(h => h !== this.exitHandler);
        if (this.ptyId) PTYKill(this.ptyId, '').catch(() => {}); if (this.isSSH && this.sshConnectionId) SSHClose({ connectionId: this.sshConnectionId }).catch(() => {});
        this.term.dispose(); this.wrapper.remove(); this.tabEl.remove();
        const idx = tabs.indexOf(this); if (idx > -1) tabs.splice(idx, 1);
        if (activeTabId === this.id) { if (tabs.length > 0) tabs[Math.min(idx, tabs.length - 1)].activate(); else { activeTabId = null; const w = document.getElementById('welcome'); if (w) w.style.display = 'flex'; } }
        saveSession();
    }

    setTitle(title) { this.title = title; const el = this.tabEl.querySelector('.tab-title'); if (el) el.textContent = title; if (activeTabId === this.id) SetWindowTitle(`Tabby — ${title}`); }
    findNext(q) { if (q) this.searchAddon.findNext(q); }
    findPrevious(q) { if (q) this.searchAddon.findPrevious(q); }
    copySelection() { const sel = this.term.getSelection(); if (sel) navigator.clipboard.writeText(sel).then(() => showToast('Copied', 'success')); }
    async pasteFromClipboard() { try { const text = await navigator.clipboard.readText(); if (text) { if (this.isSSH && this.sshConnectionId && this.sshSessionId) SSHWrite({ connectionId: this.sshConnectionId, sessionId: this.sshSessionId, data: btoa(text) }); else if (this.ptyId && !this.exited) PTYWrite(this.ptyId, btoa(text)); } } catch (_) { showToast('Clipboard access denied', 'error'); } }
}

// ===== PTY EVENTS =====
EventsOn('pty.data', (params) => { (window.__ptyDataHandlers || []).forEach(h => h(params)); });
EventsOn('pty.exit', (params) => { (window.__ptyExitHandlers || []).forEach(h => h(params)); });
EventsOn('ssh.data', (params) => { (window.__ptyDataHandlers || []).forEach(h => h(params)); });
EventsOn('ssh.exit', (params) => { (window.__ptyExitHandlers || []).forEach(h => h(params)); });

// ===== TAB MANAGEMENT =====
function newTab(shell) { const tab = new Tab(shell); tabs.push(tab); tab.activate(); tab.spawn(); return tab; }
function getActiveTab() { return tabs.find(t => t.id === activeTabId); }
function switchToTab(i) { if (i >= 0 && i < tabs.length) tabs[i].activate(); }
function nextTab() { const i = tabs.findIndex(t => t.id === activeTabId); tabs[(i + 1) % tabs.length].activate(); }
function prevTab() { const i = tabs.findIndex(t => t.id === activeTabId); tabs[(i - 1 + tabs.length) % tabs.length].activate(); }

function reorderTab(draggedId, targetId) {
    const fromIdx = tabs.findIndex(t => t.id === draggedId); const toIdx = tabs.findIndex(t => t.id === targetId);
    if (fromIdx === -1 || toIdx === -1 || fromIdx === toIdx) return;
    const [moved] = tabs.splice(fromIdx, 1); tabs.splice(toIdx, 0, moved);
    const list = document.getElementById('tab-list'); list.innerHTML = ''; tabs.forEach(t => list.appendChild(t.tabEl));
    saveSession();
}

// ===== SPLIT PANE =====
class SplitPane {
    constructor(orientation) { this.orientation = orientation; this.panes = []; this.element = document.createElement('div'); this.element.className = `split-container ${orientation}`; this.dividers = []; }
    addPane(tab, flex) { const pe = document.createElement('div'); pe.className = 'pane'; pe.style.flex = `${flex} 1 0%`; pe.appendChild(tab.wrapper); tab.wrapper.classList.add('active'); this.panes.push({ tab, element: pe, flex }); this.element.appendChild(pe); if (this.panes.length > 1) this._addDivider(); return pe; }
    _addDivider() { const d = document.createElement('div'); d.className = 'splitter'; if (this.orientation === 'vertical') d.classList.add('horizontal'); this.element.appendChild(d); this.dividers.push(d);
        let dragging = false, startPos = 0, sf0 = 50, sf1 = 50;
        const onStart = (e) => { if (this.panes.length !== 2) return; e.preventDefault(); dragging = true; startPos = this.orientation === 'vertical' ? e.clientY : e.clientX; sf0 = this.panes[0].flex; sf1 = this.panes[1].flex; d.classList.add('dragging'); };
        const onMove = (e) => { if (!dragging) return; e.preventDefault(); const pos = this.orientation === 'vertical' ? e.clientY : e.clientX; const rect = this.element.getBoundingClientRect(); const total = this.orientation === 'vertical' ? rect.height : rect.width; const pct = ((pos - startPos) / total) * 100; const f0 = Math.max(20, Math.min(80, sf0 + pct)); this.panes[0].flex = f0; this.panes[1].flex = 100 - f0; this.panes[0].element.style.flex = `${f0} 1 0%`; this.panes[1].element.style.flex = `${100 - f0} 1 0%`; };
        const onEnd = () => { if (!dragging) return; dragging = false; d.classList.remove('dragging'); this.panes.forEach(p => { p.tab.fitAddon.fit(); if (p.tab.ptyId) PTYResize(p.tab.ptyId, p.tab.term.cols, p.tab.term.rows); }); };
        d.addEventListener('mousedown', onStart); document.addEventListener('mousemove', onMove); document.addEventListener('mouseup', onEnd); }
    removePane(tab) { const idx = this.panes.findIndex(p => p.tab === tab); if (idx === -1) return null; this.panes[idx].element.remove(); this.panes.splice(idx, 1); if (this.dividers.length > 0) this.dividers.pop()?.remove(); if (this.panes.length === 1) { this.element.remove(); return this.panes[0].tab; } return null; }
}

function splitVertically() { const tab = getActiveTab(); if (!tab) return; const mc = document.getElementById('main-content'); const w = document.getElementById('welcome'); if (w) w.style.display = 'none'; const idx = tabs.indexOf(tab); if (idx > -1) tabs.splice(idx, 1); tab.wrapper.remove(); tab.tabEl.remove(); const sp = new SplitPane('vertical'); sp.addPane(tab, 50); const t2 = new Tab(tab.shell); tabs.push(t2); sp.addPane(t2, 50); mc.appendChild(sp.element); activeSplitPane = sp; t2.spawn(); tab.activate(); showToast('Split vertically', 'info'); }
function splitHorizontally() { const tab = getActiveTab(); if (!tab) return; const mc = document.getElementById('main-content'); const w = document.getElementById('welcome'); if (w) w.style.display = 'none'; const idx = tabs.indexOf(tab); if (idx > -1) tabs.splice(idx, 1); tab.wrapper.remove(); tab.tabEl.remove(); const sp = new SplitPane('horizontal'); sp.addPane(tab, 50); const t2 = new Tab(tab.shell); tabs.push(t2); sp.addPane(t2, 50); mc.appendChild(sp.element); activeSplitPane = sp; t2.spawn(); tab.activate(); showToast('Split horizontally', 'info'); }
function removeSplit() { if (!activeSplitPane) return; const tab = getActiveTab(); if (!tab) return; const remaining = activeSplitPane.removePane(tab); if (tab.ptyId) PTYKill(tab.ptyId, '').catch(() => {}); tab.term.dispose(); const idx = tabs.indexOf(tab); if (idx > -1) tabs.splice(idx, 1); const mc = document.getElementById('main-content'); if (remaining) { mc.appendChild(remaining.wrapper); document.getElementById('tab-list').appendChild(remaining.tabEl); tabs.push(remaining); activeSplitPane = null; remaining.activate(); } else if (tabs.length === 0) { activeSplitPane = null; const w = document.getElementById('welcome'); if (w) w.style.display = 'flex'; } }
function closeAllSplits() { if (!activeSplitPane) return; [...activeSplitPane.panes].forEach(p => { if (p.tab.ptyId) PTYKill(p.tab.ptyId, '').catch(() => {}); p.tab.term.dispose(); const idx = tabs.indexOf(p.tab); if (idx > -1) tabs.splice(idx, 1); }); activeSplitPane.element.remove(); activeSplitPane = null; if (tabs.length === 0) { const w = document.getElementById('welcome'); if (w) w.style.display = 'flex'; } else tabs[0].activate(); showToast('All splits closed', 'info'); }

// ===== TAB CONTEXT MENU =====
function showTabContextMenu(e, tab) {
    document.querySelectorAll('.context-menu').forEach(m => m.remove());
    const menu = document.createElement('div'); menu.className = 'context-menu';
    menu.innerHTML = `<div class="context-menu-item" data-action="rename">✏️ Rename</div><div class="context-menu-item" data-action="duplicate">📋 Duplicate</div><div class="context-menu-separator"></div><div class="context-menu-item" data-action="close-others">✕ Close Others</div><div class="context-menu-item" data-action="close-right">✕ Close to Right</div><div class="context-menu-item" data-action="close">✕ Close</div>`;
    document.body.appendChild(menu);
    menu.style.left = Math.min(e.clientX, window.innerWidth - 180) + 'px'; menu.style.top = Math.min(e.clientY, window.innerHeight - 200) + 'px';
    const close = () => menu.remove();
    menu.onclick = (ev) => { const item = ev.target.closest('.context-menu-item'); if (!item) return; close();
        switch (item.dataset.action) {
            case 'rename': { const n = prompt('Tab name:', tab.title); if (n) tab.setTitle(n); break; }
            case 'duplicate': newTab(tab.shell); break;
            case 'close-others': { [...tabs].forEach(t => { if (t !== tab) t.close(); }); tab.activate(); break; }
            case 'close-right': { const idx = tabs.indexOf(tab); for (let i = tabs.length - 1; i > idx; i--) tabs[i].close(); break; }
            case 'close': tab.close(); break;
        }
    };
    setTimeout(() => document.addEventListener('click', close, { once: true }), 10);
}

// ===== FIND BAR =====
function toggleFind() {
    let bar = document.getElementById('find-bar'); if (bar) { bar.remove(); findVisible = false; const t = getActiveTab(); if (t) t.term.focus(); return; }
    findVisible = true; bar = document.createElement('div'); bar.id = 'find-bar';
    bar.style.cssText = 'position:absolute;top:0;right:0;z-index:100;background:#2d2d2d;border-bottom:1px solid #3a3a3a;border-left:1px solid #3a3a3a;padding:6px 12px;display:flex;gap:8px;align-items:center;border-radius:0 0 0 8px;';
    bar.innerHTML = `<input type="text" id="find-input" placeholder="Find..." style="background:#1e1e1e;border:1px solid #3a3a3a;color:#ccc;padding:4px 8px;border-radius:4px;font-size:13px;width:200px;outline:none;"><button class="btn-icon" id="find-prev">↑</button><button class="btn-icon" id="find-next">↓</button><button class="btn-icon" id="find-close">×</button>`;
    document.getElementById('main-content').appendChild(bar);
    const input = document.getElementById('find-input'); input.focus();
    input.onkeydown = (e) => { const t = getActiveTab(); if (!t) return; if (e.key === 'Enter') { e.preventDefault(); e.shiftKey ? t.findPrevious(input.value) : t.findNext(input.value); } if (e.key === 'Escape') toggleFind(); };
    document.getElementById('find-next').onclick = () => { const t = getActiveTab(); if (t) t.findNext(input.value); input.focus(); };
    document.getElementById('find-prev').onclick = () => { const t = getActiveTab(); if (t) t.findPrevious(input.value); input.focus(); };
    document.getElementById('find-close').onclick = () => toggleFind();
}

// ===== SESSION PERSISTENCE =====
function saveSession() { try { const tabStates = tabs.map((t) => ({ Shell: t.shell, Title: t.title, Active: t.id === activeTabId })); SaveSessionState(tabStates).catch(() => {}); } catch (_) {} }
async function restoreSession() { try { const state = await LoadSessionState(); if (!state || !state.Tabs || state.Tabs.length === 0) return false; let activated = false; state.Tabs.forEach((saved) => { const tab = new Tab(saved.Shell); tabs.push(tab); tab.spawn(); if (saved.Active) { tab.activate(); activated = true; } else if (saved.Title) tab.setTitle(saved.Title); }); if (!activated && tabs.length > 0) tabs[0].activate(); return true; } catch (_) { return false; } }

function showStatus(msg) { const el = document.getElementById('status-text'); if (el) { el.textContent = msg; clearTimeout(window.__statusTimeout); window.__statusTimeout = setTimeout(() => { el.textContent = `${tabs.length} tab${tabs.length !== 1 ? 's' : ''}`; }, 3000); } }

// ===== KEYBOARD SHORTCUTS =====
function bindGlobalKeys() {
    document.addEventListener('keydown', (e) => {
        const ctrl = e.ctrlKey || e.metaKey; const shift = e.shiftKey;
        const inInput = document.activeElement && (document.activeElement.tagName === 'INPUT' || document.activeElement.tagName === 'TEXTAREA' || document.activeElement.tagName === 'SELECT');

        if (ctrl && !shift && e.key === ',') { e.preventDefault(); toggleSettings(); return; }
        if (ctrl && shift && (e.key === 'T' || e.key === 't')) { e.preventDefault(); newTab(); return; }
        if (ctrl && !shift && e.key === 'w' && !inInput) { e.preventDefault(); const t = getActiveTab(); if (t) t.close(); return; }
        if (ctrl && e.key === 'Tab' && !shift) { e.preventDefault(); nextTab(); return; }
        if (ctrl && e.key === 'Tab' && shift) { e.preventDefault(); prevTab(); return; }
        if (e.altKey && e.key >= '1' && e.key <= '9') { e.preventDefault(); switchToTab(parseInt(e.key) - 1); return; }
        if (ctrl && !shift && e.key === '\\') { e.preventDefault(); splitVertically(); return; }
        if (ctrl && shift && e.key === '\\') { e.preventDefault(); splitHorizontally(); return; }
        if (ctrl && shift && e.key === 'W') { e.preventDefault(); removeSplit(); return; }
        if (ctrl && shift && (e.key === 'F' || e.key === 'f')) { e.preventDefault(); toggleFind(); return; }
        if (ctrl && shift && (e.key === 'C' || e.key === 'c') && !inInput) { e.preventDefault(); const t = getActiveTab(); if (t) t.copySelection(); return; }
        if (ctrl && shift && (e.key === 'V' || e.key === 'v') && !inInput) { e.preventDefault(); const t = getActiveTab(); if (t) t.pasteFromClipboard(); return; }
        if (ctrl && (e.key === '=' || e.key === '+')) { e.preventDefault(); applyFontSize(Math.min(48, fontSize + 1)); return; }
        if (ctrl && e.key === '-') { e.preventDefault(); applyFontSize(Math.max(8, fontSize - 1)); return; }
        if (ctrl && e.key === '0') { e.preventDefault(); applyFontSize(14); return; }
    });
    window.addEventListener('focus', () => { const tab = getActiveTab(); if (tab && !tab.exited) tab.term.focus(); });
}

window.addEventListener('resize', () => {
    const tab = getActiveTab(); if (tab) { tab.fitAddon.fit(); if (tab.ptyId && !tab.exited) PTYResize(tab.ptyId, tab.term.cols, tab.term.rows); }
    if (activeSplitPane) activeSplitPane.panes.forEach(p => { p.tab.fitAddon.fit(); if (p.tab.ptyId && !p.tab.exited) PTYResize(p.tab.ptyId, p.tab.term.cols, p.tab.term.rows); });
});

init();