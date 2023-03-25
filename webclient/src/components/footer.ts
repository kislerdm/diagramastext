const copyrightStr: string = `<p style="font-size: 16px;">diagramastext.dev &copy; ${new Date().getFullYear().toString()}`;

export default function Footer(version: string) {
    return `
<footer style="padding: 1rem">
    ${copyrightStr}    
    <p style="font-size: 8px;">version: ${version}</p>
</footer>
`
}
