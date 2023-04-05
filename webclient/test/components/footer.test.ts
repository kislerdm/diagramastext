import {assert, describe, it} from 'vitest'
// @ts-ignore
import {JSDOM} from 'jsdom';

import Footer from "./../../src/components/footer";

describe("Footer component", () => {
    const got = new JSDOM(Footer()).window.document.getElementsByTagName("footer");

    it("shall include single footer tag", () => {
        assert.equal(got.length, 1)
    })

    it("shall have the padding of 1rem", () => {
        assert.equal(got[0]!.style.padding, "1rem")
    })

    const pElements = got[0]!.getElementsByTagName("p");
    it("shall have p tags", () => {
        assert(pElements.length > 0)
    })

    const copyright = pElements[0];
    it("shall include copyright which starts with the text diagramastext.dev", () => {
        assert(copyright!.innerHTML.startsWith("diagramastext.dev"))
    })

    it("shall include copyright which ends with the current year", () => {
        assert(copyright!.innerHTML.endsWith(`${new Date().getFullYear().toString()}`))
    })

    it("shall include copyright displayed as the text of 16px font size", () => {
        assert.equal(copyright!.style.fontSize, "16px")
    })

    const socialContact = pElements[1];
    it("shall include social medial links container", () => {
        assert.equal(socialContact.id, "contacts")
    })

    it("shall include social medial links container with margins-top:-10px", () => {
        assert.equal(socialContact.style.marginTop, "-10px")
    })

    const links = pElements[1].getElementsByTagName("a");
    it("shall include social medial with four links", () => {
        assert.equal(links.length, 4)
    })

    it("shall include link to github, leftmost", () => {
        assert.equal(links[0].href, "https://github.com/kislerdm/diagramastext")
    })

    it("shall include link to slack, after github", () => {
        assert(links[1].href.startsWith("https://join.slack.com"))
    })

    it("shall include link to linkedin, after slack", () => {
        assert.equal(links[2].href, "https://www.linkedin.com/in/dkisler")
    })

    it("shall include link to email, after linkedin", () => {
        assert.equal(links[3].href, "mailto:contact@diagramastext.dev")
    })

    const icons = pElements[1].getElementsByTagName("img");
    it("shall include social medial icons as links", () => {
        assert.equal(icons.length, 4)
    })

    it("shall include social medial icons: github is leftmost", () => {
        assert.equal(icons[0].alt, "github")
    })

    it("shall include social medial icons: slack is after github", () => {
        assert.equal(icons[1].alt, "slack")
    })

    it("shall include social medial icons: linkedin is after slack", () => {
        assert.equal(icons[2].alt, "linkedin")
    })

    it("shall include social medial icons: email is after linkedin", () => {
        assert.equal(icons[3].alt, "email")
    })

    it("shall include social medial icons with the width of 20px", () => {
        for (let i = 0; i < icons.length; i++) {
            assert.equal(icons[i].width, 20)
        }
    })
})

describe("Footer component with defined version", () => {
    //GIVEN
    const version = "foobar",
        wantVersion = `version: ${version}`;

    //WHEN
    const got = Footer(version);

    //THEN
    const pElements = new JSDOM(got).window.document.getElementsByTagName("footer")[0]!
            .getElementsByTagName("p"),
        versionElement = pElements[pElements.length - 1];

    it("shall include specified version beneath all elements", () => {
        assert.equal(versionElement!.innerHTML, wantVersion)
    })

    it("shall include the version displayed as the text of 6px font size", () => {
        assert.equal(versionElement!.style.fontSize, "6px")
    })
})
