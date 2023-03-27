export default function Footer(version: string): string {
    const copyrightStr: string = `<p style="font-size:16px">diagramastext.dev &copy; ${new Date().getFullYear().toString()}`;
    const versionElement = version === "" ? "" : `<p style="font-size:6px;">version: ${version}</p>`;
    return `<footer style="padding:1rem">
    ${copyrightStr}
    ${versionElement}
</footer>
`
}
