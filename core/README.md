# diagramastext backend logic

[![codecov](https://codecov.io/github/kislerdm/diagramastext/branch/master/graph/badge.svg)](https://codecov.io/github/kislerdm/diagramastext)

The codebase orchestrates transformations of the user's inquiry.

## Design

### C4 Containers

TBD

## References

- [zopfi](https://github.com/google/zopfli): The library used to compress and encode the C4 Diagram definition as code
  as the string request content to generate diagram using [PlantUML](www.plantuml.com/plantuml/uml) server.
- The encoding [logic](../web/plantuml-mimic/src/converter.js)

## Acknowledgements

- [William MacKay](https://github.com/foobaz) for the [go-zopfli](https://github.com/foobaz/go-zopfli) module
