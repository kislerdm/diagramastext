import {assert, describe, expect, it, test} from 'vitest'

import {CIAMClient, FingerprintScanner, fromBase64, get_fingerprint, TokensStore} from './../src/ciam';
import {HTTPClient, mockHTTPClient, mockResponse} from "./../src/httpclient";

describe('fingerprint', () => {
    it('shall return a sha1 of a correct user-agent', () => {
        // GIVEN
        const userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36";
        const want = "9468a4a53a2f2fd9ea96db22dc9dd9bb6ce38b7c";
        // WHEN
        const got = get_fingerprint(userAgent);

        // THEN
        assert.equal(want, got, "unexpected result.")
    })
    it('shall return NA on empty input', () => {
        // GIVEN
        const userAgent = "";
        const want = "";
        // WHEN
        const got = get_fingerprint(userAgent);

        // THEN
        assert.equal(got, want, "unexpected result.")
    })
})

describe('fromBase64', () => {
    it('shall return a JSON string given non-empty encoded input string', () => {
        const s = "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0";
        const want = JSON.stringify({
            "alg": "none",
            "typ": "JWT",
        })

        assert.equal(fromBase64(s), want)
    })

    it('shall return an empty string given empty input', () => {
        assert.equal(fromBase64(""), "")
    })
})

function toBase64(v: string): string {
    return Buffer.from(v, "binary").toString("base64");
}

test('tracing: toBase64-fromBase64', () => {
    const s: string = "foobar"
    assert.equal(fromBase64(toBase64(s)), s)
})

class mockTokensStore implements TokensStore {
    s: string | undefined;
    path: string = "/";

    constructor(obj: string | undefined = undefined) {
        this.s = obj;
    }

    read(): string | undefined {
        return this.s;
    }

    write(value: string, path: string): void {
        this.s = value;
        this.path = path;
    }
}

class mockFingerprintScanner implements FingerprintScanner {
    s: string;

    constructor(s: string = "") {
        this.s = s;
    }

    scan(): string {
        return this.s;
    }
}

describe('CIAM', () => {
    // JWT info
    const sub: string = "userID";
    const exp: number = Date.now() + 3e10;
    const email: string = "foo@bar.baz";
    const fingerprint: string = "foo";
    const role: number = 0;
    const quotas = {
        prompt_length_max: 101,
        rpm: 1,
        rpd: 1,
    }

    // JWT mock
    const jwtHeader = "foo";

    const claimsStd = {
        sub: sub,
        exp: exp,
    };

    const tokenID = `${jwtHeader}.${toBase64(JSON.stringify({
        ...claimsStd,
        email: email,
        fingerprint: fingerprint,
    }))}`;

    const tokenRefresh = `${jwtHeader}.${toBase64(JSON.stringify(claimsStd))}`;
    const tokenAccess = `${jwtHeader}.${toBase64(JSON.stringify({
        ...claimsStd,
        role: role,
        quotas: quotas,
    }))}`;

    const baseURL: string = "https://foo.bar";

    describe('init', () => {
        const cookie = new mockTokensStore();
        const ciamClient = new CIAMClient(baseURL,
            cookie,
            new mockFingerprintScanner(fingerprint),
        );

        test('shall be not authorized', () => {
            expect(ciamClient.isAuth()).toStrictEqual(false);
        })

        test('shall return a placeholder empty object as access header', () => {
            expect(ciamClient.getHeaderAccess()).toStrictEqual({});
        })

        test('shall return a placeholder empty object as refresh header', () => {
            expect(ciamClient.getHeaderRefresh()).toStrictEqual({});
        })

        test('shall return default quotas', () => {
            expect(ciamClient.getQuotas()).toStrictEqual({
                prompt_length_max: 100,
                rpm: 1,
                rpd: 1,
            });
        })

        test('shall have empty cookie', () => {
            expect(cookie.read()).toBeUndefined()
        })
    })

    describe('sign-in anonym', async () => {
        const tokens = {
            id: tokenID,
            access: tokenAccess,
            refresh: tokenRefresh,
        };
        const mockHttp = new mockHTTPClient(
            new mockResponse(200, tokens),
        );
        const cookie = new mockTokensStore();
        const ciamClient = new CIAMClient(baseURL,
            cookie,
            new mockFingerprintScanner(fingerprint),
            mockHttp,
        );

        await ciamClient.signInAnonym()

        test('shall be authorized', () => {
            expect(ciamClient.isAuth()).toStrictEqual(true);
        })

        test('shall return valid access header', () => {
            expect(ciamClient.getHeaderAccess()).toStrictEqual({
                Authorization: `Bearer ${tokenAccess}`
            });
        })

        test('shall return valid refresh header', () => {
            expect(ciamClient.getHeaderRefresh()).toStrictEqual({
                Authorization: `Bearer ${tokenRefresh}`
            });
        })

        test('shall return valid quotas', () => {
            expect(ciamClient.getQuotas()).toStrictEqual(quotas);
        })

        const httpClient = mockHttp as mockHTTPClient;
        test(`shall call CIAM server to ${baseURL}/auth/anonym`, () => {
            expect(httpClient.input).toStrictEqual(`${baseURL}/auth/anonym`);
        })

        test(`shall call CIAM server to ${baseURL}/auth/anonym`, () => {
            expect(httpClient.input).toStrictEqual(`${baseURL}/auth/anonym`);
        })

        test('shall make a POST request to call CIAM server', () => {
            expect(httpClient.init?.method).toStrictEqual("POST");
        })

        test('shall include fingerprint to the body of the request to call CIAM server', () => {
            expect(JSON.parse(httpClient.init!.body!.toString())).toStrictEqual({fingerprint: fingerprint});
        })

        test(`shall store the CIAM response's tokens to cookie`, () => {
            expect(JSON.parse(cookie.read()!)).toStrictEqual(tokens)
        })
    })

    describe('refresh token', async () => {
        const tokensRefreshed = {
            id: tokenID,
            access: tokenAccess,
        };
        const mockHttp: HTTPClient = new mockHTTPClient(
            new mockResponse(200, tokensRefreshed),
        );

        const tokenAccessExpired = `${jwtHeader}.${toBase64(JSON.stringify({
            ...{...claimsStd, exp: claimsStd.exp - 10e10},
            role: role,
            quotas: quotas,
        }))}`;

        const cookie: TokensStore = new mockTokensStore(JSON.stringify({
            access: tokenAccessExpired,
            refresh: tokenRefresh,
        }));

        const ciamClient = new CIAMClient(baseURL, cookie, undefined, mockHttp);

        // expect to be not authorised before refresh
        expect(ciamClient.isAuth()).toStrictEqual(false);
        expect(ciamClient.isExp()).toStrictEqual(true);

        await ciamClient.refreshAccessToken()

        test('shall be authorized', () => {
            expect(ciamClient.isAuth()).toStrictEqual(true);
        })

        test('shall have valid access token prior to refresh', () => {
            expect(ciamClient.isExp()).toStrictEqual(false);
        })

        test('shall return valid access header', () => {
            expect(ciamClient.getHeaderAccess()).toStrictEqual({
                Authorization: `Bearer ${tokenAccess}`
            });
        })

        test('shall return valid refresh header', () => {
            expect(ciamClient.getHeaderRefresh()).toStrictEqual({
                Authorization: `Bearer ${tokenRefresh}`
            });
        })

        test('shall return valid quotas', () => {
            expect(ciamClient.getQuotas()).toStrictEqual(quotas);
        })

        const httpClient = mockHttp as mockHTTPClient;
        test(`shall call CIAM server to ${baseURL}/auth/refresh`, () => {
            expect(httpClient.input).toStrictEqual(`${baseURL}/auth/refresh`);
        })

        test('shall make a POST request to call CIAM server', () => {
            expect(httpClient.init?.method).toStrictEqual("POST");
        })

        test('shall include refresh_token to the body of the request to call CIAM server', () => {
            expect(JSON.parse(httpClient.init!.body!.toString())).toStrictEqual({refresh_token: tokenRefresh});
        })

        test(`shall store new tokens to cookie`, () => {
            const want = tokensRefreshed;
            Object.assign(want, {refresh: tokenRefresh});
            expect(JSON.parse(cookie.read()!)).toStrictEqual(want);
        })
    })
})
