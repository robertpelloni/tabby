import './style.css';
import './app.css';
import '@xterm/xterm/css/xterm.css';

import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import { PTYSpawn, PTYWrite, PTYResize, PTYKill } from '../wailsjs/go/main/App';
import { EventsOn } from '../wailsjs/runtime/runtime';

const sidebar = document.createElement('div');
sidebar.id = 'sidebar';
sidebar.innerHTML = `
    <div id="sidebar-header">
        <h2>Sessions</h2>
        <button id="add-tab-btn">+</button>
    </div>
    <div id="tab-list"></div>
`;

const mainContent = document.createElement('div');
mainContent.id = 'main-content';

document.querySelector('#app').appendChild(sidebar);
document.querySelector('#app').appendChild(mainContent);

const tabs = [];
let activeTab = null;

class Tab {
    constructor(id, title) {
        this.id = id;
        this.title = title;
        this.ptyId = null;
        this.term = new Terminal({
            cursorBlink: true,
            fontFamily: 'Consolas, "Courier New", monospace',
            fontSize: 14,
            theme: { background: '#1e1e1e' }
        });
        this.fitAddon = new FitAddon();
        this.term.loadAddon(this.fitAddon);

        this.wrapper = document.createElement('div');
        this.wrapper.className = 'terminal-wrapper';
        mainContent.appendChild(this.wrapper);
        this.term.open(this.wrapper);

        this.tabItem = document.createElement('div');
        this.tabItem.className = 'tab-item';
        this.tabItem.innerHTML = `<span>${title}</span><span class="tab-close">×</span>`;
        document.getElementById('tab-list').appendChild(this.tabItem);

        this.tabItem.onclick = () => this.activate();
        this.tabItem.querySelector('.tab-close').onclick = (e) => {
            e.stopPropagation();
            this.close();
        };

        this.initPTY();
    }

    async initPTY() {
        try {
            const shell = navigator.platform.indexOf('Win') !== -1 ? 'powershell.exe' : '/bin/bash';
            const result = await PTYSpawn({
                command: shell,
                args: [],
                env: {}
            });
            this.ptyId = result.id;

            this.term.onData(data => {
                PTYWrite(this.ptyId, btoa(data));
            });

            EventsOn('pty.data', (params) => {
                if (params.ptyId === this.ptyId) {
                    this.term.write(atob(params.data));
                }
            });

            EventsOn('pty.exit', (params) => {
                if (params.ptyId === this.ptyId) {
                    this.term.writeln('\n\x1b[1;31mProcess exited with code ' + params.exitCode + '\x1b[0m');
                }
            });

            // Initial fit
            setTimeout(() => {
                this.fitAddon.fit();
                PTYResize(this.ptyId, this.term.cols, this.term.rows);
            }, 100);

        } catch (err) {
            this.term.writeln('\x1b[1;31mError spawning PTY: ' + err + '\x1b[0m');
        }
    }

    activate() {
        if (activeTab) {
            activeTab.wrapper.classList.remove('active');
            activeTab.tabItem.classList.remove('active');
        }
        this.wrapper.classList.add('active');
        this.tabItem.classList.add('active');
        activeTab = this;
        this.term.focus();
        this.fitAddon.fit();
        if (this.ptyId) {
            PTYResize(this.ptyId, this.term.cols, this.term.rows);
        }
    }

    close() {
        if (this.ptyId) {
            PTYKill(this.ptyId, "");
        }
        this.wrapper.remove();
        this.tabItem.remove();
        const index = tabs.indexOf(this);
        if (index > -1) tabs.splice(index, 1);
        if (activeTab === this && tabs.length > 0) {
            tabs[tabs.length - 1].activate();
        }
    }
}

function addTab() {
    const tabId = Date.now().toString();
    const tab = new Tab(tabId, `Local Shell ${tabs.length + 1}`);
    tabs.push(tab);
    tab.activate();
}

document.getElementById('add-tab-btn').onclick = addTab;

window.addEventListener('resize', () => {
    if (activeTab) {
        activeTab.fitAddon.fit();
        if (activeTab.ptyId) {
            PTYResize(activeTab.ptyId, activeTab.term.cols, activeTab.term.rows);
        }
    }
});

// Create initial tab
addTab();
