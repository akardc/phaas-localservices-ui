export namespace repo {
	
	export enum RunningStatus {
	    Running = "Running",
	    Stopped = "Stopped",
	    Unknown = "Unknown",
	}

}

export namespace repobrowser {
	
	export class ListReposOptions {
	    nameRegex: string;
	
	    static createFrom(source: any = {}) {
	        return new ListReposOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.nameRegex = source["nameRegex"];
	    }
	}
	export class RepoInfo {
	    name: string;
	    // Go type: time
	    lastModified: any;
	    branch: string;
	
	    static createFrom(source: any = {}) {
	        return new RepoInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.lastModified = this.convertValues(source["lastModified"], null);
	        this.branch = source["branch"];
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

