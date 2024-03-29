// @ts-ignore
import {arrow, box, boxText} from './main.module.css';

import Footer from "./components/footer";
import Header from "./components/header";
import {Loader, Popup} from "./components/popup";

import {Config, IsResponseError, IsResponseSVG} from "./ports";
import {CIAMClient} from "./ciam";

// @ts-ignore
import placeholderOutputSVG from "./components/svg/output-placeholder.svg?raw";
import MailToLinkStr from "./components/mailto";

function findElementByID(elements: HTMLCollectionOf<HTMLElement>, id: string): HTMLElement | undefined {
    for (let i = 0; i < elements.length; i++) {
        const element = elements.item(i);
        if (element != null && element.id == id) {
            return element;
        }
    }
    return undefined
}

export default function Main(mountPoint: HTMLDivElement, cfg: Config) {
    const placeholderInputPrompt = "C4 diagram of a Go web server reading from external Postgres database over TCP",
        id = {
            Trigger: "0",
            Output: "1",
            Download: "2",
            InputLengthCounter: "3"
        };

    const ciamClient = new CIAMClient(cfg.urlAPI, cfg.cookieStore, cfg.fingerprintScanner, cfg.httpClientCIAM);
    const promptLengthLimit = new PromptLengthLimit(
        cfg.promptMinLength, ciamClient.getQuotas().prompt_length_max,
    );

    mountPoint.innerHTML = `${Header}

<div style="font-size:30px;margin: 20px 0 10px">
    Generate <span style="font-weight:bold">diagrams</span> using 
    <span style="font-style:italic;font-weight:bold">plain English</span> in no time!
</div>

<div id="inpt">
${Input(id.Trigger, id.InputLengthCounter, promptLengthLimit, placeholderInputPrompt)}
</div>
<i class="${arrow}"></i>

${Output(id.Output, id.Download, placeholderOutputSVG)}

${Disclaimer}
<div>
    ${Popup.mount()}
    ${Loader.mount()}
</div>
${Footer(cfg.version)}
`;

    let svg = "";
    let firstTimeTriggered = true;

    // diagram generation flow
    let _fetchErrorCnt = 0;
    const _fetchErrorCntMax = 2;

    const inputBox: Element = mountPoint.getElementsByClassName(box)[0]!;
    const input: HTMLTextAreaElement = inputBox.getElementsByTagName("textarea")[0]!;
    const triggerBtn: HTMLElement = findElementByID(
        inputBox.getElementsByTagName("button"),
        id.Trigger,
    )!;

    const outputBox: Element = mountPoint.getElementsByClassName(box)[1]!;
    const output: HTMLElement = findElementByID(outputBox.getElementsByTagName("div"), id.Output)!;
    const downloadBtn: HTMLElement = findElementByID(
        outputBox!.getElementsByTagName("button"),
        id.Download,
    )!;


    const elapsedThresholdMS = 10000;
    let elapsedRequestThreshold: boolean = false;

    const controller = new AbortController();
    inputBox.addEventListener("keydown", (e) => {
        // @ts-ignore
        if (elapsedRequestThreshold && (e.key === "Escape" || e.key === "Esc")) {
            controller.abort();
        }
    });

    inputBox.addEventListener("keydown", (e) => {
        // @ts-ignore
        if (e.key === "Enter" && e.ctrlKey) {
            generateDiagram();
        }
    });

    triggerBtn.addEventListener("click", () => {
        generateDiagram();
    });

    async function generateDiagram() {
        function showError(status: number = 0, msg: string = "") {
            const errorMsg = _fetchErrorCnt >= _fetchErrorCntMax ? `The errors repreat, please
<a href="${generateFeedbackLink(prompt, cfg.version)}"
    target="_blank" rel="noopener" style="color:#3498db;font-weight:bold">report</a>` : mapStatusCode(status, msg);
            Popup.error(mountPoint, errorMsg);
        }

        //@ts-ignore
        const prompt = input!.value.trim();
        if (placeholderInputPrompt === prompt && firstTimeTriggered) {
            return;
        }

        firstTimeTriggered = false;
        const timeout = setTimeout(() => elapsedRequestThreshold = true, elapsedThresholdMS);

        Loader.show(mountPoint);

        // TODO: add popup for signup/signin using email
        try {
            if (!ciamClient.isAuth()) {
                await ciamClient.signInAnonym();
            }
            if (ciamClient.isExp()) {
                await ciamClient.refreshAccessToken();
            }
        } catch (e) {
            clearTimeout(timeout);
            elapsedRequestThreshold = false;
            Loader.hide(mountPoint);
            // @ts-ignore
            if (e.name === "AbortError") {
                Popup.show(mountPoint, "Request cancelled by client");
            } else {
                console.error(e);
                showError();
            }
        }

        const headers = {
            "Content-Type": "application/json",
        };
        Object.assign(headers, ciamClient.getHeaderAccess());

        cfg.httpClientSVGRendering.do(`${cfg.urlAPI}/generate/c4`, {
            method: "POST",
            headers: headers,
            body: JSON.stringify({
                "prompt": prompt,
            }),
            signal: controller.signal,
        }).then((resp: Response) => {
            if (!resp.ok && resp.status !== 429) {
                _fetchErrorCnt++;
            } else {
                _fetchErrorCnt = 0;
            }

            Loader.hide(mountPoint);
            resp.json()
                .then((data: any) => {
                    clearTimeout(timeout);
                    elapsedRequestThreshold = false;
                    if (IsResponseError(data)) {
                        showError(resp.status, data.error);
                    } else if (IsResponseSVG(data)) {
                        svg = scaleSVG(data.svg);
                        //@ts-ignore
                        output!.innerHTML = svg;
                        //@ts-ignore
                        downloadBtn!.disabled = false;
                    } else {
                        throw new Error("response data type not recognized")
                    }
                })
        }).catch((e) => {
            clearTimeout(timeout);
            elapsedRequestThreshold = false;
            Loader.hide(mountPoint);
            if (e.name === "AbortError") {
                Popup.show(mountPoint, "Request cancelled by client");
            } else {
                console.error(e);
                showError();
            }
        });
    }

    // download flow
    downloadBtn.addEventListener("click", () => {
        if (svg !== "") {
            const link = [...outputBox.getElementsByTagName("a")].find(link => link.download == "diagram.svg");
            link!.setAttribute("href", `data:image/svg+xml,${encodeURIComponent(svg)}`);
            link!.click();
        }
    })

    // input length counter update
    function readInputLength(input: HTMLTextAreaElement): number {
        // @ts-ignore
        return input!.value.trim().length;
    }

    input.addEventListener("input", () => {
        const l = readInputLength(input);
        const span = findElementByID(inputBox.getElementsByTagName("span"), id.InputLengthCounter)!;
        span.innerHTML = l.toString();
        span.style.color = "#fff";
        // @ts-ignore
        triggerBtn!.disabled = false;
        if (l > promptLengthLimit.Max || l < promptLengthLimit.Min) {
            span.style.color = "red";
            // @ts-ignore
            triggerBtn!.disabled = true;
        }
    })
}

export class PromptLengthLimit {
    Min: number
    Max: number

    constructor(min: number, max: number) {
        [min, max] = min < max ? [min, max] : [max, min];
        this.Min = min
        this.Max = max
    }
}

export function Input(idTrigger: string,
                      idCounter: string,
                      promptLengthLimit: PromptLengthLimit,
                      placeholder: string): string {
    function textAreaLengthMax(v: number | undefined | null): number {
        const multiplier = 1.2;
        const defaultMax = 100;
        if (v === undefined || v === null) {
            return defaultMax;
        }
        return Math.round(v * multiplier);
    }

    return `<div class="${box}" style="margin-top:20px">
    <p class="${boxText}">Input:</p>
    <textarea minlength=${promptLengthLimit.Min} maxlength=${textAreaLengthMax(promptLengthLimit.Max)} rows="3"
              style="font-size:20px;color:#fff;text-align:left;border-radius:1rem;padding:1rem;width:100%;background:#263950;box-shadow:0 0 3px 3px #2b425e"
              placeholder="Type in the diagram description">${placeholder}</textarea>
    <p style="color:#fff;font-size:12px;text-align:left;margin:0 0 -15px;">Click the button, or press Ctrl(Control)+Enter</p>
    <div style="color:white;text-align:right"><p>Prompt length: <span id="${idCounter}">${placeholder.length}</span> / ${promptLengthLimit.Max} </p></div>
    <div style="margin-top:-20px"><button id="${idTrigger}">Generate Diagram</button></div>
</div>
`
}

export function Output(idOutput: string, idDownload: string, svg: string): string {
    return `<div class="${box}" style="margin-top:20px;padding:20px">
    <p class="${boxText}">Output:</p>
    
    <div id="${idOutput}" 
    style="border:solid #2d4765 2px;background:white;box-shadow:0 0 3px 3px #2b425e;width:inherit"
>${svg}</div>
    
    <div>
        <button id="${idDownload}" disabled>Download</button>
        <a download="diagram.svg"></a>
    </div>
</div>
`
}

export const Disclaimer = `<div class="${box}" style="color:white;margin:50px 0 20px">
    <p>"A picture is worth a thousand words": diagram is a powerful conventional instrument to explain the
    meaning of complex systems, or processes. Unfortunately, substantial effort is required to develop and maintain
    a diagram. It impacts effectiveness of knowledge sharing, especially in software development. Luckily, <a
            href="https://openai.com/blog/best-practices-for-deploying-language-models/" target="_blank"
            rel="noopener noreffer">LLM</a> development reached such level when special skills are no longer needed
    to prepare standardised diagram in seconds!</p>
    
    <p>Please ${MailToLinkStr("get in touch")} for feedback and details about collaboration. Thanks!</p>
</div>`;

function generateFeedbackLink(prompt: string, version: string) {
    let url = "https://github.com/kislerdm/diagramastext/issues/new";
    const params = {
        assignee: "kislerdm",
        labels: ["feedback", "defect"],
        title: `Webclient issue`,
        body: `## Environment
- App version: ${version}

## Prompt

\`\`\`
${prompt}
\`\`\`

## Details

- Please describe your chain of actions, i.e. what preceded the state you report?
- Please attach screenshots whether possible

## Expected behaviour

Please describe what should have happened following the actions you described.
`,
    };
    //@ts-ignore
    const query = Object.keys(params).map(key => key + '=' + encodeURIComponent(params[key])).join('&');

    return `${url}?${query}`;
}

function mapStatusCode(status: number, msg: string): string {
    function fallback(fallback: string): string {
        if (msg.length > 0) {
            return msg;
        }
        return fallback;
    }

    switch (status) {
        case 400:
            return fallback("Model processing error");
        case 404:
            return fallback("Faulty path");
        case 422:
            return fallback("Faulty input");
        case 429:
            return fallback("Your quota exceeded, please get in touch to adjust it");
        default:
            return fallback("Unexpected error, please repeat request");
    }
}

/*
* scaleSVG scales SVG to fit the parent DIV.
* */
export function scaleSVG(svg: string): string {
    const parser = new DOMParser();
    let doc = parser.parseFromString(svg, "image/svg+xml")!.querySelector("svg");
    //@ts-ignore
    doc.style.preserveAspectRatio = "xMaxYMax";
    //@ts-ignore
    doc.style.width = "100%";
    //@ts-ignore
    doc.style.height = "100%";
    //@ts-ignore
    return new XMLSerializer().serializeToString(doc);
}
