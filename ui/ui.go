package ui

import (
	"context"

	ui "github.com/gizak/termui/v3"

	"miwifi-cli/client"
)

type StreamStatRead <-chan client.Stat
type StreamBandRead <-chan client.Band

type StreamStatWrite chan<- client.Stat
type StreamBandWrite chan<- client.Band

type Controller interface {
	ui.Drawable

	Resize()
	Init(ctx context.Context)
}
