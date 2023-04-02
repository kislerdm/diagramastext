// @ts-ignore
import logoGithub from "./svg/github.svg";
// @ts-ignore
import logoSlack from "./svg/slack.svg";
// @ts-ignore
import logoLinkedin from "./svg/linkedin.svg";
// @ts-ignore
import logoEmail from "./svg/email.svg";


export default function Footer(version: string = ""): string {
    const copyrightStr =
        `<p style="font-size:16px">diagramastext.dev &copy; ${new Date().getFullYear().toString()}</p>`;

    const versionElement = version === "" ? "" : `<p style="font-size:6px;">version: ${version}</p>`;

    const iconSquareSize = 20;

    const socialContact = `<p id="contacts" style="margin-top:-10px">
    <a href="https://github.com/kislerdm/diagramastext">
        <img src="${logoGithub}" width=${iconSquareSize} height=${iconSquareSize} alt="github"/>
    </a>
    <a href="https://join.slack.com/t/diagramastextdev/shared_invite/zt-1onedpbsz-ECNIfwjIj02xzBjWNGOllg">
        <img src="${logoSlack}" width=${iconSquareSize} height=${iconSquareSize} alt="slack"/>
    </a>
    <a href="https://www.linkedin.com/in/dkisler">
        <img src="${logoLinkedin}" width=${iconSquareSize} height=${iconSquareSize} alt="linkedin"/>
    </a>
    <a href="mailto:contact@diagramastext.dev">
        <img src="${logoEmail}" width=${iconSquareSize} height=${iconSquareSize} alt="email"/>
    </a>
</p>`

    return `<footer style="padding:1rem">
    ${copyrightStr}
    ${socialContact}
    ${versionElement}
</footer>
`
}
