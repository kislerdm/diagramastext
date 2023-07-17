# Changelog

## [0.0.7] - 2023-07-18

### Fixed

- Deserialization of the OpenAI response

## [0.0.6] - 2023-04-13

### Added

- Authentication and authorization layer to use the Rest API.
- `/quotas` endpoint to fetch the service usage

### Fixed

- CORS: validation of preflight calls.
- Response's content-type header value.

### Changed

- OpenAI's tokens consumption tracking.

## [0.0.6beta] - 2023-04-12

### Added

- Rest API interface to the core logic. Find the specification [here](https://diagramastext.dev/api-reference/).

### Fixed

- OpenAI's prediction parsing and deserialization when the graph's JSON is surrounded by text.

**_Example_**

_GIVEN_: The OpenAI chat response's content below  

> Here's the C4 diagram for a Python web server reading from an external Postgres database:
> 
> ```
> {
>   "title": "Python Web Server Reading from External Postgres Database",
>   "nodes": [
>     {"id": "0", "label": "Web Server", "technology": "Python"},
>     {"id": "1", "label": "Postgres", "technology": "Postgres", "external": true, "is_database": true}
>   ],
>   "links": [
>     {"from": "0", "to": "1", "label": "reads from Postgres", "technology": "TCP"}
>   ],
>   "footer": "C4 Model"
> }
> ```
> 
> The diagram shows two nodes: a Python web server and an external Postgres database. The web server reads data from the Postgres database over TCP.

_WHEN_: apply new deserialization and parsing logic

_THEN_: get the following graph definition:

```
{"title":"Python Web Server Reading from External Postgres Database","nodes":[{"id":"0","label":"Web Server","technology":"Python"},{"id":"1","label":"Postgres","technology":"Postgres","external":true,"is_database":true}],"links":[{"from":"0","to":"1","label":"reads from Postgres","technology":"TCP"}],"footer":"C4 Model"}
```

### Changed

- Migrated to [pgx](https://github.com/jackc/pgx) from [lib/pq](https://github.com/lib/pq). 

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
