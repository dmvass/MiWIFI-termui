package app

import (
	"context"
	"fmt"
	"time"

	"github.com/gizak/termui/v3"
	"github.com/hashicorp/go-retryablehttp"
	log "github.com/sirupsen/logrus"

	"miwifi-termui/internal/client"
	"miwifi-termui/internal/ui"
)

// New creates and returns new application
func New(mac, host, username, password string, interval time.Duration, logger *log.Logger) *Application {
	var app Application

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

type Application struct {
	client   *client.Client
	logger   *log.Logger
	username string
	password string
	interval time.Duration
}

func (app *Application) Run(ctlName string) (code int) {

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

func (app *Application) startPollingStat(ctx context.Context, interval time.Duration) ui.StreamStatRead {
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

func (app *Application) startPollingBand(ctx context.Context, interval time.Duration) ui.StreamBandRead {
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
