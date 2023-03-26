// @ts-ignore
import {arrow, box, boxText, punch} from './main.module.css';
import Footer from "./components/footer";
import Header from "./components/header";
import {Config, DataSVG} from "@/ports";

// @ts-ignore
import placeholderDiagram from "./components/svg/output-placeholder.svg?raw";
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

type FsmHtmlID = {
    Input: string
    Output: string
    Trigger: string
    Download: string
}

let svg: string = "";

class FSM {
    private _config: Config
    private _user: User
    private readonly _popup: Popup
    private readonly _loader: Loader
    private _fetchErrorCnt: number = 0
    private readonly _fetchErrorCntMax: number = 2
    ids: FsmHtmlID

    constructor(cfg: Config, ids: FsmHtmlID, popup: Popup, loaderSpinner: Loader) {
        FSM.validateConfig(cfg)
        this._config = cfg;
        this.ids = ids

        this._popup = popup;
        this._loader = loaderSpinner;

        this._user = new User();

        document.getElementById(this.ids.Trigger)!.addEventListener("click", () =>
            //@ts-ignore
            this.generateDiagram(document.getElementById(this.ids.Input)!.value)
        )

        //@ts-ignore
        document.getElementById(this.ids.Download)!.addEventListener("click", this.download)
    }

    static validateConfig(cfg: Config) {
        if (cfg.urlAPI === undefined || cfg.urlAPI === null || cfg.urlAPI === "") {
            throw new TypeError("config must contain urlAPI attribute");
        }
    }

    static placeholderInputPrompt(): string {
        return "C4 diagram of a Go web server reading from external Postgres database over TCP";
    }

    static placeholderOutputSVG(): string {
        return placeholderDiagram;
    }

    private validatePromptLength(prompt: string, lengthMin: number, lengthMax: number) {
        if (prompt.length < lengthMin || prompt.length > lengthMax) {
            throw new RangeError(`The prompt must be between ${lengthMin} and ${lengthMax} characters long`)
        }
    }

    private validatePrompt(prompt: string) {
        switch (this._user.is_registered()) {
            case true:
                this.validatePromptLength(prompt, this._config.promptMinLength, this._config.promptMaxLengthUserRegistered);
                break;
            default:
                this.validatePromptLength(prompt, this._config.promptMinLength, this._config.promptMaxLengthUserBase);
                break;
        }
    }

    generateDiagram(prompt: string) {
        // @ts-ignore
        prompt = prompt.trim();
        if (FSM.placeholderInputPrompt() !== prompt) {
            console.log(prompt);
            try {
                this.validatePrompt(prompt);
                // @ts-ignore
            } catch (e: Error) {
                this._popup!.error(e.message);
                return;
            }

            this._loader.show();
            fetch(this._config.urlAPI, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({
                    "prompt": prompt,
                }),
            }).then((resp: Response) => {
                if (!resp.ok) {
                    this._fetchErrorCnt++;
                    if (this._fetchErrorCnt > this._fetchErrorCntMax) {
                        const link = generateFeedbackLink(prompt, this._config.version);
                        throw new Error(`The errors repreat, please <a href="${link}" target="_blank" rel="noopener" style="color:#3498db;font-weight:bold">report</a>`);
                    }
                }

                switch (resp.status) {
                    case 200:
                        this._fetchErrorCnt = 0;
                        resp.json()
                            .then((data: DataSVG) => {
                                if (data.svg === null) {
                                    throw new Error("empty response");
                                } else {
                                    this._loader.hide();
                                    svg = data.svg;

                                    //@ts-ignore
                                    document.getElementById(this.ids.Output).innerHTML = data.svg;

                                    //@ts-ignore
                                    document.getElementById(this.ids.Download).disabled = false;
                                }
                            })
                        break;
                    case 400:
                        throw new Error("Unexpected prompt length");
                    case 404:
                        throw new Error("Faulty path");
                    case 429:
                        throw new Error("The server is experiencing high load, please try later");
                    case 500:
                        throw new Error("Unexpected error, please try later");
                    default:
                        resp.text().then((msg) => {
                            throw new Error(msg);
                        })
                }
            })
                .catch((e) => {
                    this._loader.hide();
                    this._popup!.error(e.message);
                })
        }
    }

    download() {
        if (svg !== "") {
            const link = document.createElement("a");
            link.setAttribute("download", "diagram.svg");
            link.setAttribute("href", `data:image/svg+xml,${encodeURIComponent(svg)}`);
            link.click();
        }
    }
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

export default function Main(mountPoint: HTMLDivElement, cfg: Config) {
    const id: FsmHtmlID = {
        Input: "inpt",
        Trigger: "trigger",
        Output: "output",
        Download: "download",
    };

    mountPoint.innerHTML = `${Header}

<div style="font-size:30px;margin: 20px 0 10px">
    Generate <span style="font-weight:bold">diagrams</span> using 
    <span style="font-style:italic;font-weight:bold">plain English</span> in no time!
</div>

${Input(id.Input, id.Trigger, cfg.promptMinLength, cfg.promptMaxLengthUserRegistered, FSM.placeholderInputPrompt())}

<i class="${arrow}"></i>

${Output(id.Output, id.Download, FSM.placeholderOutputSVG())}

${Disclaimer}

${Footer(cfg.version)}
`;

    new FSM(cfg, id, new Popup(mountPoint), new Loader(mountPoint));
}
