package ui

import (
	"context"
	"fmt"
	"sync"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"

	"miwifi-termui/client"
)

// NewMEMController creates and returns memory status UI controller.
func NewMEMController(streamStat StreamStatRead) *memController {
	return &memController{
		Grid:       ui.NewGrid(),
		bodyPlot:   widgets.NewPlot(),
		footText:   widgets.NewParagraph(),
		streamStat: streamStat,
	}
}

type memController struct {
	*ui.Grid

	bodyPlot *widgets.Plot
	footText *widgets.Paragraph

	streamStat StreamStatRead
	once       sync.Once
}

func (c *memController) Resize() {
	w, h := ui.TerminalDimensions()
	c.Grid.SetRect(0, 0, w, h)
}

func (c *memController) Init(ctx context.Context) {
	c.initUI()
	go c.subscribe(ctx)
}

func (c *memController) initUI() {
	c.bodyPlot.Title = "Storage"
	c.bodyPlot.Data = [][]float64{make([]float64, 2)}
	c.bodyPlot.AxesColor = ui.ColorWhite
	c.bodyPlot.LineColors[0] = ui.ColorGreen
	c.bodyPlot.MaxVal = 1.0

	c.footText.Border = false

	c.Grid.Set(
		ui.NewRow(.8, c.bodyPlot),
		ui.NewRow(.2, c.footText),
	)
}

func (c *memController) update(s client.Stat) {
	if len(c.bodyPlot.Data[0]) >= c.bodyPlot.Dx() {
		c.bodyPlot.Data[0] = c.bodyPlot.Data[0][1:]
	}
	c.bodyPlot.Data[0] = append(c.bodyPlot.Data[0], s.Mem.Usage)

	format := "Storage: %s | Usage: %.2f%% | Type: %s | Frequency: %s"
	c.footText.Text = fmt.Sprintf(format, s.Mem.Total, s.Mem.Usage*100.0, s.Mem.Type, s.Mem.Hz)
}

func (c *memController) subscribe(ctx context.Context) {
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
