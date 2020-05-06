package ui

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"

	"miwifi-termui/internal/client"
	"miwifi-termui/internal/humanize"
)

const maxDevices = 16

// NewDevController creates and returns devices status UI controller.
func NewDevController(streamStat StreamStatRead) *devController {
	return &devController{
		Grid:       ui.NewGrid(),
		bodyChart:  widgets.NewPieChart(),
		bodyTable:  widgets.NewTable(),
		footText:   widgets.NewParagraph(),
		streamStat: streamStat,
	}
}

type devController struct {
	*ui.Grid

	bodyChart *widgets.PieChart
	bodyTable *widgets.Table
	footText  *widgets.Paragraph

	streamStat StreamStatRead

	once sync.Once
}

func (c *devController) Resize() {
	w, h := ui.TerminalDimensions()
	c.Grid.SetRect(0, 0, w, h)
}

func (c *devController) Init(ctx context.Context) {
	c.initUI()
	go c.subscribe(ctx)
}

func (c *devController) initUI() {
	c.bodyChart.Title = "Connected devices bandwidth"
	c.bodyChart.LabelFormatter = func(i int, v float64) string { return strconv.Itoa(i + 1) }
	c.bodyChart.Data = make([]float64, maxDevices)

	c.bodyTable.Rows = make([][]string, maxDevices+1)
	c.bodyTable.Rows[0] = []string{"Name", "Value", "Percent"}

	c.footText.Border = false

	c.Grid.Set(
		ui.NewRow(.8,
			ui.NewCol(.4, c.bodyChart),
			ui.NewCol(.6, c.bodyTable),
		),
		ui.NewRow(.2, c.footText),
	)
}

func (c *devController) update(s client.Stat) {
	var totalDownload float64

	c.bodyChart.Data = c.bodyChart.Data[:len(s.Devices)]

	for i, device := range s.Devices {
		if i < maxDevices {
			c.bodyChart.Data[i] = float64(device.Download)
		}
		totalDownload += float64(device.Download)
	}

	c.bodyTable.Rows = c.bodyTable.Rows[:len(s.Devices)+1]

	for i, device := range s.Devices {
		if i >= maxDevices {
			break
		}
		c.bodyTable.Rows[i+1] = []string{
			fmt.Sprintf("[%d] %s", i+1, device.Name),
			humanize.Bytes(device.Download),
			fmt.Sprintf("%.2f%%", float64(device.Download)*100/totalDownload),
		}
	}

	c.footText.Text = fmt.Sprintf(
		"Total downloaded: %s | Total uploaded: %s | Devices: %d",
		humanize.Bytes(s.WAN.Download),
		humanize.Bytes(s.WAN.Upload),
		len(s.Devices),
	)
}

func (c *devController) subscribe(ctx context.Context) {
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
