package ui

import (
	"context"

	ui "github.com/gizak/termui/v3"

	"miwifi-termui/client"
)

// Read streams to update UI controllers.
type StreamStatRead <-chan client.Stat
type StreamBandRead <-chan client.Band

// Write streams to update UI controllers.
type StreamStatWrite chan<- client.Stat
type StreamBandWrite chan<- client.Band

// Controller is a drawable and resizable UI interface.
type Controller interface {
	ui.Drawable
	// Resize updates controller size.
	Resize()
	// Init initialises controller.
	Init(ctx context.Context)
}
