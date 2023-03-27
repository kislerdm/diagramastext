import { test, assert } from 'vitest'

import {get_fingerprint} from '@/user';

test('fingerprint: happy path', () => {
    // GIVEN
    const userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36";
    const want = "9468a4a53a2f2fd9ea96db22dc9dd9bb6ce38b7c";
    // WHEN
    const got = get_fingerprint(userAgent);

    // THEN
    assert.equal(want, got, "unexpected result.")
})

test('fingerprint: User-Agen unknown', () => {
    // GIVEN
    const userAgent = "";
    const want = "NA";
    // WHEN
    const got = get_fingerprint(userAgent);

    // THEN
    assert.equal(want, got, "unexpected result.")
})
