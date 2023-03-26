// @ts-ignore
import {arrow, box, boxText, punch} from './main.module.css';
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
        };

    mountPoint.innerHTML = `${Header}

<div style="font-size:30px;margin: 20px 0 10px">
    Generate <span style="font-weight:bold">diagrams</span> using 
    <span style="font-style:italic;font-weight:bold">plain English</span> in no time!
</div>

${Input(id.Input, id.Trigger, cfg.promptMinLength, cfg.promptMaxLengthUserRegistered, placeholderInputPrompt)}

<i class="${arrow}"></i>

${Output(id.Output, id.Download, placeholderOutputSVG)}

${Disclaimer}

${Footer(cfg.version)}
`;

    let svg = "";
    const errorPopup = new Popup(mountPoint),
        loadingSpinner = new Loader(mountPoint);

    const user = new User();

    // diagram generation flow
    let _fetchErrorCnt = 0;
    const _fetchErrorCntMax = 2;
    document.getElementById(id.Trigger)!.addEventListener("click", () => {
        //@ts-ignore
        const prompt = document.getElementById(id.Input)!.value.trim();

        if (placeholderInputPrompt !== prompt) {
            try {
                validatePrompt(prompt, user, cfg);
                // @ts-ignore
            } catch (e: Error) {
                errorPopup.error(e.message);
                return;
            }

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
                    _fetchErrorCnt++;

                    const errorMsg = _fetchErrorCnt > _fetchErrorCntMax ? `The errors repreat, please 
<a href="${generateFeedbackLink(prompt, cfg.version)}"
    target="_blank" rel="noopener" style="color:#3498db;font-weight:bold">report</a>` : mapStatusCode(resp.status);

                    errorPopup.error(errorMsg);
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
            })
        }
    })

    // download flow
    document.getElementById(id.Download)!.addEventListener("click", () => {
        if (svg !== "") {
            download(svg)
        }
    })
}

function Input(idInput: string, idTrigger: string, minLength: number, maxLength: number, placeholder: string): string {
    return `<div class="${box}" style="margin-top:20px">
    <p class="${boxText}">Input:</p>
    <textarea id="${idInput}" 
              minlength=${minLength} maxlength=${maxLength} rows="3"
              style="font-size:20px;color:#fff;text-align:left;border-radius:1rem;padding:1rem;width:100%;background:#263950;box-shadow:0 0 3px 3px #2b425e"
              placeholder="Type in the diagram description">${placeholder}</textarea>
    <div><button id="${idTrigger}">Generate Diagram</button></div>
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

function validatePromptLength(prompt: string, lengthMin: number, lengthMax: number) {
    if (prompt.length < lengthMin || prompt.length > lengthMax) {
        throw new RangeError(`The prompt must be between ${lengthMin} and ${lengthMax} characters long`)
    }
}

function validatePrompt(prompt: string, user: User, cfg: Config) {
    switch (user.is_registered()) {
        case true:
            validatePromptLength(prompt, cfg.promptMinLength, cfg.promptMaxLengthUserRegistered);
            break;
        default:
            validatePromptLength(prompt, cfg.promptMinLength, cfg.promptMaxLengthUserBase);
            break;
    }
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
            return "Unexpected error, please try later";
    }
}

/*
* scaleSVG scales SVG to fit the parent DIV.
* */
export function scaleSVG(svg: string): string {
    const parser = new DOMParser();
    let doc = parser.parseFromString(svg, "image/svg+xml")!.querySelector("svg");
    //@ts-ignore
    doc.style.preserveAspectRatio="xMaxYMax";
    //@ts-ignore
    doc.style.width="100%";
    //@ts-ignore
    doc.style.height="100%";
    //@ts-ignore
    return new XMLSerializer().serializeToString(doc);
}
