import {assert, describe, it} from 'vitest'
// @ts-ignore
import {JSDOM} from 'jsdom';
import Footer from "./../src/components/footer";
import {equal} from "assert";

describe('Footer component with version', () => {
    //GIVEN
    const version = "foobar",
        wantCopyrightFn = (s: string): boolean => s.startsWith("diagramastext.dev"),
        wantVersion = `version: ${version}`;

    //WHEN
    const got = Footer(version);

    //THEN
    const footer = new JSDOM(got).window.document.getElementsByTagName("footer");
    it('footer tag', () => {
        equal(footer.length, 1, "single footer element is expected")
        equal(footer[0].style.padding, "1rem", "unexpected footer padding")
    })

    const pElements = footer[0].getElementsByTagName("p");
    it('p tags - total count', () => {
        equal(pElements.length, 2, "footer shall contain two p elements");
    })

    it('p tags - copyright', () => {
        const copyright = pElements[0];
        equal(copyright!.style.fontSize, "16px", "unexpected fontsize of the copyright string")
        assert(wantCopyrightFn(copyright!.innerHTML), "unexpected copyright content")
    })

    it('p tags - version', () => {
        const version = pElements[1];
        equal(version!.style.fontSize, "6px", "unexpected fontsize of the version string")
        equal(version!.innerHTML, wantVersion, "unexpected version string")
    })
})

describe('Footer component without version', () => {
    //GIVEN
    const version = "",
        wantCopyrightFn = (s: string): boolean => s.startsWith("diagramastext.dev");

    //WHEN
    const got = Footer(version);

    //THEN
    const footer = new JSDOM(got).window.document.getElementsByTagName("footer");
    it('footer tag', () => {
        equal(footer.length, 1, "single footer element is expected")
        equal(footer[0].style.padding, "1rem", "unexpected footer padding")
    })

    const pElements = footer[0].getElementsByTagName("p");
    it('p tags - total count', () => {
        equal(pElements.length, 1, "footer shall contain two p elements");
    })

    it('p tags - copyright', () => {
        const copyright = pElements[0];
        equal(copyright!.style.fontSize, "16px", "unexpected fontsize of the copyright string")
        assert(wantCopyrightFn(copyright!.innerHTML), "unexpected copyright content")
    })
})
