package main

import (
	"image/color"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type Toast struct {
	widget.BaseWidget

	message string

	bg    *canvas.Rectangle
	label *widget.Label

	content *fyne.Container
}

func NewToast(message string, bgColor color.Color) *Toast {
	t := &Toast{
		message: message,
		label:   widget.NewLabel(message),
		bg:      canvas.NewRectangle(bgColor),
	}
	t.ExtendBaseWidget(t)

	return t
}

func (t *Toast) CreateRenderer() fyne.WidgetRenderer {
	content := container.NewStack(t.bg, t.label)
	return widget.NewSimpleRenderer(content)
}

type ToastSet struct {
	widget.BaseWidget

	lock sync.Mutex

	content *fyne.Container

	duration time.Duration
}

func NewToastSet() *ToastSet {
	ts := &ToastSet{
		lock: sync.Mutex{},
		content: container.NewVBox(
			layout.NewSpacer(),
		),
		duration: 3 * time.Second,
	}

	ts.ExtendBaseWidget(ts)

	return ts
}

func (ts *ToastSet) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(ts.content)
}

func (ts *ToastSet) CreateToastMessage(message string) {
	ts.createToast(message, color.RGBA{R: 30, G: 80, B: 100, A: 200})
}

func (ts *ToastSet) CreateToastError(message string) {
	ts.createToast(message, color.RGBA{R: 150, G: 30, B: 50, A: 200})
}

func (ts *ToastSet) createToast(message string, bgColor color.Color) {
	ts.lock.Lock()
	defer ts.lock.Unlock()

	toast := container.NewHBox(layout.NewSpacer(), NewToast(message, bgColor))
	newObjects := append([]fyne.CanvasObject{}, ts.content.Objects[:len(ts.content.Objects)-1]...)
	newObjects = append(newObjects, toast, layout.NewSpacer())

	ts.content.Objects = newObjects
	ts.content.Refresh()

	time.AfterFunc(ts.duration, func() {
		fyne.Do(func() {
			ts.lock.Lock()
			defer ts.lock.Unlock()

			newObjects := []fyne.CanvasObject{}
			for _, o := range ts.content.Objects {
				if o == toast {
					continue
				}

				newObjects = append(newObjects, o)
			}

			ts.content.Objects = newObjects
			ts.content.Refresh()
		})
	})
}
