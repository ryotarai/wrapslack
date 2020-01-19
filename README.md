# wrapslack

wrapslack executes an arbitrary command and posts a notification to Slack. This works well with cron and systemd.timer

## Installation

Prebuilt binaries are available in https://github.com/ryotarai/wrapslack/releases

## Usage

Place config file:

```
$ touch /etc/wrapslack.yaml
$ chmod 600 /etc/wrapslack.yaml
$ vim /etc/wrapslack.yaml
```

Edit `wrapslack.yaml` as follows:

```yaml
# slack-token is generated from Slack Bots app
slack-token: "..."
slack-channel: "#channel"
```

Run an arbitrary command with wrapslack:

```
$ wrapslack -- <command> <args>...
```

When the command exited with non-zero, a notification will be posted to the Slack channel.

![slack notification](https://raw.githubusercontent.com/ryotarai/wrapslack/master/doc/images/notification.png)

### Exit status code to notify

By default, wrapslack posts Slack notification when the command exits with status code other than `0`. You can customize this status code by `--notify-exit-codes` and `--ignore-exit-codes` options.

### Notification message

Notification message posted in Slack can be configured by `--slack-message-template`.

### Slack username and icon

Username and icon can be configured by `--slack-icon` and `--slack-username`.

### All options

```
$ wrapslack -h
NAME:
   wrapslack - A new cli application

USAGE:
   wrapslack [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --version                       (default: false)
   --slack-token value              [$SLACK_TOKEN]
   --slack-channel value
   --slack-message-template value  (default: "`{{.hostName}}`: `{{.command}}` exited with `{{.exitCode}}`")
   --slack-icon-emoji value        (default: ":robot_face:")
   --slack-username value          (default: "wrapslack")
   --notify-exit-codes value       Comma-separated exit status codes to notify to Slack (empty '' for all status)
   --ignore-exit-codes value       Comma-separated exit status codes not to notify to Slack (default: "0")
   --help, -h                      show help (default: false)
```

