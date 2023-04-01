// @ts-ignore
import {modal, modalContent, close, loader} from "./popup.module.css";

export class Popup {
    mount(): string {
        return `
<div class="${modal}">
    <div class="${modalContent}">
        <span class="${close}">&times;</span>
        <div id="modalMsg"></div>
    </div>
</div>`;
    }

    show(msg: string) {
        const msgDisplay = document.getElementById("modalMsg");

        //@ts-ignore
        msgDisplay.innerHTML = msg;

        const m = document.getElementsByClassName(modal)[0];
        //@ts-ignore
        m.style.display = "block";

        document.getElementsByClassName(close)[0]!.addEventListener("click", () => {
            //@ts-ignore
            msgDisplay.innerHTML = "";
            //@ts-ignore
            m.style.display = "none";
        })

        window.onclick = (event) => {
            if (event.target === m) {
                //@ts-ignore
                msgDisplay.innerHTML = "";
                //@ts-ignore
                m.style.display = "none";
            }
        }
    }

    error(msg: string) {
        //@ts-ignore
        document.getElementsByClassName(modalContent)[0]!.style.boxShadow = "0 0 3px 3px #e00d0d";
        this.show(`<p style="font-size:medium;font-weight:bold"><span style="color:red">Error! </span>${msg}</p>`);
    }
}

export class Loader {
    mount(): string {
        return `<div id="loader" class="${modal}">
<div class="${modalContent}" style="width:150px;margin-top:200px;border:none;box-shadow:none;background:none">
<div class="${loader}"></div>
</div></div>`;
    }

    show() {
        //@ts-ignore
        document.getElementById("loader").style.display = "block";
    }

    hide() {
        //@ts-ignore
        document.getElementById("loader").style.display = "none";
    }
}
