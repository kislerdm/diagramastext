# Changelog

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
  "stop": ["\n"]
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
  "stop": ["\n"]
}
```
