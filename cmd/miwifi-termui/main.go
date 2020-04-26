package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/gizak/termui/v3"
	"github.com/hashicorp/go-retryablehttp"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"

	"miwifi-termui/client"
	"miwifi-termui/ui"
)

var version = "n/a"

func newApp(mac, host, username, password string, interval time.Duration, logger *log.Logger) *application {
	var app application

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 3
	retryClient.Logger = logger

	app.client = client.New(mac, host, retryClient.StandardClient())
	app.logger = logger

	app.username = username
	app.password = password

	app.interval = interval

	return &app
}

type application struct {
	client   *client.Client
	logger   *log.Logger
	username string
	password string
	interval time.Duration
}

func (app *application) Run(ctlName string) (code int) {

	defer func() {
		if err := recover(); err != nil {
			app.logger.Error(fmt.Sprintf("panic recover: %s", err))
			code = 1
		}
	}()

	app.logger.Debug("Running application")

	fmt.Println("Connection...")
	if err := app.client.Login(app.username, app.password); err != nil {
		fmt.Println("Connection error")
		app.logger.Error(err)
		return 1
	}

	if err := termui.Init(); err != nil {
		app.logger.Error(fmt.Sprintf("failed to initialize termui: %v", err))
		return 1
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var controller ui.Controller

	switch ctlName {
	case "dash":
		controller = ui.NewDashboard(app.startPollingStat(ctx, app.interval), app.startPollingBand(ctx, app.interval))
	case "net":
		controller = ui.NewNETController(app.startPollingStat(ctx, app.interval), app.startPollingBand(ctx, app.interval))
	case "cpu":
		controller = ui.NewCPUController(app.startPollingStat(ctx, app.interval))
	case "dev":
		controller = ui.NewDevController(app.startPollingStat(ctx, app.interval))
	case "info":
		controller = ui.NewInfoController(app.startPollingStat(ctx, app.interval))
	case "mem":
		controller = ui.NewMEMController(app.startPollingStat(ctx, app.interval))
	default:
		app.logger.Fatal(fmt.Sprintf("Invalid ui controller name: %s", ctlName))
	}

	app.logger.Debug("Init UI controller: " + ctlName)

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
		return 1
	}

	return 0
}

func (app *application) startPollingStat(ctx context.Context, interval time.Duration) ui.StreamStatRead {
	stream := make(chan client.Stat, 1)

	go func() {

		defer func() {
			close(stream)

			if err := recover(); err != nil {
				app.logger.Error(fmt.Sprintf("panic recover: %s", err))
			}
		}()

		stat, err := app.client.Status()
		if err != nil {
			app.logger.Error(err)
		}
		stream <- stat

		tick := time.Tick(interval)

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

func (app *application) startPollingBand(ctx context.Context, interval time.Duration) ui.StreamBandRead {
	stream := make(chan client.Band, 1)

	go func() {

		defer func() {
			close(stream)

			if err := recover(); err != nil {
				app.logger.Error(fmt.Sprintf("panic recover: %s", err))
			}
		}()

		band, err := app.client.BandwidthTest(true)
		if err != nil {
			app.logger.Error(err)
		}
		stream <- band

		tick := time.Tick(interval)

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
			if i.Flags&net.FlagUp != 0 && !bytes.Equal(i.HardwareAddr, nil) {
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
		intervalFlag = flag.Duration("interval", time.Second*10, "fetch data interval")
		uiFlag       = flag.String("ui", "dash", `ui controller {"dash", "cpu", "dev", "info", "mem", "net"}`)
	)

	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}

	var prompt bool

	if *hostFlag == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter host: ")
		host, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		*hostFlag = strings.TrimSpace(host)
		prompt = true
	}

	if !strings.HasPrefix(*hostFlag, "http://") {
		*hostFlag = "http://" + *hostFlag
	}

	if *passwordFlag == "" {
		fmt.Print("Enter admin password: ")
		bytePassword, err := terminal.ReadPassword(syscall.Stdin)
		if err != nil {
			panic(err)
		}
		*passwordFlag = strings.TrimSpace(string(bytePassword))
		prompt = true
	}

	if prompt {
		fmt.Println()
	}

	logger := log.New()
	logger.Out = ioutil.Discard

	if *debugFlag {
		logger.SetLevel(log.DebugLevel)

		file, err := os.OpenFile("MiWIFI-termui.out.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			panic(errors.New("failed to log to file"))
		}
		logger.Out = file
	}

	app := newApp(getMacAddr(), *hostFlag, *usernameFlag, *passwordFlag, *intervalFlag, logger)
	os.Exit(app.Run(*uiFlag))
}
