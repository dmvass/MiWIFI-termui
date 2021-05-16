package ui

import (
	"context"
	"sync"

	ui "github.com/gizak/termui/v3"

	"miwifi-termui/client"
)

func NewDashboard(streamStat StreamStatRead, streamBand StreamBandRead) *dashboardController {
	ctl := &dashboardController{
		Grid:       ui.NewGrid(),
		streamStat: streamStat,
		streamBand: streamBand,
	}

	devStreamsStat := make(chan client.Stat, 1)
	ctl.dev = NewDevController(devStreamsStat)
	ctl.streamsStat = append(ctl.streamsStat, devStreamsStat)

	netStreamsStat := make(chan client.Stat, 1)
	netStreamsBand := make(chan client.Band, 1)
	ctl.net = NewNETController(netStreamsStat, netStreamsBand)
	ctl.streamsStat = append(ctl.streamsStat, netStreamsStat)
	ctl.streamsBand = append(ctl.streamsBand, netStreamsBand)

	cpuStreamsStat := make(chan client.Stat, 1)
	ctl.cpu = NewCPUController(cpuStreamsStat)
	ctl.streamsStat = append(ctl.streamsStat, cpuStreamsStat)

	memStreamsStat := make(chan client.Stat, 1)
	ctl.mem = NewMEMController(memStreamsStat)
	ctl.streamsStat = append(ctl.streamsStat, memStreamsStat)

	infoStreamsStat := make(chan client.Stat, 1)
	ctl.info = NewInfoController(infoStreamsStat)
	ctl.streamsStat = append(ctl.streamsStat, infoStreamsStat)

	return ctl
}

type dashboardController struct {
	*ui.Grid

	dev  Controller
	net  Controller
	cpu  Controller
	mem  Controller
	info Controller

	streamStat StreamStatRead
	streamBand StreamBandRead

	streamsStat []StreamStatWrite
	streamsBand []StreamBandWrite

	once sync.Once
}

func (c *dashboardController) Resize() {
	for _, ctl := range []Controller{c.dev, c.net, c.cpu, c.mem, c.info} {
		ctl.Resize()
	}
	w, h := ui.TerminalDimensions()
	c.Grid.SetRect(0, 0, w, h)
}

func (c *dashboardController) Init(ctx context.Context) {
	for _, ctl := range []Controller{c.dev, c.net, c.cpu, c.mem, c.info} {
		ctl.Init(ctx)
	}
	c.initUI()
	go c.subscribe(ctx)
}

func (c *dashboardController) initUI() {
	c.Grid.Set(
		ui.NewRow(.1, c.info),
		ui.NewRow(.5,
			ui.NewCol(.5, c.net),
			ui.NewCol(.5, c.dev),
		),
		ui.NewRow(.4,
			ui.NewCol(.5, c.cpu),
			ui.NewCol(.5, c.mem),
		),
	)
}

func (c *dashboardController) subscribe(ctx context.Context) {
	c.once.Do(func() {
		for {
			select {
			case <-ctx.Done():
				return
			case s := <-c.streamStat:
				for _, stream := range c.streamsStat {
					stream <- s
				}
			case b := <-c.streamBand:
				for _, stream := range c.streamsBand {
					stream <- b
				}
			}
		}
	})
}
