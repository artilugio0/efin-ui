package main

import (
	"fyne.io/fyne/v2/widget"
)

type ModeLabel struct {
	*widget.Label
}

func NewModeLabel() *ModeLabel {
	e := &ModeLabel{
		Label: widget.NewLabel(""),
	}

	e.ExtendBaseWidget(e)

	return e
}

func (ml *ModeLabel) SetMode(mode Mode) {
	ml.SetText(mode.String())
	ml.Refresh()
}
