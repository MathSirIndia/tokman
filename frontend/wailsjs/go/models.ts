export namespace main {
	
	export class SystemStatus {
	    status: string;
	    uptime: string;
	    uptime_seconds: number;
	    active_pools: string[];
	    cache_entries: number;
	    process_ram_mb: number;
	    system_ram_used_mb: number;
	    system_ram_total_mb: number;
	    storage_db_kb: number;
	    storage_disk_free_gb: number;
	    storage_disk_total_gb: number;
	    memory_mb: number;
	    goroutines: number;
	    port: number;
	    spend_summary?: storage.SpendSummary;
	
	    static createFrom(source: any = {}) {
	        return new SystemStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.uptime = source["uptime"];
	        this.uptime_seconds = source["uptime_seconds"];
	        this.active_pools = source["active_pools"];
	        this.cache_entries = source["cache_entries"];
	        this.process_ram_mb = source["process_ram_mb"];
	        this.system_ram_used_mb = source["system_ram_used_mb"];
	        this.system_ram_total_mb = source["system_ram_total_mb"];
	        this.storage_db_kb = source["storage_db_kb"];
	        this.storage_disk_free_gb = source["storage_disk_free_gb"];
	        this.storage_disk_total_gb = source["storage_disk_total_gb"];
	        this.memory_mb = source["memory_mb"];
	        this.goroutines = source["goroutines"];
	        this.port = source["port"];
	        this.spend_summary = this.convertValues(source["spend_summary"], storage.SpendSummary);
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

export namespace storage {
	
	export class SpendRecord {
	    id: number;
	    // Go type: time
	    timestamp: any;
	    model: string;
	    prompt_tokens: number;
	    completion_tokens: number;
	    total_tokens: number;
	    latency_ms: number;
	    cached: boolean;
	    client_id: string;
	
	    static createFrom(source: any = {}) {
	        return new SpendRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.model = source["model"];
	        this.prompt_tokens = source["prompt_tokens"];
	        this.completion_tokens = source["completion_tokens"];
	        this.total_tokens = source["total_tokens"];
	        this.latency_ms = source["latency_ms"];
	        this.cached = source["cached"];
	        this.client_id = source["client_id"];
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
	export class SpendSummary {
	    total_requests: number;
	    total_prompt_tokens: number;
	    total_completion_tokens: number;
	    total_tokens: number;
	    average_latency_ms: number;
	    cache_hit_ratio: number;
	
	    static createFrom(source: any = {}) {
	        return new SpendSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total_requests = source["total_requests"];
	        this.total_prompt_tokens = source["total_prompt_tokens"];
	        this.total_completion_tokens = source["total_completion_tokens"];
	        this.total_tokens = source["total_tokens"];
	        this.average_latency_ms = source["average_latency_ms"];
	        this.cache_hit_ratio = source["cache_hit_ratio"];
	    }
	}

}

