package efinui

import (
	"fmt"
	"image/color"
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type HelpDialog struct {
	widget.BaseWidget

	content    *fyne.Container
	background *canvas.Rectangle
	container  *fyne.Container

	descriptions map[string]string
}

func NewHelpDialog(descriptions map[string]string) *HelpDialog {
	hd := &HelpDialog{
		descriptions: descriptions,
	}

	hd.content = container.NewVBox(widget.NewLabel("Help"))
	hd.background = canvas.NewRectangle(color.RGBA{R: 255, A: 255})
	hd.container = container.NewCenter(container.NewStack(hd.background, hd.content))

	hd.ExtendBaseWidget(hd)

	return hd
}

func (hd *HelpDialog) CreateRenderer() fyne.WidgetRenderer {
	bgColor := hd.Theme().Color(ColorNameFloatingBackground, theme.VariantDark)
	hd.background.FillColor = bgColor

	return widget.NewSimpleRenderer(hd.container)
}

func (hd *HelpDialog) SetDescriptions(descriptions map[string]string) {
	hd.descriptions = descriptions
	hd.updateContent()
}

func (hd *HelpDialog) updateContent() {
	keyList := []string{}
	for k, _ := range hd.descriptions {
		keyList = append(keyList, k)
	}
	slices.Sort(keyList)

	descLabels := make([]fyne.CanvasObject, len(hd.descriptions)+1)
	descLabels[0] = widget.NewLabel("Help")

	for i, k := range keyList {
		descLabels[i+1] = widget.NewLabel(fmt.Sprintf("%s: %s", k, hd.descriptions[k]))
	}

	hd.content.Objects = descLabels
	hd.content.Refresh()
}

func (hd *HelpDialog) Refresh() {
	bgColor := hd.Theme().Color(ColorNameFloatingBackground, theme.VariantDark)
	hd.background.FillColor = bgColor

	hd.BaseWidget.Refresh()
}
