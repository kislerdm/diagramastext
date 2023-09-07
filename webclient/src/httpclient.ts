export interface HTTPClient {
    do(input: RequestInfo | URL, init?: RequestInit): Promise<Response>
}

export class defaultHTTPClient implements HTTPClient {
    do = window.fetch.bind(window)
}

export class mockResponse implements Response {
    readonly body: ReadableStream<Uint8Array> | null = null;
    readonly bodyUsed: boolean = true;
    readonly headers: Headers;
    readonly ok: boolean = false;
    readonly redirected: boolean = false;
    readonly status: number;
    readonly statusText: string = "";
    readonly type: ResponseType = "default";
    readonly url: string = "";

    private readonly response_obj: Object | undefined;

    constructor(status_code: number, response_obj?: Object, headers?: Headers) {
        this.status = status_code;
        if (status_code < 300) {
            this.ok = true
        }
        this.response_obj = response_obj;
        this.headers = headers!
    }

    arrayBuffer(): Promise<ArrayBuffer> {
        throw new Error("not implemented")
    }

    blob(): Promise<Blob> {
        throw new Error("not implemented")
    }

    formData(): Promise<FormData> {
        throw new Error("not implemented")
    }

    json(): Promise<any> {
        return Promise.resolve(this.response_obj);
    }

    text(): Promise<string> {
        throw new Error("not implemented")
    }

    clone(): Response {
        return this;
    }

}

export class mockHTTPClient implements HTTPClient {
    private readonly response: Response;
    private readonly err?: Error;

    input: RequestInfo | URL = ""
    init?: RequestInit

    constructor(response: Response, err?: Error) {
        this.response = response;
        this.err = err;
    }

    get_url(): string {
        return this.input.toString();
    }

    do(input: RequestInfo | URL, init?: RequestInit): Promise<Response> {
        this.input = input;
        this.init = init;
        return new Promise<Response>((resolve, reject) => {
            if (this.err !== undefined) {
                reject(this.err);
            } else {
                resolve(this.response)
            }
        });
    }
}
