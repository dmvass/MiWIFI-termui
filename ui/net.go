package ui

import (
	"context"
	"fmt"
	"sync"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"

	"miwifi-termui/client"
	"miwifi-termui/humanize"
)

// NewNETController creates and returns network status UI controller.
func NewNETController(streamStat StreamStatRead, streamBand StreamBandRead) *netController {
	return &netController{
		Grid:       ui.NewGrid(),
		headText:   widgets.NewParagraph(),
		bodyPlot:   widgets.NewPlot(),
		footText:   widgets.NewParagraph(),
		streamStat: streamStat,
		streamBand: streamBand,
	}
}

type netController struct {
	*ui.Grid

	headText *widgets.Paragraph
	bodyPlot *widgets.Plot
	footText *widgets.Paragraph

	streamStat StreamStatRead
	streamBand StreamBandRead
	once       sync.Once
}

func (c *netController) Resize() {
	w, h := ui.TerminalDimensions()
	c.Grid.SetRect(0, 0, w, h)
}

func (c *netController) Init(ctx context.Context) {
	c.initUI()
	go c.subscribe(ctx)
}

func (c *netController) initUI() {
	c.headText.Title = "Real-time network status"
	c.headText.PaddingTop = 1
	c.headText.PaddingLeft = 1

	c.bodyPlot.Data = [][]float64{
		make([]float64, 2),
		make([]float64, 2),
	}
	c.bodyPlot.AxesColor = ui.ColorWhite
	c.bodyPlot.LineColors[0] = ui.ColorGreen
	c.bodyPlot.LineColors[1] = ui.ColorBlue

	c.footText.Border = false

	c.Grid.Set(
		ui.NewRow(.2, c.headText),
		ui.NewRow(.6, c.bodyPlot),
		ui.NewRow(.2, c.footText),
	)
}

func (c *netController) update(s client.Stat, b client.Band) {
	c.headText.Text = fmt.Sprintf(
		"Downstream speed: %s/s | Upstream speed: %s/s",
		humanize.Bytes(s.WAN.DownSpeed),
		humanize.Bytes(s.WAN.UpSpeed),
	)

	if len(c.bodyPlot.Data[0]) >= c.bodyPlot.Dx() {
		c.bodyPlot.Data[0] = c.bodyPlot.Data[0][1:]
	}
	c.bodyPlot.Data[0] = append(c.bodyPlot.Data[0], float64(s.WAN.DownSpeed))

	if len(c.bodyPlot.Data[1]) >= c.bodyPlot.Dx() {
		c.bodyPlot.Data[1] = c.bodyPlot.Data[1][1:]
	}
	c.bodyPlot.Data[1] = append(c.bodyPlot.Data[1], float64(s.WAN.UpSpeed))

	c.footText.Text = fmt.Sprintf(
		"Bandwidth: %.2f m | Max download speed: %s/s",
		b.Bandwidth,
		humanize.Bytes(s.WAN.MaxDownloadSpeed),
	)
}

func (c *netController) subscribe(ctx context.Context) {
	c.once.Do(func() {
		var b client.Band
		for {
			select {
			case <-ctx.Done():
				return
			case b = <-c.streamBand:
			case s := <-c.streamStat:
				c.update(s, b)
			}
		}
	})
}
