export namespace app {
	
	export class InkLevel {
	    color: string;
	    level: number;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new InkLevel(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.color = source["color"];
	        this.level = source["level"];
	        this.status = source["status"];
	    }
	}
	export class PrinterInfo {
	    path: string;
	    model: string;
	    des: string;
	    mfg: string;
	    serial: string;
	
	    static createFrom(source: any = {}) {
	        return new PrinterInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.model = source["model"];
	        this.des = source["des"];
	        this.mfg = source["mfg"];
	        this.serial = source["serial"];
	    }
	}
	export class WasteCounter {
	    index: number;
	    value: number;
	    max: number;
	    ratio: number;
	
	    static createFrom(source: any = {}) {
	        return new WasteCounter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.value = source["value"];
	        this.max = source["max"];
	        this.ratio = source["ratio"];
	    }
	}
	export class StatusView {
	    state: string;
	    error: string;
	    source: string;
	    serial: string;
	    inkLevels: InkLevel[];
	    wasteCounters: WasteCounter[];
	
	    static createFrom(source: any = {}) {
	        return new StatusView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.state = source["state"];
	        this.error = source["error"];
	        this.source = source["source"];
	        this.serial = source["serial"];
	        this.inkLevels = this.convertValues(source["inkLevels"], InkLevel);
	        this.wasteCounters = this.convertValues(source["wasteCounters"], WasteCounter);
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

