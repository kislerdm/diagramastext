import {User} from './user.js';
import {config} from "./config.js";

const promptLengthMin = 3,
    promptLengthMaxBaseUser = 100,
    promptLengthMaxRegisteredUser = 300;

function generateFeedbackLink(prompt) {
    let url = "https://github.com/kislerdm/diagramastext/issues/new";
    const params = {
        assignee: "kislerdm",
        labels: ["feedback", "defect"],
        title: `Webclient issue`,
        body: `## Environment
- App version: ${config.version}

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

    const query = Object.keys(params).map(key => key + '=' + encodeURIComponent(params[key])).join('&');

    return `${url}?${query}`;
}

function showErrorPopupWithUserFeedback(msg, prompt) {
    const feedbackLink = generateFeedbackLink(prompt);

    msg += `<br>Please retry, and` +
        `<br>Leave the ` +
        `<a id="feedback-link" href=${feedbackLink} target="_blank" rel="noopener">feedback</a>`

    showErrorPopup(msg)
}

function showErrorPopup(msg) {
    const modal = document.getElementById("error-msg");
    const span = document.getElementsByClassName("close")[0];

    document.getElementById("error-msg-content").innerHTML =
        `<p style="font-size: medium;font-weight: bold;"><span style="color: red;">Error! </span>${msg}</p>`;

    modal.style.display = "block";

    span.onclick = function () {
        modal.style.display = "none";
    }

    window.onclick = function (event) {
        if (event.target === modal) {
            modal.style.display = "none";
        }
    }
}

function loaderShow() {
    document.getElementById("loader").style.display = "block";
}

function loaderHide() {
    document.getElementById("loader").style.display = "none";
}

function activateDownloadButton() {
    document.getElementById("download").disabled = false;
}

function validatePromptLength(prompt, lengthMin, lengthMax) {
    if (prompt.length < lengthMin || prompt.length > lengthMax) {
        throw new RangeError(`The prompt must be between ${lengthMin} and ${lengthMax} characters long`)
    }
}

class Flow {
    constructor(config) {
        Flow.validateConfig(config)

        this._config = config;
        this.user = new User();
        this._svg = "";
        this._prompt_placeholder = "C4 diagram of a Go web server reading from external Postgres database over TCP";

        this._diagram = document.getElementById("diagram");
    }

    static validateConfig(config) {
        if (config.url_api === undefined || config.url_api === null || config.url_api === "") {
            throw new TypeError("config must contain url_api attribute");
        }
    }

    #validatePrompt(prompt) {
        switch (this.user.is_registered()) {
            case true:
                validatePromptLength(prompt, promptLengthMin, promptLengthMaxRegisteredUser);
                break;
            default:
                validatePromptLength(prompt, promptLengthMin, promptLengthMaxBaseUser);
                break;
        }
    }

    fetchDiagram() {
        fetch(this._config.url_api, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                "prompt": prompt,
            }),
        }).then((resp) => {
            switch (resp.status) {
                case 200:
                    resp.json()
                        .then((data) => {
                            if (data.svg === null) {
                                throw new Error("empty response");
                            } else {
                                this._svg = data.svg;
                            }
                        })
                    break;
                case 400:
                    showErrorPopup("Unexpected prompt length");
                    break;
                case 429:
                    showErrorPopup("The server is experiencing high load, please try later");
                    break;
                default:
                    resp.text().then((msg) => {
                        throw new Error(msg);
                    })
            }
        })
    }

    trigger() {
        const prompt = document.getElementById("prompt").value.trim();

        if (this._prompt_placeholder !== prompt) {
            try {
                this.#validatePrompt(prompt);
            } catch (e) {
                showErrorPopup(e.message);
                return;
            }

            try {
                loaderShow();
                this.fetchDiagram();
            } catch (e) {
                loaderHide();
                showErrorPopupWithUserFeedback(e.message, prompt);
                return;
            }

            loaderHide();
            this.#setDiagramImage();
            activateDownloadButton();
        }
    }

    download() {
        if (this._svg !== "") {
            const link = document.createElement("a");
            link.setAttribute("download", "diagram.svg");
            link.setAttribute("href", `data:image/svg+xml,${encodeURIComponent(this._svg)}`);
            link.click();
        }
    }

    #setDiagramImage() {
        this._diagram.innerHTML = this._svg;
    }
}

const flow = new Flow(config);
export default flow;
