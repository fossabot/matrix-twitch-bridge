# Matrix <--> Twitch Bridge

This is an Twitch bridge for Matrix using the Application Services (AS) API.

This bridge will pass all TwitchChannel messages through to Matrix,
and all Matrix messages through to the Twitch Channel.

## Setup

### 1. Installation

To install all dependencies and add a binary `matrix-twitch-bridge`:

```bash
go get -u -v github.com/mattn/go-sqlite3  # Required for the Database
go get -u -v github.com/Nordgedanken/matrix-twitch-bridge
```

#### Requirements

- [Golang](https://golang.org/)
- A Matrix homeserver you control running Synapse v0.18.5-rc3 or above.
- A registered [Twitch App](https://dev.twitch.tv/dashboard)

### 2. Configuration and Registration

The bridge must be configured before it can be run.
This tells the bridge where to find the homeserver
and how to bridge Twitch channels/users.

- Run `matrix-twitch-bridge` and follow the interactive Terminal Guide

### 3. Running

Finally, the bridge can be run using the following command:

```bash
matrix-twitch-bridge --client_id=<client_id from your Twitch App>
```

If you have changed the default config location add the `--config`
(or `-c`) flag to the above command.

If you want to change the DB location add the `--database`
(or `-db`) flag to the above command.

## What does it do

On startup, it will listen for incoming Twitch messages
and forward them through to Matrix rooms.
Each real Matrix user is represented by an Twitch client,
and each real Twitch client is represented by a Matrix user.
Full Two-Way communication in channels and PMs are supported.
The Matrix users require to login first on Twitch before wrinting in Portal Rooms

## Usage

To join a channel on Twitch:

- Join a room with the alias
  ``#<alias_prefix><channel_name>:<homeserver_hosting_the_appservice>``
  e.g. ``#twitch_hc_dizee:example.com``.
  The template for this can be configured using the interactive config generator.

To send a Whisper to someone on Twitch:

- Start a conversation with a user ID
  ``@<user_prefix><channel_name>:<homeserver_hosting_the_appservice>``
  e.g. ``@twitch_hc_dizee:example.com``.
  The template for this can be configured using the interactive config generator.

## Contributing

Please see the [CONTRIBUTING](CONTRIBUTING.md) file for information on contributing.