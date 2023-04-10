import './global.css';
import Main from "./main";
import {Config} from "./ports";

const cfg: Config = {
    // @ts-ignore
    version: import.meta.env.VITE_VERSION,
    // @ts-ignore
    urlAPI: import.meta.env.VITE_URL_API,
    promptMinLength: 3,
    promptMaxLengthUserBase: 100,
    promptMaxLengthUserRegistered: 300,
}

const mountPoint = document.querySelector<HTMLDivElement>("main")!;

switch (window.location.pathname) {
    case "/api-reference":
        location.replace( `${window.location.pathname}/index.html`);
        break;
    default:
        Main(mountPoint, cfg);
        window.history.replaceState({}, document.title, "./");
        break;
}
