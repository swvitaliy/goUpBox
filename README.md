# GoUpBox

The update manager in systray over rsync written in go.

Props:

* cross-platform
* simple
* fast
* configurable
* Lightweight
* no runtime dependencies

This update manager puts files of the new version into the directory with the previous version.
It uses rsync implementation written in go so it is pretty fast.

In practice, It has problems with cross-compilation, and It needs the same host OS as the target.

I faced an open issue in the systray module that didn't let me compile for macOS from a Linux host - https://github.com/getlantern/systray/issues/34

## Settings

See [settings.toml](./settings.toml) file.

## Requirements

Your app files should be available through the rsync server and http server (nginx or else) same directory. 

It should have 2 types of version files:

- VERSION (fixed name of file) 
- version-1.2.3.txt (files that contain version value in the name of the file)

The VERSION file contains the latest version of your app (remote or local).
The "version-x.y.z.txt" new file saves each time an update happens. 
A command like `ls version-*.txt` returns a list of installed versions.

## Modules

It uses a few cross-platform (Win, Linux, macOS) modules:

* [systray](https://github.com/getlantern/systray)
* [go-autostart](https://github.com/sergz72/go-autostart)
* [open-golang](https://github.com/skratchdot/open-golang)

It uses other useful native modules also:

* [go-toml](https://github.com/pelletier/go-toml)
* [gocrasy/rsync](https://github.com/gokrazy/rsync) (just copied sources)
* [lumberjack](https://gopkg.in/natefinch/lumberjack.v2) - log rotation
