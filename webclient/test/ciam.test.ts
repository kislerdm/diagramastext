import {assert, describe, it, test} from 'vitest'

import {fromBase64, get_fingerprint, toBase64} from './../src/ciam';

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
        const want = "NA";
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

describe('toBase64', () => {
    it('shall return an encoded string given a JSON string input', () => {
        const s = JSON.stringify({
            "alg": "none",
            "typ": "JWT",
        })
        const want = "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0";

        assert.equal(toBase64(s), want)
    })

    it('shall return an empty string given empty input', () => {
        assert.equal(toBase64(""), "")
    })
})

test('tracing: toBase64-fromBase64', () => {
    const s = JSON.stringify({
        "alg": "none",
        "typ": "JWT",
    })
    assert.equal(fromBase64(toBase64(s)), s)
})
