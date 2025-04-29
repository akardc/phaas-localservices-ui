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

