import {beforeEach, describe, it} from 'vitest'
import {JSDOM} from 'jsdom';
import Footer from "./../src/components/footer";
import {equal} from "assert";

describe('Footer component with version', () => {
    //GIVEN
    const version = "foobar",
        wantCopyright = `diagramastext.dev  &copy; ${new Date().getFullYear().toString()}`,
        wantVersion = `version: ${version}`;

    let got: string;
    let footer: HTMLCollectionOf<HTMLElement>;
    beforeEach(() => {
        //WHEN
        got = Footer(version);
        footer = new JSDOM(got).window.document.getElementsByTagName("footer");
    })

    //THEN
    it('footer tag', () => {
        equal(footer.length, 1, "single footer element is expected")
        equal(footer[0].style.padding, "1rem", "unexpected footer padding")
    })

    it('p tags', () => {
        let pElements: HTMLCollectionOf<HTMLParagraphElement>;
        beforeEach(() => {
            pElements = footer[0].getElementsByTagName("p");
        })

        it('total count', () => {
            equal(pElements.length, 2, "footer shall contain two p elements");
        })

        it('copyright', () => {
            const copyright = pElements[0];
            equal(copyright!.style.fontSize, "16px", "unexpected fontsize of the copyright string")

            equal(copyright!.innerHTML, wantCopyright, "unexpected copyright content")
        })

        it('version', () => {
            const version = pElements[1];
            equal(version!.style.fontSize, "6px", "unexpected fontsize of the version string")

            equal(version!.innerHTML, wantVersion, "unexpected version string")
        })
    })
})

describe('Footer component without version', () => {
    //GIVEN
    const version = "",
        wantCopyright = `diagramastext.dev  &copy; ${new Date().getFullYear().toString()}`;

    let got: string;
    let footer: HTMLCollectionOf<HTMLElement>;
    beforeEach(() => {
        //WHEN
        got = Footer(version);
        footer = new JSDOM(got).window.document.getElementsByTagName("footer");
    })

    //THEN
    it('footer tag', () => {
        equal(footer.length, 1, "single footer element is expected")
        equal(footer[0].style.padding, "1rem", "unexpected footer padding")
    })

    it('p tags', () => {
        let pElements: HTMLCollectionOf<HTMLParagraphElement>;
        beforeEach(() => {
            pElements = footer[0].getElementsByTagName("p");
        })

        it('total count', () => {
            equal(pElements.length, 1, "footer shall contain one p element");
        })

        it('copyright', () => {
            const copyright = pElements[0];
            equal(copyright!.style.fontSize, "16px", "unexpected fontsize of the copyright string")

            equal(copyright!.innerHTML, wantCopyright, "unexpected copyright content")
        })
    })
})
