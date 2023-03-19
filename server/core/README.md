# diagramastext `core` backend logic

[![codecov](https://codecov.io/github/kislerdm/diagramastext/branch/master/graph/badge.svg)](https://codecov.io/github/kislerdm/diagramastext)

The codebase orchestrates transformations of the user's inquiry.

## Local Development

### Requirements

- go 1.19
- gnuMake
- docker

### Commands

Run to see available commands:

```commandline
make help
```

Run to perform unittests of all modules:

```commandline
make tests
```

Run to build the docker image with the http server app:

```commandline
make docker-build
```

## References

- [zopfi](https://github.com/google/zopfli): The library used to compress and encode the C4 Diagram definition as code
  as the string request content to generate diagram using [PlantUML](www.plantuml.com/plantuml/uml) server.
- The encoding [logic](rendering/plantuml/plantump-webclient-mimic/src/converter.js)

## Acknowledgements

- [William MacKay](https://github.com/foobaz) for the [go-zopfli](https://github.com/foobaz/go-zopfli) module
