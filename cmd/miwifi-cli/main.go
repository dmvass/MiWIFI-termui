package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"time"

	"github.com/gizak/termui/v3"
	log "github.com/sirupsen/logrus"

	"miwifi-cli/client"
	"miwifi-cli/ui"
)

var version = "n/a"

func newApp(mac, host, username, password string, logger *log.Logger) *application {
	var app application

	app.client = client.New(mac, host, nil)
	app.logger = logger

	app.username = username
	app.password = password

	return &app
}

type application struct {
	client *client.Client
	logger *log.Logger

	username string
	password string
}

func (app *application) Run(ctlName string) int {
	app.logger.Debug("Running application")

	fmt.Println("Connection...")
	if err := app.client.Login(app.username, app.password); err != nil {
		app.logger.Error(err)
		return 2
	}

	if err := termui.Init(); err != nil {
		app.logger.Error("failed to initialize termui: %v", err)
		return 2
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var controller ui.Controller

	switch ctlName {
	case "dash":
		controller = ui.NewDashboard(app.startPollingStat(ctx), app.startPollingBand(ctx))
	case "cpu":
		controller = ui.NewCPUController(app.startPollingStat(ctx))
	case "dev":
		controller = ui.NewDevController(app.startPollingStat(ctx))
	case "info":
		controller = ui.NewInfoController(app.startPollingStat(ctx))
	case "mem":
		controller = ui.NewMEMController(app.startPollingStat(ctx))
	case "net":
		controller = ui.NewNETController(app.startPollingStat(ctx), app.startPollingBand(ctx))
	default:
		app.logger.Fatal("Invalid ui controller name: %s", ctlName)
	}

	controller.Init(ctx)
	controller.Resize()

	termui.Render(controller)

	ev := termui.PollEvents()
	tick := time.Tick(time.Second)

Loop:
	for {
		select {
		case e := <-ev:
			switch {
			case e.Type == termui.KeyboardEvent && e.ID == "q":
				// quit on any keyboard event
				break Loop
			case e.Type == termui.ResizeEvent:
				controller.Resize()
			}
		case <-tick:
			termui.Render(controller)
		}
	}

	app.logger.Debug("Stopping application")

	termui.Close()

	fmt.Println("Disconnection...")
	if err := app.client.Logout(); err != nil {
		app.logger.Error(err)
		return 2
	}

	return 0
}

func (app *application) startPollingStat(ctx context.Context) ui.StreamStatRead {
	stream := make(chan client.Stat, 1)

	go func() {
		defer close(stream)

		stat, err := app.client.Status()
		if err != nil {
			app.logger.Error(err)
		}
		stream <- stat

		tick := time.Tick(time.Second * 5)

		for {
			select {
			case <-ctx.Done():
				return
			case <-tick:
				app.logger.Debug("Fetching status info")
				result, err := app.client.Status()
				if err != nil {
					app.logger.Error(err)
				} else {
					stream <- result
				}
			}
		}
	}()

	return stream
}

func (app *application) startPollingBand(ctx context.Context) ui.StreamBandRead {
	stream := make(chan client.Band, 1)

	go func() {
		defer close(stream)

		band, err := app.client.BandwidthTest(true)
		if err != nil {
			app.logger.Error(err)
		}
		stream <- band

		tick := time.Tick(time.Minute)

		for {
			select {
			case <-ctx.Done():
				return
			case <-tick:
				app.logger.Debug("Fetching bandwidth info")
				result, err := app.client.BandwidthTest(true)
				if err != nil {
					app.logger.Error(err)
				} else {
					stream <- result
				}
			}
		}
	}()

	return stream
}

func getMacAddr() (addr string) {
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, i := range interfaces {
			if i.Flags&net.FlagUp != 0 && bytes.Compare(i.HardwareAddr, nil) != 0 {
				addr = i.HardwareAddr.String()
				break
			}
		}
	}
	return
}

func main() {
	var (
		versionFlag  = flag.Bool("version", false, "application version")
		debugFlag    = flag.Bool("debug", false, "run application in debug mode")
		hostFlag     = flag.String("host", "", "MiWiFi host address")
		usernameFlag = flag.String("username", "admin", "username for login")
		passwordFlag = flag.String("password", "", "password for login")
		uiFlag       = flag.String("ui", "dash", "ui controller {dash, cpu, dev, info, mem, net}")
	)

	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}

	if *hostFlag == "" {
		fmt.Println("host is required option")
		os.Exit(1)
	}

	if !strings.HasPrefix(*hostFlag, "http://") {
		*hostFlag = "http://" + *hostFlag
	}

	if *passwordFlag == "" {
		fmt.Println("password is required option")
		os.Exit(1)
	}

	logger := log.New()
	filepath := "/var/log/miwifi-cli.log"
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		logger.Out = file
	} else {
		logger.Out = ioutil.Discard
		fmt.Printf("Failed to log to file %q, messages were been discarded\n", filepath)
	}

	if *debugFlag {
		logger.SetLevel(log.DebugLevel)
	}

	app := newApp(getMacAddr(), *hostFlag, *usernameFlag, *passwordFlag, logger)

	os.Exit(app.Run(*uiFlag))
}
