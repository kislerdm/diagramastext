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
* [Bets/Panning/Commitment](#betspanningcommitment)
* [Contribution](#contribution)
    + [Ways of work](#ways-of-work)
        - [Manifesto](#manifesto)
        - [Process](#process)
        - [Tech stack](#tech-stack)
        - [Documentation](#documentation)
* [License](#license)
    + [Codebase](#codebase)
    + [Images and diagrams](#images-and-diagrams)

## Contacts

- Submit your [request](https://github.com/kislerdm/diagramastext/issues/new?assignee=kislerdm&labels=feedback&title=PLEASE%20DEFINE%20YOUR%20REQUEST&body=%23%23%20What%0APlease%20describe%20your%20proposal.%0A%0A%23%23%20Why%0APlease%20clarify%20the%20context.%0A%0A%23%23%20How%0A%0A(optional)%20Please%20sketch%20execution%20paths.)
- Join us on [Slack](https://join.slack.com/t/diagramastextdev/shared_invite/zt-1onedpbsz-ECNIfwjIj02xzBjWNGOllg)
- Write us: <a href="mailto:hi@diagramastext.com">hi@diagramastext.com</a>
- Get in touch on [LinkedIn](https://www.linkedin.com/in/dkisler/)

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
- We follow conventional _comments_ [guideline](https://conventionalcomments.org/) for code reviews
- We follow conventional _commits_ [guideline](https://www.conventionalcommits.org/en/v1.0.0/):
    - `feat`: for features
    - `fix`: for defect fix
    - `chore`: for infra, ci, or docs adjustments; or refactoring
- We follow the [monorepo](https://monorepo.tools/) approach
- We follow [trunk-based development](https://trunkbaseddevelopment.com/) model
- We follow the release [guideline](https://keepachangelog.com/en/1.0.0/) and [semantic versioning](https://semver.org/)

#### Tech stack

- Languages:
    - Go 1.19
    - JavaScript ES2021
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
    - Cloudflare
    - namecheap
    - godaddy
- Tools:
    - gnuMake
    - Docker
    - terraform
- Logic:
    - PlantUML
    - OpenAI

#### Documentation

Please find further documentation [here](./docs).

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
