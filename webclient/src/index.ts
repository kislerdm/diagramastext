import './global.css';
import Main from "./components/main";
import {Config} from "./ports";

const cfg: Config = {
    version: "{{.Env.VERSION}}",
    urlAPI: "{{ .Env.ApiURL }}",
    promptMinLength: 3,
    promptMaxLengthUserBase: 100,
    promptMaxLengthUserRegistered: 300,
};

document.querySelector<HTMLDivElement>("main")!.innerHTML = Main(cfg);
