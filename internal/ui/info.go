package ui

import (
	"context"
	"sync"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"

	"miwifi-termui/internal/client"
)

// NewInfoController creates and returns info UI controller.
func NewInfoController(streamStat StreamStatRead) *infoController {
	return &infoController{
		Grid:       ui.NewGrid(),
		bodyTable:  widgets.NewTable(),
		streamStat: streamStat,
	}
}

type infoController struct {
	*ui.Grid

	bodyTable *widgets.Table

	streamStat StreamStatRead
	once       sync.Once
}

func (c *infoController) Resize() {
	w, h := ui.TerminalDimensions()
	c.Grid.SetRect(0, 0, w, h)
}

func (c *infoController) Init(ctx context.Context) {
	c.initUI()
	go c.subscribe(ctx)
}

func (c *infoController) initUI() {
	c.bodyTable.Title = "Info"
	c.bodyTable.Rows = make([][]string, 2)
	c.bodyTable.Rows[0] = []string{"Platform", "System version", "MAC address", "SN"}

	c.Grid.Set(ui.NewRow(1.0, c.bodyTable))
}

func (c *infoController) update(s client.Stat) {
	c.bodyTable.Rows[1] = []string{
		s.Hardware.Platform,
		s.Hardware.Version,
		s.Hardware.Mac,
		s.Hardware.SN,
	}
}

func (c *infoController) subscribe(ctx context.Context) {
	c.once.Do(func() {
		for {
			select {
			case <-ctx.Done():
				return
			case s := <-c.streamStat:
				c.update(s)
			}
		}
	})
}
