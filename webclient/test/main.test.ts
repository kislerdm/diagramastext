import {assert, describe, expect, it} from 'vitest'
// @ts-ignore
import {JSDOM} from 'jsdom';
import Main, {Input, PromptLengthLimit} from "./../src/main";
import {Config} from "./../src/ports";
import {equal} from "assert";

function filterByClassName(elements: HTMLCollectionOf<Element>, className: string): Array<Element> {
    let o: Array<Element> = [];
    for (let i = 0; i < elements.length; i++) {
        if (elements[i].className == className) {
            o.push(elements[i]);
        }
    }
    return o;
}

describe("Main page", () => {
    //GIVEN
    const mountPoint = new JSDOM("<main></main>").window.document.querySelector<HTMLDivElement>("main")!;
    const cfg: Config = {
        version: "foobar",
        urlAPI: "http://localhost:9000",
        promptMinLength: 3,
        promptMaxLengthUserBase: 100,
        promptMaxLengthUserRegistered: 300,
    };

    //WHEN
    // @ts-ignore
    Main(mountPoint, cfg)

    //THEN
    const divs = mountPoint.getElementsByTagName("div"),
        boxes = filterByClassName(divs, "box");

    it("should contain tags", () => {
        equal(mountPoint.getElementsByTagName("header").length, 1, "single 'header' element expected")
        equal(mountPoint.getElementsByTagName("footer").length, 1, "single 'footer' element expected")
        equal(mountPoint.getElementsByTagName("span").length, 7, "7 'span' elements expected")
        equal(mountPoint.getElementsByTagName("a").length, 10, "10 'a' elements expected")

        equal(divs.length, 16, "16 'div' elements expected")
        equal(boxes.length, 3, "three boxes expected")
        equal(filterByClassName(divs, "modal")!.length, 2, "two popups expected")

        // input
        equal(mountPoint.getElementsByTagName("textarea").length, 1, "single textarea expected")

        // output
        equal(mountPoint.getElementsByTagName("svg").length, 1, "svg output expected")
    })

    it("should have the elements ordered", () => {
        equal(mountPoint.children[0].tagName, "HEADER", "header is expected to be in the top")
        assert(mountPoint.children[1].innerHTML.trim().startsWith("Generate"),
            "the punchline is expected to be right after the header")

        equal(mountPoint.children[mountPoint.children.length - 1].tagName, "FOOTER",
            "header is expected to be in the bottom",
        )
        equal(boxes[0].getElementsByTagName("textarea").length, 1,
            "input box is expected to be the first",
        )
        equal(boxes[1].getElementsByTagName("button")[0]!.innerHTML, "Download",
            "output box is expected to be the second",
        )
        assert(boxes[2].innerHTML.includes("Please get in touch"), "disclaimer box is expected to be the third")
    })

    it("should have the buttons", () => {
        const btn = mountPoint.getElementsByTagName("button");

        equal(btn.length, 2)

        equal(btn[0]!.innerHTML, "Generate Diagram", "trigger button expected");
        expect(!btn[0]!.disabled, "trigger button is expected to be enabled")

        equal(btn[1]!.innerHTML, "Download", "download button expected")
        expect(btn[1]!.disabled, "download button is expected to be disabled")
    })
})

describe("Input box", () => {
    //GIVEN
    const idTrigger = "foo";
    const idCounter = "bar";
    const promptLengthLimit = new PromptLengthLimit(10, 100);
    const placeholder = "qux";

    //WHEN
    const got = Input(idTrigger, idCounter, promptLengthLimit, placeholder);

    //THEN
    const input = new JSDOM(got).window.document.querySelector("div")!;

    it("should have inline style", () => {
        equal(input.getAttribute("style"), "margin-top:20px")
    })

    it("should have four children", () => {
        equal(input.children.length, 4)
    })

    it("should have proper label", () => {
        equal(input.children[0].textContent, "Input:", "unexpected label")
    })

    it("should have the textarea", () => {
        equal(input.children[1].tagName, "TEXTAREA", "input textarea is expected")
    })

    it("should have the textarea:style", () => {
        const wantStyle = "font-size:20px;color:#fff;text-align:left;border-radius:1rem;padding:1rem;width:100%;background:#263950;box-shadow:0 0 3px 3px #2b425e"
        equal(input.children[1].getAttribute("style"), wantStyle, "unexpected style")
    })


    it("should have the textarea:size-rows", () => {
        equal(input.children[1].getAttribute("rows"), 3, "unexpected number textarea rows")
    })

    it("should have the textarea:size-minlength", () => {
        equal(input.children[1].getAttribute("minlength"), promptLengthLimit.Min, "unexpected minlength")
    })

    it("should have the textarea:size-maxlength", () => {
        const maxLengthMultiplier = 1.2;
        equal(input.children[1].getAttribute("maxlength"), promptLengthLimit.Max * maxLengthMultiplier,
            "unexpected maxlength")
    })

    it("should have the textarea:placeholder", () => {
        equal(input.children[1].getAttribute("placeholder"), "Type in the diagram description")
    })

    it("should have the textarea:placeholder-predefined-input", () => {
        equal(input.children[1].innerHTML, placeholder)
    })

    it("should have prompt length indicator", () => {
        assert(input.children[2].innerHTML.toLowerCase().includes("prompt length"))
    })

    it("should have prompt length indicator:style", () => {
        const wantStyle = "color:white;text-align:right";
        equal(input.children[2].getAttribute("style"), wantStyle, "unexpected style")
    })

    const counter = input.children[2].children[0];
    assert(counter.innerHTML.trim().startsWith("Prompt length:"), "unexpected counter")
})