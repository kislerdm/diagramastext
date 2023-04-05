# Changelog

## [0.0.5] - 2023-04-05

### Added

- User input validation according to the user status, i.e. registered/non-registered split
- Improved user feedback in case the model fails to predict a graph
- Trigger diagram generation process on the ctrl+enter keys press
- Cancellation of requests over 10 sec. on the escape key press
- Hide popup on the escape key press

### Changed

- Refactored the webclient codebase:
  - typescript
  - vite
  - tests coverage improvement

## [0.0.4] - 2023-03-23

### Added

- Support for the [`User`](https://github.com/plantuml-stdlib/C4-PlantUML/#supported-diagram-types) C4 macro.
- Diagram legend by default. It can be removed upon an explicit prompt's request only. 
**Examples**: 

The prompt yields the diagram with the legend: 
```
diagram of three boxes 
```

The prompts yield the diagram without the legend:

```
diagram of three boxes without legend
```

```
diagram of three boxes, no legend
```

```
diagram of three boxes with no legend
```

```
diagram of three boxes. remove legend
```

### Fixed

- ([#56](https://github.com/kislerdm/diagramastext/issues/56)) Moved from the model
  from "[code-davinci-002](https://platform.openai.com/docs/models/codex)"
  to "[gpt-3.5-turbo](https://platform.openai.com/docs/models/gpt-3-5)" following on the OpenAI announcement to
  discontinue support for the Codex API.

## [0.0.3] - 2023-03-21

### Fixed

- Migrated the core logic to GCP to bypass the OpenAI latency exceeding the AWS GW request limit of 29 sec. _Note_ that
  the tf codebase for AWS infra is preserved for the release.

### Changed

- Refactored the [`core`](./server/core) logic:
    - [Port-adaptor](https://web.archive.org/web/20180822100852/http://alistair.cockburn.us/Hexagonal+architecture)
      approach
    - Split sub-packages to dedicated modules in the `pkg` directory. Note that the packages below released as `v0.0.1`:
        - HTTP client with the backoff-retry mechanism
        - GCP Secretsmanager
        - Postgres client
        - OpenAI client
- [tf](infrastructure/neon) state with the Neon db provisioning migrated to GCP.

## [0.0.2] - 2023-02-21

### Changed

- Model configurations were changed:

```json
{
  "model": "code-davinci-002",
  "max_tokens": 768,
  "temperature": 0.2,
  "best_of": 3,
  "top_p": 1,
  "frequency_penalty": 0,
  "presence_penalty": 0,
  "stop": [
    "\n"
  ]
}
```

- (minor) Added the logo

### Security

- Added the vault to manage access credentials to improve security

## [0.0.1] - 2023-02-13

Beta release.

### Added

- Core server-side logic to deliver [C4 model](https://c4model.com/) diagrams
    - OpenAI agent
    - PlantUML agent
    - AWS Lambda interface
    - HTTP interface
- Webclient:
    - Freetext intput with the input validation
    - Output
    - Go and download buttons

#### Open IA Model

Configuration:

```json
{
  "model": "code-cushman-001",
  "max_tokens": 768,
  "temperature": 0.2,
  "best_of": 3,
  "top_p": 1,
  "frequency_penalty": 0,
  "presence_penalty": 0,
  "stop": [
    "\n"
  ]
}
```
