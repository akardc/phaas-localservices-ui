export namespace app {
	
	export class EnvParam {
	    key: string;
	    value: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new EnvParam(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.value = source["value"];
	        this.enabled = source["enabled"];
	    }
	}
	export class Settings {
	    reposDirPath: string;
	    dataDirPath: string;
	    envParams: EnvParam[];
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.reposDirPath = source["reposDirPath"];
	        this.dataDirPath = source["dataDirPath"];
	        this.envParams = this.convertValues(source["envParams"], EnvParam);
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

}

export namespace repo {
	
	export enum State {
	    Unknown = "unknown",
	    starting = "starting",
	    running = "running",
	    stopped = "stopped",
	}
	export class BasicDetails {
	    name: string;
	    path: string;
	    statusNotificationChannel: string;
	
	    static createFrom(source: any = {}) {
	        return new BasicDetails(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.statusNotificationChannel = source["statusNotificationChannel"];
	    }
	}
	export class Status {
	    state: State;
	
	    static createFrom(source: any = {}) {
	        return new Status(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.state = source["state"];
	    }
	}

}

