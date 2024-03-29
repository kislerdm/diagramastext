import {assert, describe, expect, it} from 'vitest'
// @ts-ignore
import {JSDOM} from 'jsdom';

import Main, {Disclaimer, Input, Output, PromptLengthLimit} from "./../src/main";
// @ts-ignore
import {arrow, box, boxText} from './../src/main.module.css';
import {Config} from "./../src/ports";

import {mockHTTPClient, mockResponse} from "./../src/httpclient";

function filterByClassName(elements: HTMLCollectionOf<Element>, className: string): Array<Element> {
    let o: Array<Element> = [];
    for (let i = 0; i < elements.length; i++) {
        if (elements[i].className == className) {
            o.push(elements[i]);
        }
    }
    return o;
}

describe("Main page component", () => {
    //GIVEN
    const mountPoint = new JSDOM("<main></main>").window.document.querySelector<HTMLDivElement>("main")!;

    const mockHttpClient = new mockHTTPClient(new mockResponse(200));

    const cfg: Config = {
        version: "foobar",
        urlAPI: "http://localhost:9000",
        promptMinLength: 3,
        promptMaxLengthUserBase: 100,
        promptMaxLengthUserRegistered: 300,
        httpClientSVGRendering: mockHttpClient,
    };

    //WHEN
    // @ts-ignore
    Main(mountPoint, cfg)

    //THEN
    const divs = mountPoint.getElementsByTagName("div"),
        boxes = filterByClassName(divs, box);

    it("shall contain three boxes", () => {
        assert.equal(boxes.length, 3)
    })

    it("shall have two popups mounted", () => {
        assert.equal(filterByClassName(divs, "modal")!.length, 2)
    })

    it("shall have header in the top", () => {
        assert.equal(mountPoint.children[0].tagName, "HEADER")
    })

    it("shall have punchline after the header", () => {
        assert(mountPoint.children[1].innerHTML.trim().startsWith("Generate"))
    })

    it("shall have the input box after the punchline", () => {
        assert.equal(boxes[0].getElementsByTagName("textarea").length, 1)
    })

    it("shall have the output box after the input", () => {
        assert.equal(boxes[1].getElementsByTagName("button")[0]!.innerHTML, "Download")
    })

    it("shall have the arrow between the input and output", () => {
        assert.equal(mountPoint.getElementsByTagName("i")[0]!.className, arrow)
    })

    it("shall have the disclaimer box after the output", () => {
        expect(boxes[2].innerHTML).toContain(`"A picture is worth a thousand words"`)
        expect(boxes[2].innerHTML).toContain(`<a href="mailto:contact@diagramastext.dev" target="_blank" rel="noopener">`)
    })

    it("shall have footer in the bottom", () => {
        assert.equal(mountPoint.children[mountPoint.children.length - 1].tagName, "FOOTER")
    })

    const btn = mountPoint.getElementsByTagName("button");
    it("shall have trigger btn to generate diagram", () => {
        assert.equal(btn[0]!.innerHTML, "Generate Diagram", "trigger button expected");
        expect(!btn[0]!.disabled, "trigger button is expected to be enabled")
    })

    it("shall have trigger btn to download generated diagram", () => {
        assert.equal(btn[1]!.innerHTML, "Download", "download button expected")
        expect(btn[1]!.disabled, "download button is expected to be disabled")
    })
})

describe("Input component", () => {
    //GIVEN
    const idTrigger = "foo";
    const idCounter = "bar";
    const promptLengthLimit = new PromptLengthLimit(1, 100);
    const placeholder = "qux";

    //WHEN
    const got = Input(idTrigger, idCounter, promptLengthLimit, placeholder);

    //THEN
    const el = new JSDOM(got).window.document.querySelector("div")!;

    it("shall have the class defined", () => {
        assert.equal(el.className, box)
    })

    it("shall have a single TEXTAREA element", () => {
        assert.equal(el.getElementsByTagName("textarea").length, 1)
    })

    it("shall have two DIV elements", () => {
        assert.equal(el.getElementsByTagName("div").length, 2)
    })

    it("shall have two P elements", () => {
        assert.equal(el.getElementsByTagName("p").length, 3)
    })

    it("shall have inline style", () => {
        assert.equal(el.getAttribute("style"), "margin-top:20px")
    })

    it("shall have four children", () => {
        assert.equal(el.children.length, 5)
    })

    it("shall have proper label", () => {
        assert.equal(el.children[0].textContent, "Input:", "unexpected label")
    })

    it("shall have the textarea", () => {
        assert.equal(el.children[1].tagName, "TEXTAREA", "input textarea is expected")
    })

    it("shall have the textarea:style", () => {
        const wantStyle = "font-size:20px;color:#fff;text-align:left;border-radius:1rem;padding:1rem;width:100%;background:#263950;box-shadow:0 0 3px 3px #2b425e"
        assert.equal(el.children[1].getAttribute("style"), wantStyle)
    })


    it("shall have the textarea:size-rows", () => {
        assert.equal(el.children[1].getAttribute("rows"), "3")
    })

    it("shall have the textarea:size-minlength", () => {
        assert.equal(el.children[1].getAttribute("minlength"), `${promptLengthLimit.Min}`)
    })

    it("shall have the textarea:size-maxlength", () => {
        const maxLengthMultiplier = 1.2;
        assert.equal(el.children[1].getAttribute("maxlength"),
            `${promptLengthLimit.Max * maxLengthMultiplier}`)
    })

    it("shall have the textarea:placeholder", () => {
        assert(el.children[1].getAttribute("placeholder"))
    })

    it("shall have the textarea:placeholder-predefined-input", () => {
        assert.equal(el.children[1].innerHTML, placeholder)
    })

    const promptLengthIndicator = el.children[3];
    it("shall have prompt length indicator", () => {
        assert(promptLengthIndicator.innerHTML.toLowerCase().includes("prompt length"))
    })

    it("shall have prompt length indicator:style", () => {
        const wantStyle = "color:white;text-align:right";
        assert.equal(promptLengthIndicator.getAttribute("style"), wantStyle)
    })

    const counter = promptLengthIndicator.children[0];
    it("shall have prompt length indicator:text prefix", () => {
        assert(counter.innerHTML.trim().startsWith("Prompt length:"))
    })

    const counterLengthIndicator = counter.getElementsByTagName("span")[0];
    it("shall have prompt length indicator:min:id", () => {
        assert.equal(counterLengthIndicator.id, idCounter)
    })

    it("shall have prompt length indicator:min:content", () => {
        assert.equal(counterLengthIndicator.innerHTML, `${placeholder.length}`)
    })

    it("shall have prompt length indicator:max:content", () => {
        assert(counter.innerHTML.trim().endsWith(`${promptLengthLimit.Max}`))
    })

    it("shall have prompt length indicator:slash-between-min-max", () => {
        assert(counter!.innerHTML.replace(`${promptLengthLimit.Max}`, "").trim().endsWith("/"))
    })

    const divs = el.getElementsByTagName("div"),
        divBtn = divs[divs.length - 1];
    it("shall have the trigger button in the bottom", () => {
        assert.equal(divBtn.children[0].tagName, "BUTTON")
        assert.equal(divBtn.children[0].id, idTrigger)
    })

    it("shall have the trigger button with defined id", () => {
        assert.equal(divBtn.children[0].id, idTrigger)
    })

    it("shall have the button's margins shrank", () => {
        assert.equal(divBtn.style.marginTop, "-20px")
    })

    const tipTriggerGeneration = el.children[2];
    it("shall have the tip on how to trigger generation", () => {
        assert.equal(tipTriggerGeneration.tagName, "P")
    })

    it("shall display 'Click the button, or press Ctrl(Control)+Enter' as a tip to trigger generation", () => {
        assert.equal(tipTriggerGeneration.innerHTML, "Click the button, or press Ctrl(Control)+Enter")
    })

    it("shall defines the style for the tip to trigger generation", () => {
        // @ts-ignore
        assert.equal(tipTriggerGeneration.style.color, "rgb(255, 255, 255)")
        // @ts-ignore
        assert.equal(tipTriggerGeneration.style.fontSize, "12px")
        // @ts-ignore
        assert.equal(tipTriggerGeneration.style.textAlign, "left")
        // @ts-ignore
        assert.equal(tipTriggerGeneration.style.marginRight, "0px")
        // @ts-ignore
        assert.equal(tipTriggerGeneration.style.marginLeft, "0px")
        // @ts-ignore
        assert.equal(tipTriggerGeneration.style.marginTop, "0px")
        // @ts-ignore
        assert.equal(tipTriggerGeneration.style.marginBottom, "-15px")
    })
})

describe("Output component", () => {
    //GIVEN
    const idOutput = "foo";
    const idDownload = "bar";
    const svg = `<svg><g><text>foo</text></g></svg>`;

    //WHEN
    const got = Output(idOutput, idDownload, svg);

    //THEN
    const el = new JSDOM(got).window.document.querySelector("div")!;

    it("shall have the class defined", () => {
        assert.equal(el.className, box)
    })

    it("shall have the margin-top set to 20px", () => {
        assert.equal(el.style.marginTop, "20px")
    })

    it("shall have the padding set to 20px", () => {
        assert.equal(el.style.padding, "20px")
    })

    it("shall have two DIV elements", () => {
        assert.equal(el.getElementsByTagName("div").length, 2)
    })

    const p = el.getElementsByTagName("p");
    it("shall have one P elements", () => {
        assert.equal(p.length, 1)
    })

    it("shall have the label 'Output:'", () => {
        assert.equal(p[0].innerHTML, "Output:")
    })

    it("shall have the class set to 'boxText' for p element", () => {
        assert.equal(p[0].className, boxText)
    })

    const divOutput = el.getElementsByTagName("div")[0]!;
    it("shall have the div element for generated diagram", () => {
        assert.equal(divOutput.id, idOutput)
    })

    it("shall have the border around generated diagram", () => {
        assert.equal(divOutput.style.borderStyle, "solid")
        assert.equal(divOutput.style.borderColor, "#2d4765")
        assert.equal(divOutput.style.borderWidth, "2px")
    })

    it("shall have the white background for generated diagram", () => {
        assert.equal(divOutput.style.background, "white")
    })

    it("shall have the shadow around generated diagram", () => {
        assert.equal(divOutput.style.boxShadow, "0 0 3px 3px #2b425e")
    })

    it("shall have inherited width for the diagram output", () => {
        assert.equal(divOutput.style.width, "inherit")
    })

    it("shall have a diagram svg as the diagram output", () => {
        assert.equal(
            new JSDOM(divOutput.innerHTML).window.document.getElementsByTagName("svg").length,
            1,
        )
    })

    const btn = el.getElementsByTagName("button");
    it("shall have one BUTTON element", () => {
        assert.equal(btn.length, 1)
    })

    it("shall have the button disabled", () => {
        assert(btn[0]!.disabled)
    })

    it("shall have the text 'Download' as the button label", () => {
        assert.equal(btn[0]!.textContent, "Download")
    })

    it("shall have specified id for the button element", () => {
        assert.equal(btn[0]!.id, idDownload)
    })

    const a = el.getElementsByTagName("a");
    it("shall have one A element", () => {
        assert.equal(a.length, 1)
    })

    it("shall have the A element without href", () => {
        assert.equal(a[0]!.href, "")
    })

    it("shall have the filename fixed as 'diagram.svg'", () => {
        assert.equal(a[0]!.download, "diagram.svg")
    })
})

describe("Disclaimer component", () => {
    const el = new JSDOM(Disclaimer).window.document.querySelector("div")!;

    it("shall have the class defined", () => {
        assert.equal(el.className, box)
    })

    it("shall have the margin-top set to 50px", () => {
        assert.equal(el.style.marginTop, "50px")
    })

    it("shall have the margin-bottom set to 20px", () => {
        assert.equal(el.style.marginBottom, "20px")
    })

    it("shall have the margin-left set to 0", () => {
        assert.equal(el.style.marginLeft, "0px")
    })

    it("shall have the margin-right set to 0", () => {
        assert.equal(el.style.marginRight, "0px")
    })

    it("shall have two P elements", () => {
        assert.equal(el.getElementsByTagName("p").length, 2)
    })

    it("shall have the feedback link", () => {
        expect(el.innerHTML).toContain("get in touch")
        expect(el.innerHTML).toContain(`<a href="mailto:`)
    })

    it("shall contain the quote-moto", () => {
        expect(el.innerHTML).toContain(`"A picture is worth a thousand words"`)
    })
})

// TODO: add UX tests with the server's mock:
// Happy path: diagram generated on the button click
// Unhappy path: invalid input prompt
// Unhappy path: popup on the server's error response
// describe("Happy path of the diagram generation logic", () => {
//     //GIVEN
//     const mountPoint = new JSDOM("<main></main>").window.document.querySelector<HTMLDivElement>("main")!;
//
//     //WHEN
//     // @ts-ignore
//     Main(mountPoint, config)
//
//     // input non-default prompt
//     mountPoint.getElementsByClassName(box)[0]!.getElementsByTagName("textarea")[0]!.innerHTML = "foobar";
//     const btnTrigger = [...mountPoint.getElementsByTagName("button")].find(el => el.id == "0")!;
//
//     //THEN
//     it("shall yield the svg output", () => {
//         assert.equal(btnTrigger!.innerHTML, "Generate Diagram", "unexpected button selected")
//         assert.equal(mountPoint.getElementsByClassName(box)[0]!.getElementsByTagName("textarea")[0]!.value,
//             "foobar", "unexpected input set")
//         // expect(mountPoint.getElementsByClassName(box)[1]!.getElementsByTagName("div")[0]!.innerHTML).toBe(MockSVGResponse.svg)
//     })
// })
