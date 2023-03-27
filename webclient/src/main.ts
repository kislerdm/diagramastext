// @ts-ignore
import {arrow, box, boxText} from './main.module.css';
import Footer from "./components/footer";
import Header from "./components/header";
import {Config, DataSVG} from "@/ports";

// @ts-ignore
import placeholderOutputSVG from "./components/svg/output-placeholder.svg?raw";
// @ts-ignore
import logoGithub from "./components/svg/github.svg";
// @ts-ignore
import logoSlack from "./components/svg/slack.svg";
// @ts-ignore
import logoLinkedin from "./components/svg/linkedin.svg";
// @ts-ignore
import logoEmail from "./components/svg/email.svg";
import {User} from "./user";
import {Loader, Popup} from "./components/popup";

export default function Main(mountPoint: HTMLDivElement, cfg: Config) {
    const placeholderInputPrompt = "C4 diagram of a Go web server reading from external Postgres database over TCP",
        id = {
            Input: "0",
            Trigger: "1",
            Output: "2",
            Download: "4",
            InputLengthCounter: "5",
        };

    const user = new User();
    const promptLengthLimit = definePromptLengthLimit(cfg, user);

    mountPoint.innerHTML = `${Header}

<div style="font-size:30px;margin: 20px 0 10px">
    Generate <span style="font-weight:bold">diagrams</span> using 
    <span style="font-style:italic;font-weight:bold">plain English</span> in no time!
</div>

${Input(id.Input, id.Trigger, id.InputLengthCounter, promptLengthLimit, placeholderInputPrompt)}

<i class="${arrow}"></i>

${Output(id.Output, id.Download, placeholderOutputSVG)}

${Disclaimer}

${Footer(cfg.version)}
`;

    let svg = "";
    let firstTimeTriggered = true;

    const errorPopup = new Popup(mountPoint),
        loadingSpinner = new Loader(mountPoint);

    // diagram generation flow
    let _fetchErrorCnt = 0;
    const _fetchErrorCntMax = 2;
    document.getElementById(id.Trigger)!.addEventListener("click", () => {
        function showError(status: number) {
            _fetchErrorCnt++;
            const errorMsg = _fetchErrorCnt >= _fetchErrorCntMax ? `The errors repreat, please 
<a href="${generateFeedbackLink(prompt, cfg.version)}"
    target="_blank" rel="noopener" style="color:#3498db;font-weight:bold">report</a>` : mapStatusCode(status);
            errorPopup.error(errorMsg);
        }

        //@ts-ignore
        const prompt = document.getElementById(id.Input)!.value.trim();
        if (placeholderInputPrompt === prompt && firstTimeTriggered) {
            return;
        }
        firstTimeTriggered = false;
        loadingSpinner.show();
        fetch(cfg.urlAPI, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                "prompt": prompt,
            }),
        }).then((resp: Response) => {
            loadingSpinner.hide();
            if (!resp.ok) {
                showError(resp.status);
            } else {
                _fetchErrorCnt = 0;
                resp.json()
                    .then((data: DataSVG) => {
                        svg = scaleSVG(data!.svg);
                        //@ts-ignore
                        document.getElementById(id.Output).innerHTML = svg;
                        //@ts-ignore
                        document.getElementById(id.Download).disabled = false;
                    })
            }
        }).catch((e) => {
            console.error(e);
            loadingSpinner.hide();
            showError(0);
        });
    })

    // download flow
    document.getElementById(id.Download)!.addEventListener("click", () => {
        if (svg !== "") {
            download(svg)
        }
    })

    // // input length counter update
    function readInputLength(id: string): number {
        // @ts-ignore
        return document.getElementById(id)!.value.length;
    }

    document.getElementById(id.Input)!.addEventListener("input", () => {
        const l = readInputLength(id.Input);
        const span = document.getElementById(id.InputLengthCounter)!;
        span.innerHTML = l.toString();
        span.style.color = "#fff";
        // @ts-ignore
        document.getElementById(id.Trigger)!.disabled = false;
        if (l > promptLengthLimit.Max || l < promptLengthLimit.Min) {
            span.style.color = "red";
            // @ts-ignore
            document.getElementById(id.Trigger)!.disabled = true;
        }
    })
}

class PromptLengthLimit {
    Min: number
    Max: number

    constructor(min: number, max: number) {
        [min, max] = min < max ? [min, max] : [max, min];
        this.Min = min
        this.Max = max
    }
}

function definePromptLengthLimit(cfg: Config, user: User): PromptLengthLimit {
    if (user.is_registered()) {
        return new PromptLengthLimit(cfg.promptMinLength, cfg.promptMaxLengthUserRegistered)
    }
    return new PromptLengthLimit(cfg.promptMinLength, cfg.promptMaxLengthUserBase)
}

function Input(idInput: string,
               idTrigger: string,
               idCounter: string,
               promptLengthLimit: PromptLengthLimit,
               placeholder: string): string {
    function textAreaLengthMax(v: number): number {
        const multiplier = 1.2;
        return Math.round(v * multiplier);
    }

    return `<div class="${box}" style="margin-top:20px">
    <p class="${boxText}">Input:</p>
    <textarea id="${idInput}" 
              minlength=${promptLengthLimit.Min} maxlength=${textAreaLengthMax(promptLengthLimit.Max)} rows="3"
              style="font-size:20px;color:#fff;text-align:left;border-radius:1rem;padding:1rem;width:100%;background:#263950;box-shadow:0 0 3px 3px #2b425e"
              placeholder="Type in the diagram description">${placeholder}</textarea>
    <div style="color:white;text-align:right"><p>Prompt length: <span id="${idCounter}">${placeholder.length}</span> / ${promptLengthLimit.Max} </p></div>
    <div style="margin-top:-20px"><button id="${idTrigger}">Generate Diagram</button></div>
</div>
`
}

function Output(idOutput: string, idDownload: string, svg: string): string {
    return `
<div class="${box}" style="margin-top: 20px; padding: 20px;">
    <p class="${boxText}">Output:</p>
    <div id="${idOutput}" 
    style="border:solid #2d4765 2px;background:white;box-shadow:0 0 3px 3px #2b425e; width:inherit"
>${svg}</div>
    <div><button id="${idDownload}" disabled>Download</button></div>
</div>
`
}

const Disclaimer = `<div class="${box}" style="color:white;margin:50px 0 20px">
    <p>"A picture is worth a thousand words": diagram is a powerful conventional instrument to explain the
    meaning of complex systems, or processes. Unfortunately, substantial effort is required to develop and maintain
    a diagram. It impacts effectiveness of knowledge sharing, especially in software development. Luckily, <a
            href="https://openai.com/blog/best-practices-for-deploying-language-models/" target="_blank"
            rel="noopener noreffer">LLM</a> development reached such level when special skills are no longer needed
    to prepare standardised diagram in seconds!</p>
    
    <p>Please get in touch for feedback and details about collaboration. Thanks!</p>
    
    <a href="https://github.com/kislerdm/diagramastext"><img src="${logoGithub}" alt="github logo"/></a>
    <a href="https://join.slack.com/t/diagramastextdev/shared_invite/zt-1onedpbsz-ECNIfwjIj02xzBjWNGOllg">
        <img src="${logoSlack}" alt="slack logo"/>
    </a>
    <a href="https://www.linkedin.com/in/dkisler"><img src="${logoLinkedin}" alt="linkedin logo"/></a>
    <a href="mailto:hi@diagramastext.dev"><img src="${logoEmail}" alt="email logo"/></a>
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

function download(svg: string) {
    const link = document.createElement("a");
    link.setAttribute("download", "diagram.svg");
    link.setAttribute("href", `data:image/svg+xml,${encodeURIComponent(svg)}`);
    link.click();
}

function mapStatusCode(status: number) {
    switch (status) {
        case 400:
            return "Unexpected prompt length";
        case 404:
            return "Faulty path";
        case 429:
            return "The server is experiencing high load, please try later";
        default:
            return "Unexpected error, please repeat request";
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
