# GoUpBox

The update manager systray written in go.

Props:

* cross-platform
* simple
* fast
* configurable
* lightweight
* no runtime dependencies

## Usage

## Settings

```toml
# Name of your application
AppName=""
```

## Modules

It uses few cross-platform (Win, Linux, MacOs) modules:

* [systray](https://github.com/getlantern/systray)
* [go-autostart](https://github.com/emersion/go-autostart)
* [open-golang](https://github.com/skratchdot/open-golang)

It uses other useful native modules also:

* [go-toml](https://github.com/pelletier/go-toml)
* [gocrasy/rsync](https://github.com/gokrazy/rsync)
* [lumberjack](https://gopkg.in/natefinch/lumberjack.v2) - log rotation