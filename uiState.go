package main

import (
	"log"
	"slices"
)

type UIState struct {
	CurrentTab       int                 `json:"current_tab"`
	Tabs             []*Pane             `json:"tabs"`
	FocusedPane      []int               `json:"focused_pane"`
	KeyBindings      map[string][]string `json:"key_bindings"`
	lastContentIndex int                 `json:"-"`
	lastPaneID       int                 `json:"-"`
}

type Pane struct {
	Layout  string  `json:"layout"`
	Panes   []*Pane `json:"panes"`
	Content int     `json:"content"`
	ID      int     `json:"id"`
}

func NewUIState() *UIState {
	return &UIState{
		FocusedPane: []int{0},
		CurrentTab:  0,
		Tabs: []*Pane{
			{
				Layout: "vsplit",
				Panes: []*Pane{
					{
						Layout:  "single",
						Content: 0,
					},
				},
			},
		},
	}
}

func (uis *UIState) IncreaseLastContentIndex() {
	uis.lastContentIndex++
}

func (uis *UIState) PaneCreate() {
	uis.FocusedPane = uis.paneCreate(uis.Tabs[uis.CurrentTab], uis.FocusedPane)
}

func (uis *UIState) paneCreate(pane *Pane, focusedPane []int) []int {
	if len(focusedPane) == 1 {
		uis.lastPaneID++
		pane.Panes = append(pane.Panes, &Pane{
			Layout:  "single",
			Content: 0,
			ID:      uis.lastPaneID,
		})
		return []int{len(pane.Panes) - 1}
	}

	restFocusedPane := uis.paneCreate(pane.Panes[focusedPane[0]], focusedPane[1:])
	newFocusedPane := []int{focusedPane[0]}
	return append(newFocusedPane, restFocusedPane...)
}

func (uis *UIState) PaneDelete() {
	uis.FocusedPane = uis.paneDelete(uis.Tabs[uis.CurrentTab], uis.FocusedPane)
	if len(uis.FocusedPane) == 0 {
		uis.Tabs = slices.Delete(uis.Tabs, uis.CurrentTab, uis.CurrentTab+1)
		uis.CurrentTab = max(0, uis.CurrentTab-1)

		if len(uis.Tabs) > 0 {
			uis.FocusedPane = []int{0}
			p := uis.Tabs[uis.CurrentTab].Panes[0]

			for len(p.Panes) > 0 {
				uis.FocusedPane = append(uis.FocusedPane, 0)
				p = p.Panes[0]
			}
		}
	}

	if len(uis.Tabs) == 0 {
		uis.Tabs = append(uis.Tabs, &Pane{
			Layout: "vsplit",
			Panes: []*Pane{
				{
					Layout:  "single",
					Content: 0,
				},
			},
		})
		uis.FocusedPane = []int{0}
	}
}

func (uis *UIState) paneDelete(pane *Pane, focusedPane []int) []int {
	if len(focusedPane) == 0 {
		return []int{}
	}

	if len(focusedPane) == 1 {
		pane.Panes = slices.Delete(pane.Panes, focusedPane[0], focusedPane[0]+1)
		if len(pane.Panes) == 0 {
			return []int{}
		}
		return []int{max(0, focusedPane[0]-1)}
	}

	restFocusedPane := uis.paneDelete(pane.Panes[focusedPane[0]], focusedPane[1:])
	if len(restFocusedPane) == 0 && pane.Panes[focusedPane[0]].Layout != "single" {
		pane.Panes = slices.Delete(pane.Panes, focusedPane[0], focusedPane[0]+1)
		if len(pane.Panes) == 0 {
			return []int{}
		}

		if len(pane.Panes) == 1 {
			pane.Layout = "single"
			pane.Content = pane.Panes[0].Content
			pane.Panes = nil

			return []int{}
		}

		return []int{max(0, focusedPane[0]-1)}
	}

	newFocusedPane := []int{focusedPane[0]}
	return append(newFocusedPane, restFocusedPane...)
}

func (uis *UIState) PaneFocusNext() {
	log.Printf("focused pane: %+v", uis.FocusedPane)
	newFocusedPane := uis.paneFocusNext(uis.Tabs[uis.CurrentTab], uis.FocusedPane)
	log.Printf("new focused pane: %+v", newFocusedPane)
	if newFocusedPane[0] != -1 {
		uis.FocusedPane = newFocusedPane
	}
}

func (uis *UIState) paneFocusNext(pane *Pane, focusedPane []int) []int {
	if len(focusedPane) == 1 {
		if focusedPane[0]+1 == len(pane.Panes) {
			return []int{-1}
		}

		p := pane.Panes[focusedPane[0]+1]
		newFocusedPane := []int{focusedPane[0] + 1}

		for len(p.Panes) > 0 {
			newFocusedPane = append(newFocusedPane, 0)
			p = p.Panes[0]
		}

		return newFocusedPane
	}

	restFocusedPane := uis.paneFocusNext(pane.Panes[focusedPane[0]], focusedPane[1:])
	if restFocusedPane[0] == -1 {
		if focusedPane[0] == len(pane.Panes)-1 {
			return []int{-1}
		}

		p := pane.Panes[focusedPane[0]+1]
		newFocusedPane := []int{focusedPane[0] + 1}
		for len(p.Panes) > 0 {
			newFocusedPane = append(newFocusedPane, 0)
			p = p.Panes[0]
		}

		return newFocusedPane
	}

	newFocusedPane := []int{focusedPane[0]}
	return append(newFocusedPane, restFocusedPane...)
}

func (uis *UIState) PaneFocusPrev() {
	newFocusedPane := uis.paneFocusPrev(uis.Tabs[uis.CurrentTab], uis.FocusedPane)
	if newFocusedPane[0] != -1 {
		uis.FocusedPane = newFocusedPane
	}
}

func (uis *UIState) paneFocusPrev(pane *Pane, focusedPane []int) []int {
	if len(focusedPane) == 1 {
		if focusedPane[0] == 0 {
			return []int{-1}
		}

		newFocusedPane := []int{focusedPane[0] - 1}
		p := pane.Panes[focusedPane[0]-1]
		for len(p.Panes) > 0 {
			newFocusedPane = append(newFocusedPane, len(p.Panes)-1)
			p = p.Panes[len(p.Panes)-1]
		}

		return newFocusedPane
	}

	restFocusedPane := uis.paneFocusPrev(pane.Panes[focusedPane[0]], focusedPane[1:])
	if restFocusedPane[0] == -1 {
		if focusedPane[0] == 0 {
			return []int{-1}
		}

		newFocusedPane := []int{focusedPane[0] - 1}
		p := pane.Panes[focusedPane[0]-1]
		for len(p.Panes) > 0 {
			newFocusedPane = append(newFocusedPane, len(p.Panes)-1)
			p = p.Panes[len(p.Panes)-1]
		}

		return newFocusedPane
	}

	newFocusedPane := []int{focusedPane[0]}
	return append(newFocusedPane, restFocusedPane...)
}

func (uis *UIState) FocusedPaneSetContent(newContent int) {
	uis.focusedPaneSetContent(uis.Tabs[uis.CurrentTab], uis.FocusedPane, newContent)
}

func (uis *UIState) FocusedPaneSetContentToLast() {
	uis.focusedPaneSetContent(uis.Tabs[uis.CurrentTab], uis.FocusedPane, uis.lastContentIndex)
}

func (uis *UIState) focusedPaneSetContent(pane *Pane, focusedPane []int, newContent int) {
	if len(focusedPane) == 1 {
		pane.Panes[focusedPane[0]].Content = newContent
		return
	}

	uis.focusedPaneSetContent(pane.Panes[focusedPane[0]], focusedPane[1:], newContent)
}

func (uis *UIState) LayoutVSplit() {
	uis.layoutSet(uis.Tabs[uis.CurrentTab], uis.FocusedPane, "vsplit")
}

func (uis *UIState) LayoutHSplit() {
	uis.layoutSet(uis.Tabs[uis.CurrentTab], uis.FocusedPane, "hsplit")
}

func (uis *UIState) layoutSet(pane *Pane, focusedPane []int, layout string) {
	if len(focusedPane) == 1 {
		pane.Layout = layout
		return
	}

	uis.layoutSet(pane.Panes[focusedPane[0]], focusedPane[1:], layout)
}

func (uis *UIState) PaneVSplit() {
	uis.FocusedPane = uis.paneSplit(uis.Tabs[uis.CurrentTab], uis.FocusedPane, "vsplit")
}

func (uis *UIState) PaneHSplit() {
	uis.FocusedPane = uis.paneSplit(uis.Tabs[uis.CurrentTab], uis.FocusedPane, "hsplit")
}

func (uis *UIState) paneSplit(pane *Pane, focusedPane []int, layout string) []int {
	if len(focusedPane) == 0 {
		if pane.Layout != "single" {
			return []int{}
		}

		pane.Layout = layout
		pane.Panes = append(
			[]*Pane{},
			&Pane{
				Layout:  "single",
				Content: pane.Content,
				ID:      uis.lastPaneID + 1,
			},
			&Pane{
				Layout:  "single",
				Content: 0,
				ID:      uis.lastPaneID + 2,
			},
		)

		pane.Content = 0
		uis.lastPaneID += 2

		return []int{1}
	}

	restFocusedPane := uis.paneSplit(pane.Panes[focusedPane[0]], focusedPane[1:], layout)
	newFocusedPane := []int{focusedPane[0]}
	return append(newFocusedPane, restFocusedPane...)
}

func (uis *UIState) TabCreate() {
	uis.Tabs = append(uis.Tabs, &Pane{
		Layout: "vsplit",
		Panes: []*Pane{
			{
				Layout:  "single",
				Content: 0,
			},
		},
	})

	uis.CurrentTab = len(uis.Tabs) - 1
	uis.FocusedPane = []int{0}
}

func (uis *UIState) TabFocusNext() {
	uis.CurrentTab = (uis.CurrentTab + 1) % len(uis.Tabs)

	uis.FocusedPane = []int{0}
	p := uis.Tabs[uis.CurrentTab].Panes[0]
	for len(p.Panes) > 0 {
		p = p.Panes[0]
		uis.FocusedPane = append(uis.FocusedPane, 0)
	}
}

func (uis *UIState) TabFocusPrev() {
	uis.CurrentTab = (uis.CurrentTab + len(uis.Tabs) - 1) % len(uis.Tabs)

	uis.FocusedPane = []int{0}
	p := uis.Tabs[uis.CurrentTab].Panes[0]
	for len(p.Panes) > 0 {
		p = p.Panes[0]
		uis.FocusedPane = append(uis.FocusedPane, 0)
	}
}
