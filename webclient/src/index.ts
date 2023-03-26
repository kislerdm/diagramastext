import './global.css';
import Main from "./main";
import {Config} from "./ports";

const cfg: Config = {
    version: "{{.Env.VERSION}}",
    urlAPI: "{{ .Env.ApiURL }}",
    promptMinLength: 3,
    promptMaxLengthUserBase: 100,
    promptMaxLengthUserRegistered: 300,
};

Main(document.querySelector<HTMLDivElement>("main")!, cfg);
