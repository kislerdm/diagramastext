window.onload = function() {
  window.ui = SwaggerUIBundle({
    url: "./api-spec.yaml",
    domNode: document.getElementsByClassName("main")[0],
    deepLinking: true,
    docExpansion: "list",
    syntaxHighlight: {
      activate: true,
      theme: "tomorrow-night",
    },
    displayRequestDuration: true,
    defaultModelRendering: "example",
    supportedSubmitMethods: [],
    defaultModelsExpandDepth: -1,
    request: {
      curlOptions: ["--limit-rate 1"],
    },
  });
};
