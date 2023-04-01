import {assert, describe, expect, it} from 'vitest'
// @ts-ignore
import {JSDOM} from 'jsdom';
import Main, {Input, PromptLengthLimit} from "./../src/main";
import {Config} from "./../src/ports";

function filterByClassName(elements: HTMLCollectionOf<Element>, className: string): Array<Element> {
    let o: Array<Element> = [];
    for (let i = 0; i < elements.length; i++) {
        if (elements[i].className == className) {
            o.push(elements[i]);
        }
    }
    return o;
}

describe("Main page structure", () => {
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

    it("should contain three boxes", () => {
        assert.equal(boxes.length, 3)
    })

    it("should have two popups mounted", () => {
        assert.equal(filterByClassName(divs, "modal")!.length, 2)
    })

    it("should have header in the top", () => {
        assert.equal(mountPoint.children[0].tagName, "HEADER")
    })

    it("should have punchline after the header", () => {
        assert(mountPoint.children[1].innerHTML.trim().startsWith("Generate"))
    })

    it("should have the input box after the punchline", () => {
        assert.equal(boxes[0].getElementsByTagName("textarea").length, 1)
    })

    it("should have the output box after the input", () => {
        assert.equal(boxes[1].getElementsByTagName("button")[0]!.innerHTML, "Download")
    })

    it("should have the arrow between the input and output", () => {
        assert.equal(mountPoint.getElementsByTagName("i")[0]!.className, "arrow")
    })

    it("should have the disclaimer box after the output", () => {
        assert(boxes[2].innerHTML.includes("Please get in touch"))
    })

    it("should have footer in the bottom", () => {
        assert.equal(mountPoint.children[mountPoint.children.length - 1].tagName, "FOOTER")
    })

    const btn = mountPoint.getElementsByTagName("button");
    it("should have trigger btn to generate diagram", () => {
        assert.equal(btn[0]!.innerHTML, "Generate Diagram", "trigger button expected");
        expect(!btn[0]!.disabled, "trigger button is expected to be enabled")
    })

    it("should have trigger btn to download generated diagram", () => {
        assert.equal(btn[1]!.innerHTML, "Download", "download button expected")
        expect(btn[1]!.disabled, "download button is expected to be disabled")
    })
})

describe("Input structure", () => {
    //GIVEN
    const idTrigger = "foo";
    const idCounter = "bar";
    const promptLengthLimit = new PromptLengthLimit(1, 100);
    const placeholder = "qux";

    //WHEN
    const got = Input(idTrigger, idCounter, promptLengthLimit, placeholder);

    //THEN
    const input = new JSDOM(got).window.document.querySelector("div")!;

    it("should have a single text element", () => {
        assert.equal(input.getElementsByTagName("textarea").length, 1)
    })

    it("should have two div elements", () => {
        assert.equal(input.getElementsByTagName("div").length, 2)
    })

    it("should have two p elements", () => {
        assert.equal(input.getElementsByTagName("p").length, 2)
    })

    it("should have inline style", () => {
        assert.equal(input.getAttribute("style"), "margin-top:20px")
    })

    it("should have four children", () => {
        assert.equal(input.children.length, 4)
    })

    it("should have proper label", () => {
        assert.equal(input.children[0].textContent, "Input:", "unexpected label")
    })

    it("should have the textarea", () => {
        assert.equal(input.children[1].tagName, "TEXTAREA", "input textarea is expected")
    })

    it("should have the textarea:style", () => {
        const wantStyle = "font-size:20px;color:#fff;text-align:left;border-radius:1rem;padding:1rem;width:100%;background:#263950;box-shadow:0 0 3px 3px #2b425e"
        assert.equal(input.children[1].getAttribute("style"), wantStyle)
    })


    it("should have the textarea:size-rows", () => {
        assert.equal(input.children[1].getAttribute("rows"), "3")
    })

    it("should have the textarea:size-minlength", () => {
        assert.equal(input.children[1].getAttribute("minlength"), `${promptLengthLimit.Min}`)
    })

    it("should have the textarea:size-maxlength", () => {
        const maxLengthMultiplier = 1.2;
        assert.equal(input.children[1].getAttribute("maxlength"),
            `${promptLengthLimit.Max * maxLengthMultiplier}`)
    })

    it("should have the textarea:placeholder", () => {
        assert(input.children[1].getAttribute("placeholder"))
    })

    it("should have the textarea:placeholder-predefined-input", () => {
        assert.equal(input.children[1].innerHTML, placeholder)
    })

    it("should have prompt length indicator", () => {
        assert(input.children[2].innerHTML.toLowerCase().includes("prompt length"))
    })

    it("should have prompt length indicator:style", () => {
        const wantStyle = "color:white;text-align:right";
        assert.equal(input.children[2].getAttribute("style"), wantStyle)
    })

    const counter = input.children[2].children[0];
    it("should have prompt length indicator:text prefix", () => {
        assert(counter.innerHTML.trim().startsWith("Prompt length:"))
    })

    const counterLengthIndicator = counter.getElementsByTagName("span")[0];
    it("should have prompt length indicator:min:id", () => {
        assert.equal(counterLengthIndicator.id, idCounter)
    })

    it("should have prompt length indicator:min:content", () => {
        assert.equal(counterLengthIndicator.innerHTML, `${placeholder.length}`)
    })

    it("should have prompt length indicator:max:content", () => {
        assert(counter.innerHTML.trim().endsWith(`${promptLengthLimit.Max}`))
    })

    it("should have prompt length indicator:slash-between-min-max", () => {
        assert(counter!.innerHTML.replace(`${promptLengthLimit.Max}`, "").trim().endsWith("/"))
    })

    const divs = input.getElementsByTagName("div"),
        divBtn = divs[divs.length - 1];
    it("should have the trigger button in the bottom", () => {
        assert(divBtn.children[0].tagName, "BUTTON")
        assert(divBtn.children[0].id, idTrigger)
    })

    it("should have the trigger button with defined id", () => {
        assert(divBtn.children[0].id, idTrigger)
    })

    it("should have the button's margins shrank", () => {
        assert(divBtn.style.marginTop, "-20px")
    })
})
