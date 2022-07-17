# grafana-interacter

![Latest release](https://img.shields.io/github/v/release/Freak12techno/grafana-interacter)
[![Actions Status](https://github.com/Freak12techno/grafana-interacter/workflows/test/badge.svg)](https://github.com/Freak12techno/grafana-interacter/actions)

grafana-interacter is a tool to interact with your Grafana instance via a Telegram bot. Here's the list of currently supported commands:
- `/render [<opts>] <panel name>` - renders the panel and sends it as image. If there are multiple panels with the same name (for example, you have a `dashboard1` and `dashboard2` both containing panel with name `panel`), it will render the first panel it will find. For specifying it, you may add the dashboard name as a prefix to your query (like `/render dashboard1 panel`). You can also provide options in a `key=value` format, which will be internally passed to a `/render` query to Grafana. Some examples are `from`, `to`, `width`, `height` (the command would look something like `/render from=now-14d to=now-7d width=100 height=100 dashboard1 panel`). By default, the params are: `width=1000&height=500&from=now-30m&to=now&tz=Europe/Moscow`.
- `/dashboards` - will list Grafana dashboards and links to them.
- `/dashboard <name>` - will return a link to a dashboard and its panels.
- `/datasources` - will return Grafana datasources.
- `/alerts` - will list both Grafana alerts and Prometheus alerts from all Prometheus datasources, if any
- `/silence <duration> <params>` - creates a silence for Grafana alert. You need to pass a duration (like `/silence 2h test alert`) and some params for matching alerts to silence. You may use `=` for matching the value exactly (example: `/silence 2h host=localhost`), `!=` for matching everything except this value (example: `/silence 2h host!=localhost`), `=~` for matching everything that matches the regexp (example: `/silence 2h host=~local`), , `!~` for matching everything that doesn't the regexp (example: `/silence 2h host!~local`), or just provide a string that will be treated as an alert name (example: `/silence 2h test alert`).
- `/silences` - list silences (both active and expired).
- `/unsilence <silence ID>` - deletes a silence.
- `/alertmanager_silences` - same as `/silences`, but using external Alertmanager.
- `/alertmanager_silence` - same as `/silence`, but using external Alertmanager.
- `/alertmanager_unsilence` - same as `/unsilence`, but using external Alertmanager.

## How can I set it up?

Prerequisite: You need Grafana itself with new alerting enabled, as well as the [`grafana-image-renderer`](https://grafana.com/grafana/plugins/grafana-image-renderer/) plugin for rendering dashboards.

Before starting, you need to create a Telegram bot. Go to @Botfather at Telegram and create a new bot there. For bot commands, put the following:

```
render - Render a panel
dashboards - List dashboards
dashboard - See dashboard and its panels
alerts - See alerts
firing - See firing and pending alerts
datasources - See Grafana datasources
silence - Creates a new silence
silences - List all silences
unsilence - Deletes a silence
alertmanager_silence - Creates a new Alertmanager silence
alertmanager_silences - List all Alertmanager silences
alertmanager_unsilence - Deletes an Alertmanager silence
```

Save the bot token somewhere, you'll need it later to for grafana-interacter to function.

Then, you need to download the latest release from [the releases page](https://github.com/Freak12techno/grafana-interacter/releases/). After that, you should unzip it and you are ready to go:

```sh
wget <the link from the releases page>
tar xvfz grafana-interacter-*
./grafana-interacter --config <path to config>
```

What you probably want to do is to have it running in the background in a detached mode. For that, first of all, we have to copy the file to the system apps folder:

```sh
sudo cp ./grafana-interacter /usr/bin
```

Then we need to create a systemd service for our app:

```sh
sudo nano /etc/systemd/system/grafana-interacter.service
```

You can use this template (change the user to whatever user you want this to be executed from. It's advised to create a separate user for that instead of running it from root):

```
[Unit]
Description=grafana-interacter
After=network-online.target

[Service]
User=<username>
TimeoutStartSec=0
CPUWeight=95
IOWeight=95
ExecStart=grafana-interacter --config <path to config>
Restart=always
RestartSec=2
LimitNOFILE=800000
KillSignal=SIGTERM

[Install]
WantedBy=multi-user.target
```

Then we'll add this service to the autostart and run it:

```sh
sudo systemctl enable grafana-interacter # set it to start on system load
sudo systemctl start grafana-interacter  # start it
sudo systemctl status grafana-interacter # validate it's running
```

If you need to, you can also see the logs of the process:

```sh
sudo journalctl -u grafana-interacter -f --output cat
```

## How does it work?

It queries Grafana via its API and returns the data as a Telegram message.

## How can I configure it?

All configuration is executed via a `.yml` config, which is passed as a `--config` variable. Check out `config.example.yml` for reference.

## How can I contribute?

Bug reports and feature requests are always welcome! If you want to contribute, feel free to open issues or PRs.
