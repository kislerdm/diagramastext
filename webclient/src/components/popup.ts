// @ts-ignore
import {close, loader, modal, modalContent} from "./popup.module.css";

export class Popup {
    static mount(): string {
        return `<div id="popup" class="${modal}">
    <div class="${modalContent}">
        <span class="${close}">&times;</span>
        <div id="modalMsg"></div>
    </div>
</div>`;
    }

    static show(parent: HTMLDivElement, msg: string) {
        const popup = [...parent.getElementsByClassName(modal)].find(el => el.id === "popup")!;
        const msgEl = [...popup.getElementsByTagName("div")].find(el => el.id === "modalMsg")!;
        const popupClose = popup.getElementsByClassName(close)[0]!;
        //@ts-ignore
        msgEl.innerHTML = msg;

        //@ts-ignore
        popup.style.display = "block";

        function hide() {
            //@ts-ignore
            msgEl.innerHTML = "";
            //@ts-ignore
            popup.style.display = "none";
        }

        popupClose.addEventListener("click", () => hide())

        window.onclick = (event) => {
            if (event.target === popup) {
                hide()
            }
        }

        parent.addEventListener("keydown", (event) => {
            if ((event.key === "Escape" || event.key === "Esc")) {
                hide();
            }
        })
    }

    static error(parent: HTMLDivElement, msg: string) {
        const content = parent.getElementsByClassName(modalContent)[0]!;
        //@ts-ignore
        content.style.boxShadow = "0 0 3px 3px #e00d0d";
        Popup.show(parent,`<p style="font-size:medium;font-weight:bold"><span style="color:red">Error! </span>${msg}</p>`);
    }
}

export class Loader {
    static mount(): string {
        return `<div id="loader" class="${modal}">
    <div class="${modalContent}" style="width:150px;margin-top:200px;border:none;box-shadow:none;background:none">
        <div class="${loader}"></div>
    </div>
</div>`;
    }

    private static setStyle(parent: HTMLDivElement, style: Object) {
        // @ts-ignore
        const s = [...parent.getElementsByClassName(modal)].find(el => el.id === "loader")!.style;
        for (const [key, value] of Object.entries(style)) {
            s[key] = value;
        }
    }

    static show(parent: HTMLDivElement) {
         Loader.setStyle(parent,{display: "block"});
    }

    static hide(parent: HTMLDivElement) {
        Loader.setStyle(parent,{display: "none"});
    }
}
