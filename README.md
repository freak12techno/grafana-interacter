# grafana-interacter

![Latest release](https://img.shields.io/github/v/release/Freak12techno/grafana-interacter)
[![Actions Status](https://github.com/Freak12techno/grafana-interacter/workflows/test/badge.svg)](https://github.com/Freak12techno/grafana-interacter/actions)

grafana-interacter is a tool to interact with your Grafana instance via a Telegram bot. Here's what you can currently do with it:
- Render dashboards panels and receive them as images

## How can I set it up?

Prerequisite: You need the [`grafana-image-renderer`](https://grafana.com/grafana/plugins/grafana-image-renderer/) plugin for rendering dashboards.

First of all, you need to download the latest release from [the releases page](https://github.com/Freak12techno/grafana-interacter/releases/). After that, you should unzip it and you are ready to go:

```sh
wget <the link from the releases page>
tar xvfz grafana-interacter-*
./grafana-interacter
```

That's not really interesting, what you probably want to do is to have it running in the background. For that, first of all, we have to copy the file to the system apps folder:

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
sudo systemctl enable grafana-interacter
sudo systemctl start grafana-interacter
sudo systemctl status grafana-interacter # validate it's running
```

If you need to, you can also see the logs of the process:

```sh
sudo journalctl -u grafana-interacter -f --output cat
```

## How does it work?

It queries 

## How can I configure it?

All configuration is executed via a `.yml` config, which is passed to

## How can I contribute?

Bug reports and feature requests are always welcome! If you want to contribute, feel free to open issues or PRs.
