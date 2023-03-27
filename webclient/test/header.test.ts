import {describe, it} from 'vitest'
import {equal} from "assert";
import {JSDOM} from 'jsdom';

import Header from "./../src/components/header";

describe('Header', () => {
    //WHEN
    const got = new JSDOM(Header).window.document.getElementsByTagName("header");

    //THEN
    it('header tag', () => {
        equal(got.length, 1, 'unexpected number of header elements');
    })

    const header = got[0];
    it('header style', () => {
        equal(header.style.padding, "1em", 'unexpected header padding')
        equal(header.style.width, "100%", 'unexpected header width')
    })

    const div = header.getElementsByTagName('div');
    it('div tag', () => {
        equal(div.length, 1, 'unexpected number of div children');
    })

    it('span tag', () => {
        const span = div[0]!.getElementsByTagName("span");
        equal(span[0]!.innerHTML, "Diagram", 'unexpected content of span[0]');
        equal(span[0]!.style.fontWeight, "bold", 'unexpected font weight of the span[0]');

        equal(span[1]!.innerHTML, " as ", 'unexpected content of span[1]');

        equal(span[2]!.innerHTML, "Text", 'unexpected content of span[2]');
        equal(span[2]!.style.fontWeight, "bold", 'unexpected font weight of the span[2]');
        equal(span[2]!.style.fontStyle, "italic", 'unexpected font style of the span[2]');
    })
})
