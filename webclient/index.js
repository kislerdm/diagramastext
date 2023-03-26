function generateFeedbackLink(prompt) {
    let url = "https://github.com/kislerdm/diagramastext/issues/new";
    const params = {
        assignee: "kislerdm",
        labels: ["feedback", "defect"],
        title: `Webclient issue`,
        body: `## Environment
- App version: ${config.version}

## Prompt

\`\`\`
${prompt}
\`\`\`

## Details

- Please describe your chain of actions, i.e. what preceded the state you report?
- Please attach screenshots whether possible

## Expected behaviour

Please describe what should have happened following the actions you described.
`,
    };

    const query = Object.keys(params).map(key => key + '=' + encodeURIComponent(params[key])).join('&');

    return `${url}?${query}`;
}

function showErrorPopupWithUserFeedback(msg, prompt) {
    const feedbackLink = generateFeedbackLink(prompt);

    msg += `<br>Please retry, and` +
        `<br>Leave the ` +
        `<a id="feedback-link" href=${feedbackLink} target="_blank" rel="noopener">feedback</a>`

    showErrorPopup(msg)
}
