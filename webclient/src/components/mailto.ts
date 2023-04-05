export default function MailToLinkStr(innerHTML: string = ""): string {
    return `<a href="mailto:contact@diagramastext.dev" target="_blank" rel="noopener">${innerHTML}</a>`;
}
