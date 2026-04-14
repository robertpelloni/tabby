export namespace api {
	
	export class PTYSpawnParams {
	    id: string;
	    command: string;
	    args?: string[];
	    env?: Record<string, string>;
	    cwd?: string;
	    columns: number;
	    rows: number;
	
	    static createFrom(source: any = {}) {
	        return new PTYSpawnParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.command = source["command"];
	        this.args = source["args"];
	        this.env = source["env"];
	        this.cwd = source["cwd"];
	        this.columns = source["columns"];
	        this.rows = source["rows"];
	    }
	}
	export class PTYSpawnResult {
	    id: string;
	    pid: number;
	
	    static createFrom(source: any = {}) {
	        return new PTYSpawnResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.pid = source["pid"];
	    }
	}
	export class SSHAlgorithms {
	    hmac?: string[];
	    kex?: string[];
	    cipher?: string[];
	    serverHostKey?: string[];
	    compression?: string[];
	
	    static createFrom(source: any = {}) {
	        return new SSHAlgorithms(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hmac = source["hmac"];
	        this.kex = source["kex"];
	        this.cipher = source["cipher"];
	        this.serverHostKey = source["serverHostKey"];
	        this.compression = source["compression"];
	    }
	}
	export class SSHAuthParams {
	    type: string;
	    password?: string;
	    privateKey?: string;
	    privateKeyPaths?: string[];
	    agentSocketPath?: string;
	    agentType?: string;
	    keyboardInteractivePassthrough?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SSHAuthParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.password = source["password"];
	        this.privateKey = source["privateKey"];
	        this.privateKeyPaths = source["privateKeyPaths"];
	        this.agentSocketPath = source["agentSocketPath"];
	        this.agentType = source["agentType"];
	        this.keyboardInteractivePassthrough = source["keyboardInteractivePassthrough"];
	    }
	}
	export class SSHConnectParams {
	    host: string;
	    port?: number;
	    user: string;
	    auth: SSHAuthParams;
	    keepaliveInterval?: number;
	    keepaliveCountMax?: number;
	    readyTimeout?: number;
	    agentForward?: boolean;
	    x11?: boolean;
	    x11Display?: string;
	    jumpHost?: SSHConnectParams;
	    algorithms?: SSHAlgorithms;
	    proxyCommand?: string;
	    socksProxyHost?: string;
	    socksProxyPort?: number;
	    httpProxyHost?: string;
	    httpProxyPort?: number;
	    environment?: Record<string, string>;
	    verifyHostKey?: boolean;
	    knownHostsPath?: string;
	    skipBanner?: boolean;
	    password?: string;
	
	    static createFrom(source: any = {}) {
	        return new SSHConnectParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.host = source["host"];
	        this.port = source["port"];
	        this.user = source["user"];
	        this.auth = this.convertValues(source["auth"], SSHAuthParams);
	        this.keepaliveInterval = source["keepaliveInterval"];
	        this.keepaliveCountMax = source["keepaliveCountMax"];
	        this.readyTimeout = source["readyTimeout"];
	        this.agentForward = source["agentForward"];
	        this.x11 = source["x11"];
	        this.x11Display = source["x11Display"];
	        this.jumpHost = this.convertValues(source["jumpHost"], SSHConnectParams);
	        this.algorithms = this.convertValues(source["algorithms"], SSHAlgorithms);
	        this.proxyCommand = source["proxyCommand"];
	        this.socksProxyHost = source["socksProxyHost"];
	        this.socksProxyPort = source["socksProxyPort"];
	        this.httpProxyHost = source["httpProxyHost"];
	        this.httpProxyPort = source["httpProxyPort"];
	        this.environment = source["environment"];
	        this.verifyHostKey = source["verifyHostKey"];
	        this.knownHostsPath = source["knownHostsPath"];
	        this.skipBanner = source["skipBanner"];
	        this.password = source["password"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SSHConnectionResult {
	    connectionId: string;
	    serverVersion: string;
	    remoteAddress: string;
	    banner?: string;
	    authMethods: string[];
	
	    static createFrom(source: any = {}) {
	        return new SSHConnectionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connectionId = source["connectionId"];
	        this.serverVersion = source["serverVersion"];
	        this.remoteAddress = source["remoteAddress"];
	        this.banner = source["banner"];
	        this.authMethods = source["authMethods"];
	    }
	}
	export class SSHSessionParams {
	    connectionId: string;
	    columns: number;
	    rows: number;
	    terminal?: string;
	    command?: string;
	    agentForward?: boolean;
	    x11?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SSHSessionParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connectionId = source["connectionId"];
	        this.columns = source["columns"];
	        this.rows = source["rows"];
	        this.terminal = source["terminal"];
	        this.command = source["command"];
	        this.agentForward = source["agentForward"];
	        this.x11 = source["x11"];
	    }
	}
	export class SSHSessionResult {
	    sessionId: string;
	
	    static createFrom(source: any = {}) {
	        return new SSHSessionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessionId = source["sessionId"];
	    }
	}
	export class SSHWriteParams {
	    connectionId: string;
	    sessionId: string;
	    data: string;
	
	    static createFrom(source: any = {}) {
	        return new SSHWriteParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connectionId = source["connectionId"];
	        this.sessionId = source["sessionId"];
	        this.data = source["data"];
	    }
	}

}

