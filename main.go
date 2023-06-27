package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/getlantern/systray"
	"github.com/pelletier/go-toml/v2"
	"github.com/sergz72/go-autostart"
	"github.com/skratchdot/open-golang/open"
	"gopkg.in/natefinch/lumberjack.v2"
	"goupbox/icon"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
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
	execPath, err0 := os.Executable()
	if err0 != nil {
		log.Fatal(err0)
	}

	// --settings=/etc/goupbox.conf
	settingsPath := flag.String("settings", "", "path to settings file")
	logPath := flag.String("log", "", "path to log file")
	flag.Parse()

	var cfgFile string
	if *settingsPath != "" {
		cfgFile = *settingsPath
	} else {
		cfgFile = path.Join(filepath.Dir(execPath), "settings.toml")
	}

	f, err := os.Open(cfgFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if *logPath != "" {
		lumberJackLog.Filename = *logPath
	} else {
		lumberJackLog.Filename = path.Join(filepath.Dir(execPath), "logs", "goupbox.log")
	}

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

	exec := []string{execPath}
	if *settingsPath != "" {
		exec = append(exec, "-settings", *settingsPath)
	}
	if *logPath != "" {
		exec = append(exec, "-log", *logPath)
	}
	app = &autostart.App{
		Name:        cfg.AppName,
		DisplayName: cfg.AppName,
		Exec:        exec,
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

func resolveTemplate(name string, str string, params CheckUrlParams) string {
	t, err := template.New(name).Parse(str)
	if err != nil {
		log.Fatal(err)
	}
	var buf bytes.Buffer
	err = t.Execute(&buf, params)
	if err != nil {
		log.Fatal(err)
	}
	return buf.String()
}

func checkForUpdates(params CheckUrlParams) (status bool, remoteVersion string, localVersion string) {
	// Replace in url template with params
	verUrl := resolveTemplate("checkForUpdatesVersionUrl", cfg.CheckForUpdatesVersionUrl, params)

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

	localPath := resolveTemplate("localPath", cfg.AppDirectory, params)
	localVersionPath := path.Join(localPath, "VERSION")
	f, err := os.Open(localVersionPath)
	if err != nil {
		log.Printf("There isn't local version file  %s.", path.Join(cfg.AppDirectory, "VERSION"))
		//log.Fatal(err)
		if stat, err2 := os.Stat(localPath); err2 != nil {
			if os.IsNotExist(err2) {
				log.Printf("Local project directory is NOT exists %s...", localPath)
				log.Printf("Returns it has NOT a new remote version %s...", remoteVersion)
				return false, "", ""
			} else {
				log.Fatal(err2)
			}
		} else {
			if !stat.IsDir() {
				log.Printf("Local project directory \"%s\" not detected as a diectory", localPath)
				panic(1)
			}
			log.Printf("Local project directory is exists %s...", localPath)
			log.Printf("Returns it has a new remote version %s...", remoteVersion)
			return true, remoteVersion, ""
		}
	}
	defer f.Close()

	fileBytes, err := ioutil.ReadAll(f)
	if err != nil {
		log.Println("Error: read a local version")
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

		// TODO That if file VERSION updated and update process stopped by user => lock file
		// TODO Compare checksums of each downloaded file

		updateInProgress := func() {}
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
						log.Printf("Found remote version of %s: %s", cfg.AppName, remoteVersion)
						mUpdate.SetTitle(fmt.Sprintf("Download %s version %s", cfg.AppName, remoteVersion))

						updateInProgress = func() {
							mUpdate.Disable()
							mUpdate.SetTitle(fmt.Sprintf("Downloading %s...", cfg.AppName))
						}
					} else {
						log.Printf("Found new version of %s: %s\n", cfg.AppName, remoteVersion)
						mUpdate.SetTitle(fmt.Sprintf("Update %s version %s to %s", cfg.AppName, localVersion, remoteVersion))

						updateInProgress = func() {
							mUpdate.Disable()
							mUpdate.SetTitle(fmt.Sprintf("Updating %s: %s to %s...", cfg.AppName, localVersion, remoteVersion))
						}
					}
					mUpdate.Enable()
				} else {
					log.Printf("There isn't a new version %s", cfg.AppName)
					mUpdate.SetTitle(fmt.Sprintf("There isn't a new version %s", cfg.AppName))
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
				updateInProgress()
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
