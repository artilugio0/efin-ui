package efinui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type MultiSplit struct {
	widget.BaseWidget

	OnFocusMove func(fyne.CanvasObject)

	objectsGrid [][]fyne.CanvasObject

	layout Layout

	focusedIndex1 int
	focusedIndex2 int

	search              string
	searchCaseSensitive bool

	container *fyne.Container
}

func NewMultiSplit() *MultiSplit {
	flexGrid := NewFlexGrid([]int{}).WithPadding(3)
	ms := &MultiSplit{
		objectsGrid:   [][]fyne.CanvasObject{},
		layout:        LayoutRowColumn,
		focusedIndex1: 0,
		focusedIndex2: 0,
		container:     container.New(flexGrid),
	}

	ms.ExtendBaseWidget(ms)

	return ms
}

func (ms *MultiSplit) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(ms.container)
}

func (ms *MultiSplit) PaneLineAdd(o fyne.CanvasObject) {
	ms.objectsGrid = append(ms.objectsGrid, []fyne.CanvasObject{o})
	ms.focusedIndex1 = len(ms.objectsGrid) - 1
	ms.focusedIndex2 = 0

	ms.refreshContainer()
	ms.callOnFocusMove()
}

func (ms *MultiSplit) PaneCreate(o fyne.CanvasObject) {
	if ms.search != "" {
		if searchable, ok := o.(Searcher); ok {
			searchable.Search(ms.search, ms.searchCaseSensitive)
		}
	}

	if len(ms.objectsGrid) == 0 {
		ms.objectsGrid = [][]fyne.CanvasObject{[]fyne.CanvasObject{}}
		ms.focusedIndex1 = 0
		ms.focusedIndex2 = 0
	}

	ms.objectsGrid[ms.focusedIndex1] = append(ms.objectsGrid[ms.focusedIndex1], o)
	ms.focusedIndex2 = len(ms.objectsGrid[ms.focusedIndex1]) - 1

	ms.refreshContainer()
	ms.callOnFocusMove()
}

func (ms *MultiSplit) refreshContainer() {
	objects := []fyne.CanvasObject{}
	for i, g := range ms.objectsGrid {
		for j, o := range g {
			if i == ms.focusedIndex1 && j == ms.focusedIndex2 {
				o = ms.widgetWithFocusStyle(o)
			}
			objects = append(objects, o)
		}
	}

	groupsLenghts := make([]int, len(ms.objectsGrid))
	for i, g := range ms.objectsGrid {
		groupsLenghts[i] = len(g)
	}

	flexGrid := ms.container.Layout.(*FlexGrid)
	flexGrid.ItemsPerGroup = groupsLenghts

	ms.container.Objects = objects
	ms.container.Refresh()
}

func (ms *MultiSplit) PaneDelete() {
	if len(ms.objectsGrid) <= ms.focusedIndex1 || len(ms.objectsGrid[ms.focusedIndex1]) <= ms.focusedIndex2 {
		return
	}

	group := []fyne.CanvasObject{}
	for i, o := range ms.objectsGrid[ms.focusedIndex1] {
		if i == ms.focusedIndex2 {
			continue
		}

		group = append(group, o)
	}

	ms.objectsGrid[ms.focusedIndex1] = group

	if len(ms.objectsGrid[ms.focusedIndex1]) == 0 {
		objects := [][]fyne.CanvasObject{}
		for i, g := range ms.objectsGrid {
			if i == ms.focusedIndex1 {
				continue
			}

			objects = append(objects, g)
		}

		ms.objectsGrid = objects
	}

	if len(ms.objectsGrid) == 0 {
		ms.focusedIndex1 = 0
		ms.focusedIndex2 = 0
	} else {
		ms.focusedIndex1 = min(ms.focusedIndex1, max(0, len(ms.objectsGrid)-1))
		ms.focusedIndex2 = min(ms.focusedIndex2, max(0, len(ms.objectsGrid[ms.focusedIndex1])-1))
	}

	ms.refreshContainer()
	ms.callOnFocusMove()
}

func (ms *MultiSplit) SetCurrentPane(o fyne.CanvasObject) {
	if ms.search != "" {
		if searchable, ok := o.(Searcher); ok {
			searchable.Search(ms.search, ms.searchCaseSensitive)
		}
	}

	if len(ms.objectsGrid) == 0 {
		ms.objectsGrid = [][]fyne.CanvasObject{
			[]fyne.CanvasObject{o},
		}
	}

	ms.objectsGrid[ms.focusedIndex1][ms.focusedIndex2] = o

	ms.refreshContainer()
	ms.callOnFocusMove()
}

func (ms *MultiSplit) PaneFocusUp() {
	if ms.layout == LayoutRowColumn {
		ms.focusedIndex1 = max(ms.focusedIndex1-1, 0)
		ms.focusedIndex2 = min(ms.focusedIndex2, len(ms.objectsGrid[ms.focusedIndex1])-1)
	} else {
		ms.focusedIndex2 = max(ms.focusedIndex2-1, 0)
	}

	ms.refreshContainer()
	ms.callOnFocusMove()
}

func (ms *MultiSplit) callOnFocusMove() {
	if ms.OnFocusMove != nil {
		if len(ms.objectsGrid) == 0 || len(ms.objectsGrid[ms.focusedIndex1]) == 0 {
			ms.OnFocusMove(nil)
		} else {
			ms.OnFocusMove(ms.objectsGrid[ms.focusedIndex1][ms.focusedIndex2])
		}
	}
}

func (ms *MultiSplit) PaneFocusDown() {
	if ms.layout == LayoutRowColumn {
		ms.focusedIndex1 = min(ms.focusedIndex1+1, len(ms.objectsGrid)-1)
		ms.focusedIndex2 = min(ms.focusedIndex2, len(ms.objectsGrid[ms.focusedIndex1])-1)
	} else {
		ms.focusedIndex2 = min(ms.focusedIndex2+1, len(ms.objectsGrid[ms.focusedIndex1])-1)
	}

	ms.refreshContainer()
	ms.callOnFocusMove()
}

func (ms *MultiSplit) PaneFocusLeft() {
	if ms.layout == LayoutRowColumn {
		ms.focusedIndex2 = max(ms.focusedIndex2-1, 0)
	} else {
		ms.focusedIndex1 = max(ms.focusedIndex1-1, 0)
		ms.focusedIndex2 = min(ms.focusedIndex2, len(ms.objectsGrid[ms.focusedIndex1])-1)
	}

	ms.refreshContainer()
	ms.callOnFocusMove()
}

func (ms *MultiSplit) PaneFocusRight() {
	if ms.layout == LayoutRowColumn {
		ms.focusedIndex2 = min(ms.focusedIndex2+1, len(ms.objectsGrid[ms.focusedIndex1])-1)
	} else {
		ms.focusedIndex1 = min(ms.focusedIndex1+1, len(ms.objectsGrid)-1)
		ms.focusedIndex2 = min(ms.focusedIndex2, len(ms.objectsGrid[ms.focusedIndex1])-1)
	}

	ms.refreshContainer()
	ms.callOnFocusMove()
}

func (ms *MultiSplit) Search(search string, caseSensitive bool) {
	ms.search = search
	ms.searchCaseSensitive = caseSensitive

	for _, group := range ms.objectsGrid {
		for _, o := range group {
			if searchable, ok := o.(Searcher); ok {
				searchable.Search(search, caseSensitive)
			}
		}
	}
}

func (ms *MultiSplit) SearchClear() {
	ms.search = ""
	ms.searchCaseSensitive = false

	for _, group := range ms.objectsGrid {
		for _, o := range group {
			if searchable, ok := o.(Searcher); ok {
				searchable.SearchClear()
			}
		}
	}
}

type Layout int

const (
	LayoutColumnRow Layout = iota
	LayoutRowColumn
)

func (ms *MultiSplit) widgetWithFocusStyle(w fyne.CanvasObject) fyne.CanvasObject {
	// Create a rectangle for the background
	thm := ms.Theme()
	backgroundColor := thm.Color(ColorNameActiveBackground, theme.VariantDark)
	if backgroundColor == nil {
		backgroundColor = color.NRGBA64{
			0xFFFF, 0xFFFF, 0xFFFF, 0x0A00,
		}
	} else {
		r, g, b, a := backgroundColor.RGBA()
		if r == 0 && g == 0 && b == 0 && a == 0 {
			backgroundColor = color.NRGBA64{
				0xFFFF, 0xFFFF, 0xFFFF, 0x0A00,
			}
		}
	}

	background := canvas.NewRectangle(backgroundColor)
	return container.NewStack(background, w)
}

func (ms *MultiSplit) Refresh() {
	ms.refreshContainer()
}
