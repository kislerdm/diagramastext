// @ts-ignore
import {arrow, box, boxText, punch} from './main.module.css';
import Footer from "./footer";
import Header from "./header";
import {Config} from "@/ports";

// @ts-ignore
import placeholderDiagram from "./svg/output-placeholder.svg?raw";
import logoGithub from "./svg/github.svg";
import logoSlack from "./svg/slack.svg";
import logoLinkedin from "./svg/linkedin.svg";
import logoEmail from "./svg/email.svg";

type FsmHtmlID = {
    Input: string
    Output: string
    Trigger: string
    Download: string
}

class User {}

class FSM {
    private _config: Config
    private _svg: string
    private _prompt: string
    private readonly _placeholderPrompt: string
    private readonly _placeholderSVG: string
    private _user: User

    constructor(cfg: Config, ids: FsmHtmlID) {
        FSM.validateConfig(cfg)
        this._config = cfg;

        this._placeholderPrompt = "C4 diagram of a Go web server reading from external Postgres database over TCP";
        this._placeholderSVG = placeholderDiagram;
        this._prompt = "";
        this._svg = "";
        this._user = new User();
    }

    static validateConfig(cfg: Config) {
        if (cfg.urlAPI === undefined || cfg.urlAPI === null || cfg.urlAPI === "") {
            throw new TypeError("config must contain urlAPI attribute");
        }
    }

    get placeholderInputPrompt(): string {
        return this._placeholderPrompt;
    }

    get placeholderOutputSVG(): string {
        return this._placeholderSVG;
    }

    // #validatePrompt(prompt: string) {
    //     switch (this._user.is_registered()) {
    //         case true:
    //             validatePromptLength(prompt, promptLengthMin, promptLengthMaxRegisteredUser);
    //             break;
    //         default:
    //             validatePromptLength(prompt, promptLengthMin, promptLengthMaxBaseUser);
    //             break;
    //     }
    // }
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

export default function Main(cfg: Config): string {
    const id: FsmHtmlID = {
        Input: "0",
        Trigger: "1",
        Output: "2",
        Download: "3",
    };

    const fms = new FSM(cfg, id);

    return `${Header}

<div style="font-size:30px;margin: 20px 0 10px">
    Generate <span style="font-weight:bold">diagrams</span> using 
    <span style="font-style:italic;font-weight:bold">plain English</span> in no time!
</div>

${Input(id.Input, id.Trigger, cfg.promptMinLength, cfg.promptMaxLengthUserRegistered, fms.placeholderInputPrompt)}

<i class="${arrow}"></i>

${Output(id.Output, id.Download, fms.placeholderOutputSVG)}

${Disclaimer}

${Footer(cfg.version)}
`;
}
