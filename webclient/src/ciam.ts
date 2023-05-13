import {getCookie, setCookie} from "typescript-cookie";

const defaultNA = "NA";

type jwt_header = {
    alg: string;
    typ: string;
}

type user_quotas = {
    prompt_length_max: number;
    rpm: number;
    rpd: number;
}

type jwt_payload = {
    email_verified: string;
    email: string;
    fingerprint: string;
    role: string;
    quotas: user_quotas;
    sub: string;
    iss: string;
    aud: string;
    iat: bigint;
    exp: bigint;
}

class JWT {
    protected _default_aud: string = "https://diagramastext.dev";
    protected _default_iss: string = "https://ciam.diagramastext.dev";

    private readonly header: jwt_header;
    private readonly payload: jwt_payload;
    private readonly signature: string = "";

    constructor(v: string) {
        const el = v.split(".");

        if (el.length < 2) {
            throw new Error("invalid JWT")
        }

        this.header = JSON.parse(fromBase64(el[0]));
        this.payload = JSON.parse(fromBase64(el[1]));

        if (el.length < 3) {
            this.signature = el[3];
        }

        this._validate()
    }

    isExpired(): boolean {
        const nowSec = Date.now() / 1000;
        return this.payload.exp < nowSec || this.payload.exp <= this.payload.iat;
    }

    serialize(): string {
        let o = `${toBase64(JSON.stringify(this.header))}.${toBase64(JSON.stringify(this.payload))}`;
        if (this.signature !== "") {
            o += `.${this.signature}`
        }
        return o;
    }

    getQuotas(): user_quotas | undefined {
        return this.payload.quotas;
    }

    getFingerprint(): string {
        return this.payload.fingerprint;
    }

    getEmail(): string {
        return this.payload.email;
    }

    getSub(): string {
        return this.payload.sub;
    }

    private _validate() {
        if (this.header.alg !== "none" && this.signature === "") {
            throw new Error("invalid JWT: no signature found")
        }
        if (this.payload.aud !== this._default_aud) {
            throw new Error("invalid JWT: faulty aud")
        }
        if (this.payload.iss !== this._default_iss) {
            throw new Error("invalid JWT: faulty iss")
        }
    }
}

class Tokens {
    identity: JWT;
    access: JWT | null;
    refresh: JWT | null;

    constructor(rawTokensStr: string) {
        const tkn: rawTokens = JSON.parse(rawTokensStr);

        try {
            this.identity = new JWT(tkn.identity);
        } catch (e) {
            // @ts-ignore
            throw new Error(`identity token: ${e.message}`)
        }

        try {
            this.access = this._parseToken(tkn.access);
        } catch (e) {
            // @ts-ignore
            throw new Error(`access token: ${e.message}`)
        }

        try {
            this.refresh = this._parseToken(tkn.refresh);
        } catch (e) {
            // @ts-ignore
            throw new Error(`refresh token: ${e.message}`)
        }
    }

    private _parseToken(v: string): JWT | null {
        if (v !== "" && v !== null && v !== undefined) {
            return new JWT(v);
        }
        return null;
    }
}

type rawTokens = {
    identity: string;
    access: string;
    refresh: string;
};

export function fromBase64(v: string): string {
    return Buffer.from(v, "base64").toString("binary");
}

export function toBase64(v: string): string {
    return Buffer.from(v, "binary").toString("base64url");
}

export class CIAMClient {
    _cookie_tokens_key = "tokens";
    _cookie_tokens_exp_days = 7;
    private readonly _ciam_base_url: string;

    private readonly _fingerprint: string;
    protected _tokens: Tokens | undefined;
    readonly quotas: user_quotas;

    email: string = "";

    constructor(ciam_base_url: string) {
        this._ciam_base_url = ciam_base_url;

        const s = getCookie(this._cookie_tokens_key);
        if (s !== undefined) {
            this._tokens = new Tokens(s);
            if (this._tokens.access?.isExpired()) {
                this.refreshTokens()
            }
        }

        this.quotas = this._tokens?.access?.getQuotas()!;

        if (this.quotas === undefined) {
            this.quotas = {
                prompt_length_max: 100,
                rpm: 1,
                rpd: 1,
            };
        }

        // TODO: add verification of the fingerprint
        // if the current fingerprint does not match the one in token
        // the CIAM shall be called to update the fingerprint and issue new tokens,
        // or to create a new anonym user
        // @ts-ignore
        const userAgent = import.meta.env.DEV ? "NA" : navigator.userAgent;
        this._fingerprint = get_fingerprint(userAgent);
    }

    isAuth(): boolean {
        return this._tokens !== undefined;
    }

    // Implements the logic to signin anonym user.
    signInAnonym() {
        throw Error("method mot implemented")
    }

    // Implements the logic to initialise user registration.
    registerInit(email: string) {
        throw Error("method mot implemented")
    }

    // Implements the logic to confirm user registration.
    registerConfirm(secret_code: string) {
        throw Error("method mot implemented");
    }

    private refreshTokens() {
        throw Error("method mot implemented");
    }

    isRegistered(): boolean {
        return this._tokens?.identity.getEmail() !== "";
    }

    getId(): string {
        return this._tokens?.identity.getSub()!;
    }

    private setTokensCache(ciamResponseTokens: string) {
        setCookie(this._cookie_tokens_key, ciamResponseTokens, {
            expires: this._cookie_tokens_exp_days,
            sameSite: "strict",
            secure: true,
            path: "/",
            // TODO: add domain verification
        });
    }
}

/**
 * Generates a user's fingerprint to identify the session.
 * from: https://coursesweb.net/javascript/sha1-encrypt-data_cs
 *
 * @param {string} userAgent: user-agent string.
 * @return {string} fingerprint string.
 */
export function get_fingerprint(userAgent: string): string {
    if (userAgent === "") {
        return defaultNA;
    }

    function rotate_left(n: number, s: number): number {
        return (n << s) | (n >>> (32 - s));
    }

    function cvt_hex(val: number): string {
        var str = '';
        var i;
        var v;
        for (i = 7; i >= 0; i--) {
            v = (val >>> (i * 4)) & 0x0f;
            str += v.toString(16);
        }
        return str;
    }

    function Utf8Encode(string: string): string {
        string = string.replace(/\r\n/g, '\n');
        var utftext = '';
        for (var n = 0; n < string.length; n++) {
            var c = string.charCodeAt(n);
            if (c < 128) {
                utftext += String.fromCharCode(c);
            } else if ((c > 127) && (c < 2048)) {
                utftext += String.fromCharCode((c >> 6) | 192);
                utftext += String.fromCharCode((c & 63) | 128);
            } else {
                utftext += String.fromCharCode((c >> 12) | 224);
                utftext += String.fromCharCode(((c >> 6) & 63) | 128);
                utftext += String.fromCharCode((c & 63) | 128);
            }
        }
        return utftext;
    }

    var blockstart;
    var i, j;
    var W = new Array(80);
    var H0 = 0x67452301;
    var H1 = 0xEFCDAB89;
    var H2 = 0x98BADCFE;
    var H3 = 0x10325476;
    var H4 = 0xC3D2E1F0;
    var A, B, C, D, E;
    var temp: number;
    const msg: string = Utf8Encode(userAgent);
    var msg_len = msg.length;
    var word_array: number[] = new Array<number>();
    for (i = 0; i < msg_len - 3; i += 4) {
        j = msg.charCodeAt(i) << 24 | msg.charCodeAt(i + 1) << 16 | msg.charCodeAt(i + 2) << 8 | msg.charCodeAt(i + 3);
        word_array.push(j);
    }
    switch (msg_len % 4) {
        case 0:
            i = 0x080000000;
            break;
        case 1:
            i = msg.charCodeAt(msg_len - 1) << 24 | 0x0800000;
            break;
        case 2:
            i = msg.charCodeAt(msg_len - 2) << 24 | msg.charCodeAt(msg_len - 1) << 16 | 0x08000;
            break;
        case 3:
            i = msg.charCodeAt(msg_len - 3) << 24 | msg.charCodeAt(msg_len - 2) << 16 | msg.charCodeAt(msg_len - 1) << 8 | 0x80;
            break;
    }
    word_array.push(i);
    while ((word_array.length % 16) != 14) word_array.push(0);
    word_array.push(msg_len >>> 29);
    word_array.push((msg_len << 3) & 0x0ffffffff);
    for (blockstart = 0; blockstart < word_array.length; blockstart += 16) {
        for (i = 0; i < 16; i++) W[i] = word_array[blockstart + i];
        for (i = 16; i <= 79; i++) W[i] = rotate_left(W[i - 3] ^ W[i - 8] ^ W[i - 14] ^ W[i - 16], 1);
        A = H0;
        B = H1;
        C = H2;
        D = H3;
        E = H4;
        for (i = 0; i <= 19; i++) {
            temp = (rotate_left(A, 5) + ((B & C) | (~B & D)) + E + W[i] + 0x5A827999) & 0x0ffffffff;
            E = D;
            D = C;
            C = rotate_left(B, 30);
            B = A;
            A = temp;
        }
        for (i = 20; i <= 39; i++) {
            temp = (rotate_left(A, 5) + (B ^ C ^ D) + E + W[i] + 0x6ED9EBA1) & 0x0ffffffff;
            E = D;
            D = C;
            C = rotate_left(B, 30);
            B = A;
            A = temp;
        }
        for (i = 40; i <= 59; i++) {
            temp = (rotate_left(A, 5) + ((B & C) | (B & D) | (C & D)) + E + W[i] + 0x8F1BBCDC) & 0x0ffffffff;
            E = D;
            D = C;
            C = rotate_left(B, 30);
            B = A;
            A = temp;
        }
        for (i = 60; i <= 79; i++) {
            temp = (rotate_left(A, 5) + (B ^ C ^ D) + E + W[i] + 0xCA62C1D6) & 0x0ffffffff;
            E = D;
            D = C;
            C = rotate_left(B, 30);
            B = A;
            A = temp;
        }
        H0 = (H0 + A) & 0x0ffffffff;
        H1 = (H1 + B) & 0x0ffffffff;
        H2 = (H2 + C) & 0x0ffffffff;
        H3 = (H3 + D) & 0x0ffffffff;
        H4 = (H4 + E) & 0x0ffffffff;
    }

    const out: string = cvt_hex(H0) + cvt_hex(H1) + cvt_hex(H2) + cvt_hex(H3) + cvt_hex(H4);
    return out.toLowerCase();
}
