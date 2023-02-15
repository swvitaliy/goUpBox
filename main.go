package main

import (
	"fmt"
	"github.com/emersion/go-autostart"
	"github.com/getlantern/systray"
	"github.com/pelletier/go-toml/v2"
	"github.com/skratchdot/open-golang/open"
	"gopkg.in/natefinch/lumberjack.v2"
	"goupbox/gokr-rsync"
	"goupbox/icon"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"time"
)

var lumberJackLog = &lumberjack.Logger{
	Filename:   "./logs/gupbox.log",
	MaxSize:    100, // megabytes
	MaxBackups: 2,
	MaxAge:     7, //days
}

var cfg struct {
	AppName                   string
	AppDirectory              string
	AppUrl                    string
	CheckForUpdatesVersionUrl string
	RSyncArgs                 []string
}

var app *autostart.App

func main() {
	f, err := os.Open("settings.toml")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	bDoc, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	err = toml.Unmarshal(bDoc, &cfg)
	if err != nil {
		panic(err)
	}
	fmt.Println("AppName:", cfg.AppName)

	log.Printf("Switching logs output to '%s'...", lumberJackLog.Filename)
	log.SetOutput(lumberJackLog)

	app = &autostart.App{
		Name:        cfg.AppName,
		DisplayName: cfg.AppName,
		//Exec:        []string{"sh", "-c", "echo autostart >> ~/autostart.txt"},
	}

	onExit := func() {
		now := time.Now()
		log.Printf("%d %s", now.UnixNano(), now.String())
	}

	systray.Run(onReady, onExit)
}

func onStartup(enabling bool) {
	if enabling {
		if app.IsEnabled() {
			log.Println(cfg.AppName + " is already enabled...")
		} else {
			log.Println("Enabling " + cfg.AppName + "...")

			if err := app.Enable(); err != nil {
				log.Fatal(err)
			}
		}
	} else {
		if app.IsEnabled() {
			log.Println("Disabling " + cfg.AppName + "...")

			if err := app.Disable(); err != nil {
				log.Fatal(err)
			}
		}
	}
}

func sync() {
	if len(cfg.RSyncArgs) == 0 {
		log.Fatal("No rsync arguments specified...")
		return
	}
	rsync.Main(cfg.RSyncArgs, os.Stdin, os.Stdout, os.Stderr)
}

func checkForUpdates() (bool, string) {
	resp, err := http.Get(cfg.CheckForUpdatesVersionUrl)
	if err != nil {
		log.Println(err)
		return false, ""
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != 200 {
		log.Println("Error: " + string(body))
		return false, ""
	}

	f, err := os.Open(path.Join(cfg.AppDirectory, "VERSION"))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	localVersion, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	return string(body) != string(localVersion), string(body)
}

func onReady() {
	systray.SetTemplateIcon(icon.Data, icon.Data)
	systray.SetTitle("")
	systray.SetTooltip(cfg.AppName)

	// We can manipulate the systray in other goroutines
	go func() {
		systray.SetTemplateIcon(icon.Data, icon.Data)
		systray.SetTitle("")
		systray.SetTooltip(cfg.AppName)

		mAutostart := systray.AddMenuItemCheckbox("Launch on Startup", "Launch on Startup", app.IsEnabled())

		mCheckForUpdates := systray.AddMenuItem("Check for Updates...", "Check for Updates")
		mUpdate := systray.AddMenuItem("Update...", "Update")

		checkForUpdates1 := func() {
			mCheckForUpdates.SetTitle("Checking...")
			go func() {
				defer func() {
					mCheckForUpdates.SetTitle("Check for Updates...")
					mCheckForUpdates.Enable()
				}()
				mCheckForUpdates.Disable()
				if status, version := checkForUpdates(); status {
					log.Println("Found new version of " + cfg.AppName + ":" + version)
					mUpdate.SetTitle("Update to new version of " + cfg.AppName + ":" + version)
					mUpdate.Enable()
				} else {
					log.Println("No new version of " + cfg.AppName)
					mUpdate.SetTitle("No new version of " + cfg.AppName)
					mUpdate.Disable()
				}
			}()
		}

		checkForUpdates1()

		//subMenuTop := systray.AddMenuItem("SubMenuTop", "SubMenu Test (top)")
		//subMenuMiddle := subMenuTop.AddSubMenuItem("SubMenuMiddle", "SubMenu Test (middle)")
		//subMenuBottom2 := subMenuMiddle.AddSubMenuItem("SubMenuBottom - Panic!", "SubMenu Test (bottom)")

		mUrl := systray.AddMenuItem("Help", "Open in Browser")

		systray.AddSeparator()
		mQuit := systray.AddMenuItem("Quit", "Quit")
		for {
			select {
			case <-mAutostart.ClickedCh:
				onStartup(!app.IsEnabled())
				if app.IsEnabled() {
					mAutostart.Check()
				} else {
					mAutostart.Uncheck()
				}
				break
			case <-mCheckForUpdates.ClickedCh:
				checkForUpdates1()
				break
			case <-mUpdate.ClickedCh:
				sync()
				checkForUpdates1()
				break
			case <-mUrl.ClickedCh:
				open.Run(cfg.AppUrl)
				break
			case <-mQuit.ClickedCh:
				fmt.Println("Requesting quit")
				systray.Quit()
				fmt.Println("Quit now...")
				return
			}
		}
	}()
}
