/*
Align is a scheduling tool that allows users to schedule events with other users. It is designed to be modular, so that
users can easily receive schedule reminders and updates through different platforms. Align's configuration file has
settings described below:

```yaml
settings:

	title: "Group Meetup"        # Title of the event
	interval: 7                  # How many days to ask for availability
	offset: 2                    # How many days past the contact time to ask for availability
	timezone: "America/New_York" # Timezone that cron strings are based on
	contact_time: "0 10 * * 0"   # Contact time cron string (Sunday at 10:00 AM)
	deadline_time: "0 10 * * 1"  # Deadline time cron string (Monday at 10:00 AM)

persons:

  - name: "Person 1"   	         # Name of the person
    request_method: "discord"    # Method to request information from
    response_method: "discord"   # Method to respond with information
    id: "PERSONS_ID"             # Identifiying string for the person (Discord ID, Telegram ID, etc.)

  - name: "Person 2"
    ...

```

Currently, the `request_method` and `response_methods` must be the same value, but this will be changed in future updates.

Examples for each module can be found in the 'examples/' directory. These directories contain the most barebones setup
align needs to function. If you are using align in a more complicated package, you can provide the same types in the
examples to get align working.

## SQL

Align has an option to use SQL to store availability data. This is useful if align ever stops running (server resetting,
power outages, etc). If align is restarted without persisting data, the availability data may be lost, and the subsequent
schedule alignment may be incorrect (align tries to mitigate this fact as much as possible, but some necessary data cannot
be recovered in this case, such as discord message IDs).

You can provide SQL credentials to the align configuration file to use SQL. The yaml format is as follows:

```yaml
sql:

	user: SQL_USER
	passwd: SQL_PASSWD
	net: SQL_NET
	addr: SQL_ADDR
	dbname: SQL_DBNAME

```

## Discord

Discord is easy to set up with align. Simply providing a Discord session to align will allow it to send and receive
messages. Keep in mind that, in order for a Discord bot to send a message to a user, it must be in a mutual server
with said user. This is a limitation of the Discord API, and align cannot bypass this.

To collect Discord IDs, you can right click on a profile you want to contact and click 'Copy User ID.' You can provide
this information to align's configuration file.

## Telegram

To initialize telegram with align, you can start a telegram session using [telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api).
This package is used to interact with the Telegram Bot API. Once you have started this session, align can use it to send and receive messages
for easy and convienient scheduling.

However, Telegram is more difficult to set up and maintain with align. These constraints originate from the [Telegram Bot API](https://core.telegram.org/bots/api) itself. These reasons are:
* Telegram bots are not allowed to send messages to users who have not initiated some sort of conversation with the bot
* Telegram servers only store updates for 24 hours, so if the bot is down for more than 24 hours, it may not receive poll updates

So, to use Telegram with align, you must:
* Have users initiate a conversation with the bot using '/start', or clicking the bottom of the bar when messaging the bot. The bot does not have to be online, but it must be activated within 24 hours to receive the update
* Keep the bot online at least once every 24 hours so it can receive updates from telegram. If the bot is down for more than 24 hours, it may not receive poll updates and return incorrect schedule times

The best way to do this in practice is to approach the user you want to contact using Telegram and have them start a
conversation with the bot while the bot is online (or during the 24 hour update period). This way, the bot can send
messages to the user without any issues. Secondly, you need to receive this user's Telegram User ID (not username). This can
be done by having that user message '@userinfobot', clicking 'start', and recording the 'User Id Information' field.
*/
package align
