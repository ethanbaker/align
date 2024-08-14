<!--
  Created by: Ethan Baker (contact@ethanbaker.dev)
  
  Adapted from:
    https://github.com/othneildrew/Best-README-Template/

Here are different preset "variables" that you can search and replace in this template.
`project_description`
`documentation_link`
-->

<div id="top"></div>


<!-- PROJECT SHIELDS/BUTTONS -->
[![GoDoc](https://godoc.org/github.com/ethanbaker/align?status.svg)](https://godoc.org/github.com/ethanbaker/align)
[![Go Report Card](https://goreportcard.com/badge/github.com/ethanbaker/align)](https://goreportcard.com/report/github.com/ethanbaker/align)

<!--NEED GITHUB WORKFLOW [![Go Coverage](https://github.com/ethanbaker/align/wiki/coverage.svg)](https://raw.githack.com/wiki/ethanbaker/align/coverage.html)-->
![1.0.0](https://img.shields.io/badge/status-1.0.0-red)
[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![License][license-shield]][license-url]
[![LinkedIn][linkedin-shield]][linkedin-url]

<!-- PROJECT LOGO -->
<br><br><br>
<div align="center">
  <a href="https://github.com/ethanbaker/align">
    <img src="./docs/logo.png" alt="Logo" width="80" height="80">
  </a>

  <h3 align="center">Align</h3>

  <p align="center">
    Easily schedule nights to meet with friends!
  </p>
</div>


<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About</a>
      <ul>
        <li><a href="#built-with">Built With</a></li>
      </ul>
    </li>
    <li><a href="#getting-started">Getting Started</a></li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#roadmap">Roadmap</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
    <li><a href="#contact">Contact</a></li>
    <li><a href="#acknowledgments">Acknowledgments</a></li>
  </ol>
</details>


<!-- ABOUT -->
## About

Align is a scheduling tool that allows users to schedule events with other users. It is
designed to be modular, so that users can easily receive schedule reminders and updates
through different platforms. Align's uses a configuration file combined with 
user-controlled sessions to seamlessly integrate with your own custom tools.

Currently, align allows you to contact users through Discord or Telegram. More outreach
methods are planned in the future!

Check out align's example usages [here](https://github.com/ethanbaker/align/tree/main/examples).

<p align="right">(<a href="#top">back to top</a>)</p>


### Built With

* [Golang](https://go.dev)
* [Cron](https://en.wikipedia.org/wiki/Cron)
* [Discord Go](https://github.com/bwmarrin/discordgo)
* [Telegram-Bot-API](https://github.com/go-telegram-bot-api/telegram-bot-api)

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- GETTING STARTED -->
## Getting Started

To get started with align, you can follow one of the ready-made examples [here](https://github.com/ethanbaker/align/tree/main/examples).

The general gist of align is as follows:
* You have a project that utilizes a session (Discord, Telegram, etc)
* You have an align configuration file
* You initialize an align manager using your provided config file that attaches itself to the running session

In doing this, you can attach align ontop of other programs, such as a ready-made Discord/Telegram bot.

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- USAGE EXAMPLES -->
## Usage

Examples can be found [here](https://github.com/ethanbaker/align/tree/main/examples).
Existing examples include:
* [Discord Implementation](https://github.com/ethanbaker/align/tree/main/examples/discord). 
* [Telegram Implementation](https://github.com/ethanbaker/align/tree/main/examples/telegram).
* [Multi-Module Implementation](https://github.com/ethanbaker/align/tree/main/examples/all).

These examples show how align can be attached to already-running sessions with an example
configuration file.

_For more details, please refer to the [documentation][documentation-url]._

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- ROADMAP -->
## Roadmap

- [ ] Allow for different request/response methods
- [ ] Add SMS outreach method

See the [open issues][issues-url] for a full list of proposed features (and known issues).

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- CONTRIBUTING -->
## Contributing

For issues and suggestions, please include as much useful information as possible.
Review the [documentation][documentation-url] and make sure the issue is actually
present or the suggestion is not included. Please share issues/suggestions on the
[issue tracker][issues-url].

For patches and feature additions, please submit them as [pull requests][pulls-url]. 
Please adhere to the [conventional commits][conventional-commits-url]. standard for
commit messaging. In addition, please try to name your git branch according to your
new patch. [These standards][conventional-branches-url] are a great guide you can follow.

You can follow these steps below to create a pull request:

1. Fork the Project
2. Create your Feature Branch (`git checkout -b branch_name`)
3. Commit your Changes (`git commit -m "commit_message"`)
4. Push to the Branch (`git push origin branch_name`)
5. Open a Pull Request

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- LICENSE -->
## License

This project uses the Apache 2.0 License. You can find more information in the [LICENSE][license-url] file.

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- CONTACT -->
## Contact

Ethan Baker - contact@ethanbaker.dev - [LinkedIn][linkedin-url]

Project Link: [https://github.com/ethanbaker/align][project-url]

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- ACKNOWLEDGMENTS -->
## Acknowledgments

* All the friendgroups out there who struggle to connect regularly!

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->
[contributors-shield]: https://img.shields.io/github/contributors/ethanbaker/align.svg
[forks-shield]: https://img.shields.io/github/forks/ethanbaker/align.svg
[stars-shield]: https://img.shields.io/github/stars/ethanbaker/align.svg
[issues-shield]: https://img.shields.io/github/issues/ethanbaker/align.svg
[license-shield]: https://img.shields.io/github/license/ethanbaker/align.svg
[linkedin-shield]: https://img.shields.io/badge/-LinkedIn-black.svg?logo=linkedin&colorB=555

[contributors-url]: <https://github.com/ethanbaker/align/graphs/contributors>
[forks-url]: <https://github.com/ethanbaker/align/network/members>
[stars-url]: <https://github.com/ethanbaker/align/stargazers>
[issues-url]: <https://github.com/ethanbaker/align/issues>
[pulls-url]: <https://github.com/ethanbaker/align/pulls>
[license-url]: <https://github.com/ethanbaker/align/blob/master/LICENSE>
[linkedin-url]: <https://linkedin.com/in/ethandbaker>
[project-url]: <https://github.com/ethanbaker/align>

[product-screenshot]: path_to_demo
[documentation-url]: <documentation_link>

[conventional-commits-url]: <https://www.conventionalcommits.org/en/v1.0.0/#summary>
[conventional-branches-url]: <https://docs.microsoft.com/en-us/azure/devops/repos/git/git-branching-guidance?view=azure-devops>