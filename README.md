# Diagram As Text

The tool to generate diagram based on textual description.

🚀🚀🚀 _Mission_: to enable anyone to explain complex system, or process in a simple way.

🚀🚀 _Objective_: to streamline knowledge sharing through diagrams.

🚀 _MVP_: plain english to [C4 container](http://c4model.com/) diagrams.

The work is driven by the motivation to streamline knowledge sharing by enabling effective generation of visual
materials. We aim to take a step beyond the "diagram as code" approach and remove the additional step between the idea
of what shall be displayed and the illustration.

We all know that “a picture is worth a thousand words”. Although it takes quite some effort to make a diagram, LLM is
here for the rescue! All one needs is, a description in plain English to get desired result in no time! 🤖🦾

## Outlook

* [Contacts](#contacts)
* [Bets/Panning/Commitment](#bets-panning-commitment)
* [Contribution](#contribution)
    + [Ways of work](#ways-of-work)
        - [Manifesto](#manifesto)
        - [Process](#process)
        - [Tech stack](#tech-stack)
* [License](#license)
    + [Codebase](#codebase)
    + [Images and diagrams](#images-and-diagrams)

## Contacts

- <a href="mailto:hi@diagramastext.com">hi@diagramastext.com</a>
- [LinkedIn](https://www.linkedin.com/in/dkisler/)

## Bets/Panning/Commitment

- [Issues Board](https://github.com/users/kislerdm/projects/5/views/)

## Contribution

🔔 **Wanted**: founding contributors 🔔

### Ways of work

#### Manifesto

- We are driven by the mission
- We respect one another and the community
- We deliver in lean iterations
- We work async with pairing programming sessions
- We share the work openly, see the [license details](#license)

#### Process

- We follow [TDD](https://www.guru99.com/test-driven-development.html)
- We follow [RDD](https://tom.preston-werner.com/2010/08/23/readme-driven-development.html)
- We maintain flat modules structure whether possible
- We aim for simplicity with the least external dependencies
- We follow the release [guideline](https://keepachangelog.com/en/1.0.0/) and [semantic versioning](https://semver.org/)
- We follow conventional _commits_ [guideline](https://www.conventionalcommits.org/en/v1.0.0/):
  - `feat`: for features
  - `fix`: for defect fix
  - `chore`: for infra, ci, or docs adjustments; or refactoring
- We follow conventional _comments_ [guideline](https://conventionalcomments.org/) for code reviews
- We follow the [monorepo](https://monorepo.tools/) approach

#### Tech stack

- Languages:
  - Go 1.19
  - Vanilla javascript
  - Python 3.9
- Markup:
  - Markdown
  - HTML5
  - CSS3
- CI:
  - GitHub Actions
- Infra:
  - AWS:
    - IAM
    - Lambda
    - API Gateway
  - GitHub Pages
  - [Neon](https://neon.tech/)
  - _Temp_: bare server hosted on [netcup](https://www.netcup.de/)
  - Cloudflare
  - namecheap
  - godaddy
- Tools:
  - gnuMake
  - Docker
  - terraform
  - slack
- Logic:
  - PlantUML
  - OpenAI

## License

### Codebase

The codebase is distributed under the [Apache 2.0 licence](LICENSE).

### Images and diagrams

<a rel="license" href="http://creativecommons.org/licenses/by-nc-sa/4.0/"><img alt="Creative Commons Licence" style="border-width:0" src="https://i.creativecommons.org/l/by-nc-sa/4.0/80x15.png" /></a><br />
This work
by <a xmlns:cc="http://creativecommons.org/ns#" href="diagramastext.dev" property="cc:attributionName" rel="cc:attributionURL">
diagramastext.dev</a> is licensed under a <a rel="license" href="http://creativecommons.org/licenses/by-nc-sa/4.0/">
Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International License</a>.
Based on a work at [https://github.com/kislerdm/diagramastext](https://github.com/kislerdm/diagramastext).
Permissions beyond the scope of this license may be available
at <a xmlns:cc="http://creativecommons.org/ns#" href="diagramastext.dev" rel="cc:morePermissions">diagramastext.dev</a>. 
