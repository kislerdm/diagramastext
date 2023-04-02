import {assert, describe, it} from 'vitest'
// @ts-ignore
import {JSDOM} from 'jsdom';

import MailToLinkStr from "./../../src/components/mailto";

describe("GIVEN: no innerHTML provided", () => {
    // WHEN
    const got = MailToLinkStr();

    // THEN
    const gotDom = new JSDOM(got).window.document;
    const tags = gotDom.getElementsByTagName("a");
    it("is expected to generate one A tag", () => {
        assert.equal(tags.length, 1)
    })

    it("is expected that the link points to the mailto", () => {
        assert.equal(tags[0].href, "mailto:contact@diagramastext.dev")
    })

    it("is expected that the A tag includes target _blank", () => {
        assert.equal(tags[0].target, "_blank")
    })

    it("is expected that the A tag includes rel noopener", () => {
        assert.equal(tags[0].rel, "noopener")
    })
})

describe("GIVEN: innerHTML", () => {
    const wantInnerHTML = "foo"

    // WHEN
    const got = MailToLinkStr(wantInnerHTML);

    // THEN
    it("is expected that the A tag includes provided innerHTML string", () => {
        const link = new JSDOM(got).window.document.getElementsByTagName("a")[0]!;
        assert.equal(link.innerHTML, wantInnerHTML)
    })
})
