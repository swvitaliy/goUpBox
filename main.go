package main

import (
	"bytes"
	"fmt"
	"github.com/emersion/go-autostart"
	"github.com/getlantern/systray"
	"github.com/pelletier/go-toml/v2"
	"github.com/skratchdot/open-golang/open"
	"gopkg.in/natefinch/lumberjack.v2"
	"goupbox/icon"
	"html/template"
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
	MaxAge:     14, //days
}

var cfg struct {
	AppName                   string
	AppDirectory              string
	AppUrl                    string
	CheckForUpdatesVersionUrl string
	Platform                  string
	RsyncArgs                 []string
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
	if len(cfg.RsyncArgs) == 0 {
		log.Fatal("No rsync arguments specified...")
		return
	}
	RsyncMain(cfg.RsyncArgs, os.Stdin, os.Stdout, os.Stderr)
}

type CheckUrlParams struct {
	Platform string
}

func checkForUpdates(params CheckUrlParams) (status bool, remoteVersion string, localVersion string) {
	// Replace in url template with params
	tmpl, err := template.New("checkForUpdates").Parse(cfg.CheckForUpdatesVersionUrl)
	if err != nil {
		log.Fatal(err)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, params)
	verUrl := buf.String()

	log.Printf("Check for remote version file by url %s", verUrl)
	resp, err := http.Get(verUrl)
	if err != nil {
		log.Printf("There isn't remote version file by url %s", verUrl)
		log.Println(err)
		return false, "", ""
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	remoteVersion = string(body)

	// When remote version is empty, it means that there is no remote version file
	if resp.StatusCode != 200 {
		log.Println("Error: " + remoteVersion)
		return false, "", ""
	}

	f, err := os.Open(path.Join(cfg.AppDirectory, "VERSION"))
	if err != nil {
		log.Printf("There isn't local version file  %s.", path.Join(cfg.AppDirectory, "VERSION"))
		//log.Fatal(err)
		if stat, err2 := os.Stat(cfg.AppDirectory); err2 != nil {
			if os.IsNotExist(err2) {
				log.Printf("Local project directory is NOT exists %s...", cfg.AppDirectory)
				log.Printf("Returns it has NOT a new remote version %s...", remoteVersion)
				return false, "", ""
			} else {
				log.Fatal(err2)
			}
		} else {
			if !stat.IsDir() {
				log.Printf("Local project directory \"%s\" not detected as a diectory", cfg.AppDirectory)
				panic(1)
			}
			log.Printf("Local project directory is exists %s...", cfg.AppDirectory)
			log.Printf("Returns it has a new remote version %s...", remoteVersion)
			return true, remoteVersion, ""
		}
	}
	defer f.Close()

	fileBytes, err := ioutil.ReadAll(f)
	if err != nil {
		log.Println("Read local version file error")
		log.Println(err)
		return false, "", ""
	}

	localVersion = string(fileBytes)

	return remoteVersion != localVersion, remoteVersion, localVersion
}

func onReady() {
	systray.SetTemplateIcon(icon.Data, icon.Data)
	systray.SetTitle("")
	systray.SetTooltip(cfg.AppName)

	currentPlatform := cfg.Platform // can be replaced by autodetect
	log.Printf("Platform: %s", currentPlatform)

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
				params := CheckUrlParams{Platform: currentPlatform}
				if status, remoteVersion, localVersion := checkForUpdates(params); status {
					if localVersion == "" {
						log.Printf("There is no local version file...")
						log.Println("Found remote version of " + cfg.AppName + ":" + remoteVersion)
						mUpdate.SetTitle("Download application \"" + cfg.AppName + "\" of version " + remoteVersion)
					} else {
						log.Println("Found new version of " + cfg.AppName + ":" + remoteVersion)
						mUpdate.SetTitle("Update to new version of " + cfg.AppName + ":" + remoteVersion)
					}
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
