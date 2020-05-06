package ui

import (
	"context"
	"fmt"
	"sync"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"

	"miwifi-termui/internal/client"
)

// NewCPUController creates and returns CPU status UI controller.
func NewCPUController(streamStat StreamStatRead) *cpuController {
	return &cpuController{
		Grid:       ui.NewGrid(),
		bodyPlot:   widgets.NewPlot(),
		footText:   widgets.NewParagraph(),
		streamStat: streamStat,
	}
}

type cpuController struct {
	*ui.Grid

	bodyPlot *widgets.Plot
	footText *widgets.Paragraph

	streamStat StreamStatRead
	once       sync.Once
}

func (c *cpuController) Resize() {
	w, h := ui.TerminalDimensions()
	c.Grid.SetRect(0, 0, w, h)
}

func (c *cpuController) Init(ctx context.Context) {
	c.initUI()
	go c.subscribe(ctx)
}

func (c *cpuController) initUI() {
	c.bodyPlot.Title = "CPU"
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

func (c *cpuController) update(s client.Stat) {
	if len(c.bodyPlot.Data[0]) > c.bodyPlot.Dx() {
		c.bodyPlot.Data[0] = c.bodyPlot.Data[0][1:]
	}
	c.bodyPlot.Data[0] = append(c.bodyPlot.Data[0], s.CPU.Load)

	format := "CPU: %d | Load: %.2f%% | Core frequency: %s"
	c.footText.Text = fmt.Sprintf(format, s.CPU.Core, s.CPU.Load*100.0, s.CPU.Hz)
}

func (c *cpuController) subscribe(ctx context.Context) {
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
