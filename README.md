# GoUpBox

The update manager systray over rsync written in go.

Props:

* cross-platform
* simple
* fast
* configurable
* lightweight
* no runtime dependencies

This update manager put files of the new version into the directory with a previous version.
It uses rsync implementation written in go so it is really fast.

In practice, It has problems with the cross compilation, and It needed the same host os as the target.

I faced the open issue in systray module that didn't let me compile for macOS from linux host - https://github.com/getlantern/systray/issues/34

## Settings

See [settings.toml](./settings.toml) file for example.

## Requirements

Your app files should be available through rsync server and http server (nginx or else) same directory. 

It should has 2 types of version files:

- VERSION (fixed name of file) 
- version-1.2.3.txt (files that contains version value in the name of file)

The VERSION file contains latest version of your app (remote or local).
The "version-x.y.z.txt" new file saves each time when update happened. 
Command like `ls version-*.txt` returns list of installed versions.

## Modules

It uses few cross-platform (Win, Linux, MacOs) modules:

* [systray](https://github.com/getlantern/systray)
* [go-autostart](https://github.com/sergz72/go-autostart)
* [open-golang](https://github.com/skratchdot/open-golang)

It uses other useful native modules also:

* [go-toml](https://github.com/pelletier/go-toml)
* [gocrasy/rsync](https://github.com/gokrazy/rsync) (just copied sources)
* [lumberjack](https://gopkg.in/natefinch/lumberjack.v2) - log rotation