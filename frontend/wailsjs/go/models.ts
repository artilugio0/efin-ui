export namespace main {
	
	export class Header {
	    name: string;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new Header(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.value = source["value"];
	    }
	}
	export class Pane {
	    layout: string;
	    panes: Pane[];
	    content: number;
	
	    static createFrom(source: any = {}) {
	        return new Pane(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.layout = source["layout"];
	        this.panes = this.convertValues(source["panes"], Pane);
	        this.content = source["content"];
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
	export class Request {
	    id: string;
	    body: string;
	    host: string;
	    method: string;
	    timestamp: string;
	    url: string;
	    headers: Header[];
	
	    static createFrom(source: any = {}) {
	        return new Request(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.body = source["body"];
	        this.host = source["host"];
	        this.method = source["method"];
	        this.timestamp = source["timestamp"];
	        this.url = source["url"];
	        this.headers = this.convertValues(source["headers"], Header);
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
	export class Response {
	    id: string;
	    status_code: number;
	    body: string;
	    headers: Header[];
	
	    static createFrom(source: any = {}) {
	        return new Response(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.status_code = source["status_code"];
	        this.body = source["body"];
	        this.headers = this.convertValues(source["headers"], Header);
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
	export class RequestResponseDetail {
	    request?: Request;
	    response?: Response;
	
	    static createFrom(source: any = {}) {
	        return new RequestResponseDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.request = this.convertValues(source["request"], Request);
	        this.response = this.convertValues(source["response"], Response);
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
	
	export class UIAction {
	    action_type: string;
	    command_submitted?: string;
	    row_submitted?: Record<string, string>;
	    command_suggestion_requested?: string;
	
	    static createFrom(source: any = {}) {
	        return new UIAction(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.action_type = source["action_type"];
	        this.command_submitted = source["command_submitted"];
	        this.row_submitted = source["row_submitted"];
	        this.command_suggestion_requested = source["command_suggestion_requested"];
	    }
	}
	export class UIState {
	    current_tab: number;
	    tabs: Pane[];
	    focused_pane: number[];
	
	    static createFrom(source: any = {}) {
	        return new UIState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.current_tab = source["current_tab"];
	        this.tabs = this.convertValues(source["tabs"], Pane);
	        this.focused_pane = source["focused_pane"];
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
	export class UIActionResult {
	    result_type: string;
	    error: string;
	    request_response_table: string[][];
	    request_response_detail: RequestResponseDetail;
	    command_suggestion: string[];
	    ui_state?: UIState;
	
	    static createFrom(source: any = {}) {
	        return new UIActionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.result_type = source["result_type"];
	        this.error = source["error"];
	        this.request_response_table = source["request_response_table"];
	        this.request_response_detail = this.convertValues(source["request_response_detail"], RequestResponseDetail);
	        this.command_suggestion = source["command_suggestion"];
	        this.ui_state = this.convertValues(source["ui_state"], UIState);
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

