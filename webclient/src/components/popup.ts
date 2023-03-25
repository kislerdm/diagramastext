// @ts-ignore
import {modal, modalContent, msg} from "./popup.module.css";

export function ShowPopup(msg: HTMLElement) {
    document.querySelector<HTMLDivElement>("main")!.innerHTML += `
<div class="${modal}">
    <div class="${modalContent}">
        <span class="${close}">&times;</span>
        <div>${msg}</div>
    </div>
</div>
`
}
