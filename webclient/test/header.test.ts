import {assert, describe, it} from 'vitest'
import {JSDOM} from 'jsdom';

import Header from "./../src/components/header";

describe("Header component", () => {
    const got = new JSDOM(Header).window.document.getElementsByTagName("header");

    it('shall include single header tag', () => {
        assert.equal(got.length, 1)
    })

    const header = got[0]
    it("shall have the padding of 1em", () => {
        assert.equal(header.style.padding, "1em")
    })

    it("shall have the width of 100%", () => {
        assert.equal(header.style.width, "100%")
    })

    const div = header.getElementsByTagName('div')
    it("shall have single div tag", () => {
        assert.equal(div.length, 1)
    })

    const span = div[0]!.getElementsByTagName("span")
    it("shall have three span tags", () => {
        assert.equal(span.length, 3)
    })

    it("shall have the span[0] with defined content", () => {
        assert.equal(span[0]!.innerHTML, "Diagram")
    })

    it("shall have the span[0] with the font set to bold", () => {
        assert.equal(span[0]!.style.fontWeight, "bold")
    })

    it("shall have the span[1] with defined content", () => {
        assert.equal(span[1]!.innerHTML, " as ")
    })

    it("shall have the span[2] with defined content", () => {
        assert.equal(span[2]!.innerHTML, "Text")
    })

    it("shall have the span[2] with the font set to bold italic", () => {
        assert.equal(span[2]!.style.fontWeight, "bold")
        assert.equal(span[2]!.style.fontStyle, "italic")
    })
})
