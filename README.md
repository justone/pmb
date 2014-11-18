# Personal Message Bus

The Personal Message Bus (PMB) is a system that provides a means of communication between applications (aka agents) for personal productivity.  It ties together these agents securely, allowing your life to be more asynchronous.

# Architecture

The backbone of the PMB is a [RabbitMQ](http://www.rabbitmq.com/) service.  It provides a single global messaging topic that all agents send messages to and receive messages from. Individual agents connect and disconnect frequently and there is no durability of messages.

Each agent connects via Introduction (see below) and after connecting, every message is encrypted with AES256 using a shared key. Agents are free to send messages and react to incoming messages.

## Introduction

The purpose of introduction is to allow quick and easy distribution of the shared encryption key.

First, the introducer agent is run locally:

```
$ pmb introducer
```

Then, an unauthenticated agent is run on remote server (this example runs the `long_process.rb` script and sends a notification when it completes):

```
$ pmb notify -- long_process.rb
```

What happens next is introduction:

1. The notify agent sends an unencrypted message of type "RequestAuth" and prompts for the encryption key.
2. The introducer reacts to the "RequestAuth" message by copying the key into the local clipboard.
3. The user then simply pastes the key into the prompt and the notify agent can now send secure encrypted messages.

In the above case, the notify agent will send a "DisplayNotification" message when the `long_process.rb` process finishes. The introducer agent reacts to that by displaying an appropriate message to the user (e.g. via Growl).

# Setup

Setup documentation is coming.

# Status

Currently, PMB can do the following:

* Copy data from remote systems into your local clipboard
* Notify you when a long running program finishes
* Send arbitrary notifications from remote systems to your desktop
* Run an arbitrary command as a plugin for consuming or producing messages

Plans for the future:

* Wait to run a command until a particular message is received
* Tail a log file and send each line as a message and then collect those messages on another system for display
* Redeliver high priority notifications to your phone if your laptop is asleep
* A Go API for integrating any application into the bus

# Thank you

This project wouldn't be possible without the following libraries:

* [amqp](https://github.com/streadway/amqp)
* [go-flags](https://github.com/jessevdk/go-flags)
* [loggo](https://github.com/loggo/loggo)
* [osext](http://godoc.org/bitbucket.org/kardianos/osext)

# License

Copyright © 2014 Nate Jones

Distributed under the Apache License Version 2.0.
