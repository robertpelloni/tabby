#!/usr/bin/env python3
"""Patch main.js to add Tab Reconnect, Profile Groups, and Enhanced Session Persistence."""
import re

with open('main.js', 'rb') as f:
    data = f.read()

content = data.decode('utf-8')
changes = 0

# 1. Add GetHomeDir and GetPlatform to imports
old = 'GetUsername,\n}'
new = 'GetUsername, GetHomeDir, GetPlatform,\n}'
if old in content:
    content = content.replace(old, new)
    changes += 1
    print('1. Added GetHomeDir, GetPlatform to imports')

# 2. Enhance session persistence to save connection type and options
old_save = """function saveSession() {
  try {
    const tabStates = tabs.map((t) => ({
      Shell: t.shell,
      Title: t.title,
      Active: t.id === activeTabId
    }));
    SaveSessionState(tabStates).catch(() => {});
  } catch (_) {}
}"""
new_save = """function saveSession() {
  try {
    const tabStates = tabs.map((t) => ({
      Shell: t.shell,
      Title: t.title,
      Active: t.id === activeTabId,
      Type: t.isSSH ? 'ssh' : (t.isSerial ? 'serial' : (t.isTelnet ? 'telnet' : 'local')),
      Host: t.isSSH ? t.sshHost : (t.isTelnet ? t.telnetHost : ''),
      Port: t.isSSH ? t.sshPort : (t.isTelnet ? t.telnetPort : 0),
      User: t.isSSH ? t.sshUser : '',
      Exited: t.exited,
    }));
    SaveSessionState(tabStates).catch(() => {});
  } catch (_) {}
}"""
if old_save in content:
    content = content.replace(old_save, new_save)
    changes += 1
    print('2. Enhanced saveSession with connection type and host info')

# 3. Enhance restoreSession to reconnect SSH/telnet tabs
old_restore = """async function restoreSession() {
  try {
    const state = await LoadSessionState();
    if (!state || !state.Tabs || state.Tabs.length === 0) return false;
    let activated = false;
    state.Tabs.forEach((saved) => {
      const tab = new Tab(saved.Shell);
      tabs.push(tab);
      tab.spawn();
      if (saved.Active) {
        tab.activate();
        activated = true;
      } else if (saved.Title) tab.setTitle(saved.Title);
    });
    if (!activated && tabs.length > 0) tabs[0].activate();
    return true;
  } catch (_) { return false; }
}"""
new_restore = """async function restoreSession() {
  try {
    const state = await LoadSessionState();
    if (!state || !state.Tabs || state.Tabs.length === 0) return false;
    let activated = false;
    for (const saved of state.Tabs) {
      const tabType = saved.Type || 'local';
      if (tabType === 'ssh' && saved.Host && !saved.Exited) {
        // Show a reconnect prompt for SSH tabs
        const tab = new Tab(saved.Shell || defaultShell);
        tabs.push(tab);
        tab.activate();
        tab.setTitle((saved.User || 'root') + '@' + saved.Host + ' [reconnecting...]');
        tab.tabEl.querySelector('.tab-icon').textContent = '\\u23f3';
        tab.term.writeln('\\x1b[1;33m[Reconnecting to ' + (saved.User || 'root') + '@' + saved.Host + '...]\\x1b[0m\\r\\n');
        // Attempt auto-reconnect
        try {
          const result = await SSHConnect({
            host: saved.Host,
            port: saved.Port || 22,
            user: saved.User || 'root',
            auth: { type: 'agent' },
            keepaliveInterval: 30,
            keepaliveCountMax: 3,
            readyTimeout: 15000,
          });
          tab.sshConnectionId = result.connectionId;
          tab.sshHost = saved.Host;
          tab.sshPort = saved.Port || 22;
          tab.sshUser = saved.User || 'root';
          tab.isSSH = true;
          let jumpLabel = '';
          if (result.jumpChain && result.jumpChain.length > 0) {
            jumpLabel = ' (via ' + result.jumpChain.join(' -> ') + ')';
          }
          const shellResult = await SSHStartShell({
            connectionId: result.connectionId,
            columns: tab.term.cols,
            rows: tab.term.rows,
            terminal: 'xterm-256color',
          });
          tab.sshSessionId = shellResult.sessionId;
          tab.setTitle((saved.User || 'root') + '@' + saved.Host + jumpLabel);
          tab.tabEl.querySelector('.tab-icon').textContent = '\\U0001f510';
          tab.term.onData((data) => {
            if (tab.sshConnectionId && tab.sshSessionId)
              SSHWrite({ connectionId: tab.sshConnectionId, sessionId: tab.sshSessionId, data: btoa(data) });
          });
          tab.term.writeln('\\x1b[1;32m[Reconnected]\\x1b[0m\\r\\n');
          showStatus('SSH - ' + (saved.User || 'root') + '@' + saved.Host);
          showToast('Reconnected to ' + saved.Host, 'success');
        } catch (err) {
          tab.term.writeln('\\x1b[1;31m[Reconnect failed: ' + err + ']\\x1b[0m\\r\\n');
          tab.setTitle((saved.User || 'root') + '@' + saved.Host + ' [disconnected]');
          tab.tabEl.querySelector('.tab-icon').textContent = '\\u2715';
          tab.tabEl.querySelector('.tab-icon').style.color = '#f44747';
          tab.exited = true;
        }
        if (saved.Active) activated = true;
      } else if (tabType === 'telnet' && saved.Host && !saved.Exited) {
        const tab = new Tab(saved.Shell || defaultShell);
        tabs.push(tab);
        tab.activate();
        tab.setTitle(saved.Host + ':' + (saved.Port || 23) + ' [reconnecting...]');
        try {
          const result = await TelnetConnect(saved.Host, saved.Port || 23);
          tab.telnetConnectionId = result.ConnectionID || result.connectionId;
          tab.telnetHost = saved.Host;
          tab.telnetPort = saved.Port || 23;
          tab.isTelnet = true;
          tab.setTitle(saved.Host + ':' + (saved.Port || 23));
          tab.tabEl.querySelector('.tab-icon').textContent = '\\U0001f310';
          tab.telnetDataHandler = (params) => {
            const cid = params.ConnectionID || params.connectionId;
            if (cid === tab.telnetConnectionId) tab.term.write(atob(params.Data || params.data));
          };
          window.__telnetDataHandlers.push(tab.telnetDataHandler);
          tab.term.onData((data) => {
            if (tab.telnetConnectionId) TelnetWrite(tab.telnetConnectionId, btoa(data));
          });
          showStatus('Telnet - ' + saved.Host + ':' + (saved.Port || 23));
          showToast('Reconnected to ' + saved.Host, 'success');
        } catch (err) {
          tab.term.writeln('\\x1b[1;31m[Reconnect failed: ' + err + ']\\x1b[0m\\r\\n');
          tab.setTitle(saved.Host + ' [disconnected]');
          tab.exited = true;
        }
        if (saved.Active) activated = true;
      } else {
        // Local shell tab
        const tab = new Tab(saved.Shell);
        tabs.push(tab);
        tab.spawn();
        if (saved.Active) { tab.activate(); activated = true; }
        else if (saved.Title) tab.setTitle(saved.Title);
      }
    }
    if (!activated && tabs.length > 0) tabs[0].activate();
    return true;
  } catch (_) { return false; }
}"""
if old_restore in content:
    content = content.replace(old_restore, new_restore)
    changes += 1
    print('3. Enhanced restoreSession with SSH/telnet reconnection')

# 4. Add reconnect button to tab context menu
old_ctx = "case 'close': tab.close(); break;"
new_ctx = "case 'reconnect': reconnectTab(tab); break;\n      case 'close': tab.close(); break;"
if old_ctx in content:
    content = content.replace(old_ctx, new_ctx)
    changes += 1
    print('4. Added reconnect to tab context menu')

# 5. Store SSH host/port/user on the tab object when connecting
old_ssh = "tab.isSSH = true;"
new_ssh = "tab.isSSH = true; tab.sshHost = host; tab.sshPort = port; tab.sshUser = user;"
if old_ssh in content:
    content = content.replace(old_ssh, new_ssh)
    changes += 1
    print('5. Added sshHost/sshPort/sshUser to SSH tab')

# 6. Store telnet host/port on the tab object when connecting
old_telnet = "tab.isTelnet = true;"
new_telnet = "tab.isTelnet = true; tab.telnetHost = host; tab.telnetPort = port;"
if old_telnet in content:
    content = content.replace(old_telnet, new_telnet)
    changes += 1
    print('6. Added telnetHost/telnetPort to Telnet tab')

# 7. Inject reconnect function and profile groups before PROFILES section
marker = '// ===== PROFILES ====='
if marker in content:
    new_funcs = """

// ===== TAB RECONNECT =====
async function reconnectTab(tab) {
  if (tab.isSSH && tab.sshHost) {
    tab.term.writeln('\\r\\n\\x1b[1;33m[Reconnecting to ' + tab.sshUser + '@' + tab.sshHost + '...]\\x1b[0m\\r\\n');
    tab.setTitle(tab.sshUser + '@' + tab.sshHost + ' [reconnecting...]');
    tab.tabEl.querySelector('.tab-icon').textContent = '\\u23f3';
    tab.tabEl.querySelector('.tab-icon').style.color = '';
    tab.exited = false;
    try {
      if (tab.sshConnectionId) SSHClose({ connectionId: tab.sshConnectionId }).catch(() => {});
      const result = await SSHConnect({
        host: tab.sshHost,
        port: tab.sshPort || 22,
        user: tab.sshUser || 'root',
        auth: { type: 'agent' },
        keepaliveInterval: 30,
        keepaliveCountMax: 3,
        readyTimeout: 15000,
      });
      tab.sshConnectionId = result.connectionId;
      const shellResult = await SSHStartShell({
        connectionId: result.connectionId,
        columns: tab.term.cols,
        rows: tab.term.rows,
        terminal: 'xterm-256color',
      });
      tab.sshSessionId = shellResult.sessionId;
      tab.exited = false;
      let jumpLabel = '';
      if (result.jumpChain && result.jumpChain.length > 0) {
        jumpLabel = ' (via ' + result.jumpChain.join(' -> ') + ')';
      }
      tab.setTitle(tab.sshUser + '@' + tab.sshHost + jumpLabel);
      tab.tabEl.querySelector('.tab-icon').textContent = '\\U0001f510';
      tab.term.onData((data) => {
        if (tab.sshConnectionId && tab.sshSessionId)
          SSHWrite({ connectionId: tab.sshConnectionId, sessionId: tab.sshSessionId, data: btoa(data) });
      });
      tab.term.writeln('\\x1b[1;32m[Reconnected]\\x1b[0m\\r\\n');
      showStatus('SSH - ' + tab.sshUser + '@' + tab.sshHost);
      showToast('Reconnected to ' + tab.sshHost, 'success');
    } catch (err) {
      tab.term.writeln('\\x1b[1;31m[Reconnect failed: ' + err + ']\\x1b[0m\\r\\n');
      tab.setTitle(tab.sshUser + '@' + tab.sshHost + ' [disconnected]');
      tab.tabEl.querySelector('.tab-icon').textContent = '\\u2715';
      tab.tabEl.querySelector('.tab-icon').style.color = '#f44747';
      tab.exited = true;
      showToast('Reconnect failed: ' + err, 'error');
    }
  } else if (tab.isTelnet && tab.telnetHost) {
    tab.term.writeln('\\r\\n\\x1b[1;33m[Reconnecting to ' + tab.telnetHost + '...]\\x1b[0m\\r\\n');
    try {
      if (tab.telnetConnectionId) TelnetClose(tab.telnetConnectionId).catch(() => {});
      const result = await TelnetConnect(tab.telnetHost, tab.telnetPort || 23);
      tab.telnetConnectionId = result.ConnectionID || result.connectionId;
      tab.exited = false;
      tab.setTitle(tab.telnetHost + ':' + (tab.telnetPort || 23));
      tab.tabEl.querySelector('.tab-icon').textContent = '\\U0001f310';
      tab.telnetDataHandler = (params) => {
        const cid = params.ConnectionID || params.connectionId;
        if (cid === tab.telnetConnectionId) tab.term.write(atob(params.Data || params.data));
      };
      window.__telnetDataHandlers = (window.__telnetDataHandlers || []).filter(h => h !== tab.telnetDataHandler);
      window.__telnetDataHandlers.push(tab.telnetDataHandler);
      tab.term.onData((data) => {
        if (tab.telnetConnectionId) TelnetWrite(tab.telnetConnectionId, btoa(data));
      });
      tab.term.writeln('\\x1b[1;32m[Reconnected]\\x1b[0m\\r\\n');
      showToast('Reconnected to ' + tab.telnetHost, 'success');
    } catch (err) {
      tab.term.writeln('\\x1b[1;31m[Reconnect failed: ' + err + ']\\x1b[0m\\r\\n');
      tab.exited = true;
      showToast('Reconnect failed: ' + err, 'error');
    }
  } else {
    // Re-spawn local shell
    tab.exited = false;
    tab.tabEl.querySelector('.tab-icon').textContent = '\\u2318';
    tab.tabEl.querySelector('.tab-icon').style.color = '';
    tab.spawn();
    showToast('Shell restarted', 'info');
  }
}

// ===== PROFILE GROUPS =====
function getProfileGroups() {
  const groups = {};
  savedProfiles.forEach(p => {
    const group = p.group || 'Ungrouped';
    if (!groups[group]) groups[group] = [];
    groups[group].push(p);
  });
  return groups;
}

"""
    content = content.replace(marker, new_funcs + '\n' + marker)
    changes += 1
    print('7. Injected reconnect and profile groups functions')

# 8. Add 'Reconnect' option to tab context menu buildUI
old_menu = "'close': 'Close Tab'"
new_menu = "'reconnect': 'Reconnect',\n        'close': 'Close Tab'"
if old_menu in content:
    content = content.replace(old_menu, new_menu)
    changes += 1
    print('8. Added Reconnect to context menu items')

# Write back
with open('main.js', 'wb') as f:
    f.write(content.encode('utf-8'))
print(f'\nTotal: {changes} changes applied.')
