import {getCookie, setCookie} from "typescript-cookie";
import {v} from "msw/lib/glossary-de6278a9";

const defaultNA: string = "";

type user_quotas = {
    prompt_length_max: number;
    rpm: number;
    rpd: number;
}

const defaultQuotas = {
    prompt_length_max: 100,
    rpm: 1,
    rpd: 1,
};

type claimsStd = {
    sub: string;
    exp: number;
}

type claimsRefresh = claimsStd;

type claimsID = claimsStd & {
    email: string;
    fingerprint: string;
}

type claimsAccess = claimsStd & {
    role: number;
    quotas: user_quotas;
}

export function fromBase64(v: string): string {
    return Buffer.from(v, "base64").toString("binary");
}

interface Tokens {
    id: string | undefined;
    access: string | undefined;
    refresh: string | undefined;

    setID(s: string): void;

    setAccess(s: string): void;

    setRefresh(s: string): void;

    userID(): string | undefined;

    quotas(): user_quotas | undefined;

    stringify(): string;

    isAuthUser(): boolean;
}

class tokens implements Tokens {
    id: string | undefined;
    access: string | undefined;
    refresh: string | undefined;

    private _id: claimsID | undefined;
    private _access: claimsAccess | undefined;
    private _refresh: claimsRefresh | undefined;

    constructor(s: string | undefined) {
        if (s === undefined || s === "") {
            return
        }

        const tkns = JSON.parse(s);
        if (tkns["id"] !== undefined) {
            this.setID(tkns["id"]);
        }

        if (tkns["access"] !== undefined) {
            this.setAccess(tkns["access"])
        }

        if (tkns["refresh"] !== undefined) {
            this.setRefresh(tkns["refresh"])
        }
    }

    setID(s: string): void {
        this.id = s;
        this._id = cast(this.parseJWTClaims(this.id), r("claimsID"));
    }

    setAccess(s: string): void {
        this.access = s;
        this._access = cast(this.parseJWTClaims(this.access), r("claimsAccess"));
    }

    setRefresh(s: string): void {
        this.refresh = s;
        this._refresh = cast(this.parseJWTClaims(this.refresh), r("claimsRefresh"));
    }

    stringify(): string {
        return JSON.stringify({"id": this.id, "access": this.access, "refresh": this.refresh})
    }

    private parseJWTClaims(s: string): any {
        const els = s.split(".");
        if (els.length < 2) {
            throw new Error("faulty JWT format")
        }
        return JSON.parse(fromBase64(els[1]));
    }

    quotas(): user_quotas {
        if (this._access !== undefined) {
            return this._access.quotas;
        }
        return defaultQuotas;
    }

    userID(): string | undefined {
        return this._id?.sub;
    }

    isAuthUser(): boolean {
        return this._access?.exp! > Date.now();
    }
}

export class CIAMClient {
    _cookie_tokens_key: string = "tokens";
    _cookie_tokens_exp_days: number = 100;
    private readonly _ciam_base_url: string;

    private readonly _fingerprint: string;

    readonly tokens: tokens;

    constructor(ciam_base_url: string) {
        this._ciam_base_url = ciam_base_url;
        this.tokens = new tokens(getCookie(this._cookie_tokens_key));

        // TODO: add verification of the fingerprint
        // if the current fingerprint does not match the one in token
        // the CIAM shall be called to update the fingerprint and issue new tokens,
        // or to create a new anonym user
        // @ts-ignore
        const userAgent: string = import.meta.env.DEV ? "NA" : navigator.userAgent;
        this._fingerprint = get_fingerprint(userAgent);
    }

    isAuth(): boolean {
        return this.tokens.isAuthUser();
    }

    // Implements the logic to signin anonym user.
    signInAnonym(): void {
        fetch(`${this._ciam_base_url}/auth/anonym`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                "fingerprint": this._fingerprint,
            }),
        }).then(resp => {
            if (resp.status !== 200) {
                throw new Error("error auth anonym")
            }
            return resp.json()
        }).then(data => {
            this.tokens.setID(data["id"])
            this.tokens.setAccess(data["access"])
            this.tokens.setRefresh(data["refresh"])
        })
    }

    private setTokensCache(): void {
        setCookie(this._cookie_tokens_key, this.tokens?.stringify(), {
            expires: this._cookie_tokens_exp_days,
            sameSite: "strict",
            secure: true,
            path: "/",
            // TODO: add domain verification
        });
    }

    getHeaderAccess(): Object {
        if (this.tokens.access === "") {
            return {}
        }
        return {
            "Authorization": `Bearer ${this.tokens.access}`
        }
    }

    getHeaderRefresh(): Object {
        if (this.tokens.refresh === "") {
            return {}
        }
        return {
            "Authorization": `Bearer ${this.tokens.refresh}`
        }
    }
}

/**
 * GENERATED by https://app.quicktype.io/
 * */
function invalidValue(typ: any, val: any, key: any, parent: any = ''): never {
    const prettyTyp = prettyTypeName(typ);
    const parentText = parent ? ` on ${parent}` : '';
    const keyText = key ? ` for key "${key}"` : '';
    throw Error(`Invalid value${keyText}${parentText}. Expected ${prettyTyp} but got ${JSON.stringify(val)}`);
}

function prettyTypeName(typ: any): string {
    if (Array.isArray(typ)) {
        if (typ.length === 2 && typ[0] === undefined) {
            return `an optional ${prettyTypeName(typ[1])}`;
        } else {
            return `one of [${typ.map(a => {
                return prettyTypeName(a);
            }).join(", ")}]`;
        }
    } else if (typeof typ === "object" && typ.literal !== undefined) {
        return typ.literal;
    } else {
        return typeof typ;
    }
}

function jsonToJSProps(typ: any): any {
    if (typ.jsonToJS === undefined) {
        const map: any = {};
        typ.props.forEach((p: any) => map[p.json] = {key: p.js, typ: p.typ});
        typ.jsonToJS = map;
    }
    return typ.jsonToJS;
}

function transform(val: any, typ: any, getProps: any, key: any = '', parent: any = ''): any {
    function transformPrimitive(typ: string, val: any): any {
        if (typeof typ === typeof val) return val;
        return invalidValue(typ, val, key, parent);
    }

    function transformUnion(typs: any[], val: any): any {
        // val must validate against one typ in typs
        const l = typs.length;
        for (let i = 0; i < l; i++) {
            const typ = typs[i];
            try {
                return transform(val, typ, getProps);
            } catch (_) {
            }
        }
        return invalidValue(typs, val, key, parent);
    }

    function transformEnum(cases: string[], val: any): any {
        if (cases.indexOf(val) !== -1) return val;
        return invalidValue(cases.map(a => {
            return l(a);
        }), val, key, parent);
    }

    function transformArray(typ: any, val: any): any {
        // val must be an array with no invalid elements
        if (!Array.isArray(val)) return invalidValue(l("array"), val, key, parent);
        return val.map(el => transform(el, typ, getProps));
    }

    function transformDate(val: any): any {
        if (val === null) {
            return null;
        }
        const d = new Date(val);
        if (isNaN(d.valueOf())) {
            return invalidValue(l("Date"), val, key, parent);
        }
        return d;
    }

    function transformObject(props: { [k: string]: any }, additional: any, val: any): any {
        if (val === null || typeof val !== "object" || Array.isArray(val)) {
            return invalidValue(l(ref || "object"), val, key, parent);
        }
        const result: any = {};
        Object.getOwnPropertyNames(props).forEach(key => {
            const prop = props[key];
            const v = Object.prototype.hasOwnProperty.call(val, key) ? val[key] : undefined;
            result[prop.key] = transform(v, prop.typ, getProps, key, ref);
        });
        Object.getOwnPropertyNames(val).forEach(key => {
            if (!Object.prototype.hasOwnProperty.call(props, key)) {
                result[key] = transform(val[key], additional, getProps, key, ref);
            }
        });
        return result;
    }

    if (typ === "any") return val;
    if (typ === null) {
        if (val === null) return val;
        return invalidValue(typ, val, key, parent);
    }
    if (typ === false) return invalidValue(typ, val, key, parent);
    let ref: any = undefined;
    while (typeof typ === "object" && typ.ref !== undefined) {
        ref = typ.ref;
        typ = typeMap[typ.ref];
    }
    if (Array.isArray(typ)) return transformEnum(typ, val);
    if (typeof typ === "object") {
        return typ.hasOwnProperty("unionMembers") ? transformUnion(typ.unionMembers, val)
            : typ.hasOwnProperty("arrayItems") ? transformArray(typ.arrayItems, val)
                : typ.hasOwnProperty("props") ? transformObject(getProps(typ), typ.additional, val)
                    : invalidValue(typ, val, key, parent);
    }
    // Numbers can be parsed by Date but shouldn't be.
    if (typ === Date && typeof val !== "number") return transformDate(val);
    return transformPrimitive(typ, val);
}

function cast<T>(val: any, typ: any): T {
    return transform(val, typ, jsonToJSProps);
}

function l(typ: any) {
    return {literal: typ};
}

function a(typ: any) {
    return {arrayItems: typ};
}

function u(...typs: any[]) {
    return {unionMembers: typs};
}

function o(props: any[], additional: any) {
    return {props, additional};
}

function r(name: string) {
    return {ref: name};
}

const typeMap: any = {
    "Graph": o([
        {json: "links", js: "links", typ: u(undefined, a(r("Link")))},
        {json: "nodes", js: "nodes", typ: a(r("Node"))},
    ], "any"),
};

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
