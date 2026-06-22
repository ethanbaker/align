/*
Package align is a wire-in library that calculates time windows when a group
of people are mutually available. It is designed to be imported by a consumer
binary that provides the platform-specific communication layer.

# Architecture

The module is split into three layers:

  - Core package (this package): domain types, date-intersection logic, and the
    Contactor and Persister interfaces.
  - Adapter packages (discord, telegram): thin wrappers around platform SDKs
    that implement Contactor.
  - Consumer binary (example): reads config, constructs adapters and a
    Scheduler, then starts the scheduling loop.

The dependency direction is strict and one-way: adapter packages import the
core package; the core package knows nothing about the adapters.

# Configuration

Provide a YAML configuration file with the following structure:

	settings:
	  title: "Group Meetup"        # label shown in poll messages
	  interval: 7                  # number of days to poll per cycle
	  offset: 2                    # days after contact_time to start the window
	  timezone: "America/New_York" # IANA timezone for cron strings
	  contact_time: "0 10 * * 0"   # cron: when to send availability polls
	  deadline_time: "0 10 * * 1"  # cron: when to gather responses and notify

	persons:
	  - name: "Alice"
	    request_method: "discord"    # platform used to ask for availability
	    response_method: "discord"   # platform used to send the result
	    id: "ALICE_DISCORD_USER_ID"

	  - name: "Bob"
	    request_method: "telegram"
	    response_method: "telegram"
	    id: "BOB_TELEGRAM_USER_ID"

request_method and response_method must match a key in the map of Contactors
passed to NewScheduler (e.g. "discord" or "telegram"). Different platforms may
be used for request and response.

# Contactor interface

Platform adapters implement Contactor:

  - Request: send an availability poll to a person.
  - Gather: collect their responses (returns an AvailabilityMap).
  - Notify: deliver the final scheduling result.

Adapters are thin translators; all business logic lives in the core package.

# Persister interface

Persisters allow an in-progress session to survive a process restart:

  - Save: write Session state to durable storage.
  - Load: restore a Session by name. Return (nil, nil) if none exists.

NopPersister is provided for in-memory-only use. Implement Persister backed by
SQL, a file, or any other store for production durability.

# Quick start

	// Create platform adapters.
	discordAdapter := discord.New(discordSession)
	telegramAdapter := telegram.New(telegramBot)

	// Create a scheduler.
	scheduler, err := align.NewScheduler(
	    "my-group",
	    "./config.yml",
	    map[string]align.Contactor{
	        "discord":  discordAdapter,
	        "telegram": telegramAdapter,
	    },
	    align.NopPersister{},
	)

	// Start the cron-driven loop (non-blocking).
	scheduler.Start()

See the example/ directory for a complete runnable program.

# Discord notes

The Discord bot must share a server with each user it messages — this is a
Discord API constraint. Collect user IDs by right-clicking a profile and
choosing "Copy User ID" (Developer Mode must be enabled).

# Telegram notes

Telegram bots cannot initiate conversations. Each person must send the bot at
least one message (e.g. /start) before the bot can message them. Additionally,
Telegram stores updates for only 24 hours, so the bot must receive updates at
least once per day to avoid missing poll responses.
*/
package align
