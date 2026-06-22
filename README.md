<!--
  Created by: Ethan Baker (contact@ethanbaker.dev)
  
  Adapted from:
    https://github.com/othneildrew/Best-README-Template/
-->

<div id="top"></div>

<!-- PROJECT SHIELDS/BUTTONS -->
[![GoDoc](https://godoc.org/github.com/ethanbaker/align?status.svg)](https://godoc.org/github.com/ethanbaker/align)
[![Go Report Card](https://goreportcard.com/badge/github.com/ethanbaker/align)](https://goreportcard.com/report/github.com/ethanbaker/align)
![status](https://img.shields.io/badge/status-2.0.0-blue)
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
    Easily find times when a group of friends are all free!
  </p>
</div>


<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li><a href="#about-the-project">About</a></li>
    <li><a href="#architecture">Architecture</a></li>
    <li><a href="#getting-started">Getting Started</a></li>
    <li><a href="#configuration">Configuration</a></li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#roadmap">Roadmap</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
    <li><a href="#contact">Contact</a></li>
  </ol>
</details>


<!-- ABOUT -->
## About

Align is a wire-in scheduling library that finds time windows when a group of
people are mutually available. You supply the platform sessions (Discord,
Telegram, or anything else you implement); align handles the polling, response
collection, date intersection, and result delivery.

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- ARCHITECTURE -->
## Architecture

```
github.com/ethanbaker/align/         ← core package
github.com/ethanbaker/align/discord/ ← Discord adapter
github.com/ethanbaker/align/telegram/← Telegram adapter
example/                             ← runnable consumer binary
```

**Core package** (`package align`) — domain types (`Person`, `Day`,
`AvailabilityMap`, `Session`), date-intersection logic, and two interfaces:

- **`Contactor`** — what a platform adapter must implement (`Request`,
  `Gather`, `Notify`).
- **`Persister`** — what a storage backend must implement (`Save`, `Load`).
  `NopPersister` is provided for in-memory use.

**Adapter packages** — each implements `Contactor` for one platform.
No business logic lives here; adapters only translate between the platform's
event model and the core package's method signatures.

**Consumer binary** — the only place where concrete adapter types and the core
`Scheduler` are wired together. See [`example/`](./example/) for a complete
working program.

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- GETTING STARTED -->
## Getting Started

```bash
go get github.com/ethanbaker/align
go get github.com/ethanbaker/align/discord   # if using Discord
go get github.com/ethanbaker/align/telegram  # if using Telegram
```

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- CONFIGURATION -->
## Configuration

Create a YAML file — e.g. `config.yml`:

```yaml
settings:
  title: "Group Meetup"
  interval: 7                   # days to poll each cycle
  offset: 2                     # days after contact_time to start the window
  timezone: "America/New_York"  # IANA timezone for the cron strings below
  contact_time: "0 10 * * 0"   # Sunday 10 AM — send availability polls
  deadline_time: "0 10 * * 1"  # Monday 10 AM — gather responses, send result

persons:
  - name: "Alice"
    request_method: "discord"   # platform to send the poll on
    response_method: "discord"  # platform to send the result on
    id: "ALICE_DISCORD_USER_ID"

  - name: "Bob"
    request_method: "telegram"
    response_method: "telegram"
    id: "BOB_TELEGRAM_USER_ID"
```

`request_method` and `response_method` must match a key in the `Contactor`
map you pass to `NewScheduler`. They may be different (e.g. ask via Discord,
respond via Telegram).

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- USAGE -->
## Usage

```go
package main

import (
    "github.com/ethanbaker/align"
    "github.com/ethanbaker/align/discord"
    "github.com/ethanbaker/align/telegram"
)

func main() {
    discordAdapter := discord.New(discordSession)
    telegramAdapter := telegram.New(telegramBot)

    scheduler, err := align.NewScheduler(
        "my-group",
        "./config.yml",
        map[string]align.Contactor{
            "discord":  discordAdapter,
            "telegram": telegramAdapter,
        },
        align.NopPersister{},                 // swap for a SQL persister in production
        align.SchedulerOptions{
            OnCompletion: func(days []align.Day) {
                // optional callback after alignment is computed
            },
        },
    )

    scheduler.Start() // cron-driven; non-blocking
    // ... signal handling
}
```

A complete runnable example is in [`example/`](./example/).

### Platform notes

**Discord** — the bot must share a server with every user it messages (Discord
API limitation). Enable Developer Mode, then right-click a user profile and
choose "Copy User ID" to get their ID.

**Telegram** — bots cannot initiate conversations; each person must first send
the bot a message (e.g. `/start`). Telegram retains updates for only 24 hours,
so the bot must poll at least daily to capture all poll responses.

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- ROADMAP -->
## Roadmap

- [ ] SQL-backed `Persister` implementation
- [ ] Allow different request and response methods per person
- [ ] SMS outreach adapter

See the [open issues][issues-url] for a full list of proposed features.

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- CONTRIBUTING -->
## Contributing

For issues and suggestions, please include as much useful information as
possible and share them on the [issue tracker][issues-url].

For patches and feature additions, please submit them as [pull requests][pulls-url]
following [conventional commits][conventional-commits-url].

1. Fork the Project
2. Create your Feature Branch (`git checkout -b branch_name`)
3. Commit your Changes (`git commit -m "commit_message"`)
4. Push to the Branch (`git push origin branch_name`)
5. Open a Pull Request

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- LICENSE -->
## License

Apache 2.0 — see the [LICENSE][license-url] file.

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- CONTACT -->
## Contact

Ethan Baker - contact@ethanbaker.dev - [LinkedIn][linkedin-url]

Project Link: [https://github.com/ethanbaker/align][project-url]

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- MARKDOWN LINKS & IMAGES -->
[contributors-shield]: https://img.shields.io/github/contributors/ethanbaker/align.svg
[forks-shield]: https://img.shields.io/github/forks/ethanbaker/align.svg
[stars-shield]: https://img.shields.io/github/stars/ethanbaker/align.svg
[issues-shield]: https://img.shields.io/github/issues/ethanbaker/align.svg
[license-shield]: https://img.shields.io/github/license/ethanbaker/align.svg
[linkedin-shield]: https://img.shields.io/badge/-LinkedIn-black.svg?logo=linkedin&colorB=555

[contributors-url]: https://github.com/ethanbaker/align/graphs/contributors
[forks-url]: https://github.com/ethanbaker/align/network/members
[stars-url]: https://github.com/ethanbaker/align/stargazers
[issues-url]: https://github.com/ethanbaker/align/issues
[pulls-url]: https://github.com/ethanbaker/align/pulls
[license-url]: https://github.com/ethanbaker/align/blob/master/LICENSE
[linkedin-url]: https://linkedin.com/in/ethandbaker
[project-url]: https://github.com/ethanbaker/align

[conventional-commits-url]: https://www.conventionalcommits.org/en/v1.0.0/#summary
[conventional-branches-url]: https://docs.microsoft.com/en-us/azure/devops/repos/git/git-branching-guidance?view=azure-devops
