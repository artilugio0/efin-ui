package main

type UIState struct {
	CurrentTab       int                 `json:"current_tab"`
	Tabs             []*Pane             `json:"tabs"`
	FocusedPane      []int               `json:"focused_pane"`
	KeyBindings      map[string][]string `json:"key_bindings"`
	lastContentIndex int                 `json:"-"`
}

type Pane struct {
	Layout  string  `json:"layout"`
	Panes   []*Pane `json:"panes"`
	Content int     `json:"content"`
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
		pane.Panes = append(pane.Panes, &Pane{
			Layout:  "single",
			Content: uis.lastContentIndex,
		})
		return []int{len(pane.Panes) - 1}
	}

	restFocusedPane := uis.paneCreate(pane.Panes[focusedPane[0]], focusedPane[1:])
	newFocusedPane := []int{focusedPane[0]}
	return append(newFocusedPane, restFocusedPane...)
}

func (uis *UIState) PaneDelete() {
	if len(uis.Tabs[uis.CurrentTab].Panes) <= 1 {
		return
	}

	uis.FocusedPane = uis.paneDelete(uis.Tabs[uis.CurrentTab], uis.FocusedPane)
}

func (uis *UIState) paneDelete(pane *Pane, focusedPane []int) []int {
	if len(focusedPane) == 1 {
		newPanes := []*Pane{}
		for i, p := range pane.Panes {
			if i != focusedPane[0] {
				newPanes = append(newPanes, p)
			}
		}
		pane.Panes = newPanes
		if len(pane.Panes) == 0 {
			return []int{}
		}
		return []int{0}
	}

	restFocusedPane := uis.paneDelete(pane.Panes[focusedPane[0]], focusedPane[1:])
	newFocusedPane := []int{focusedPane[0]}
	return append(newFocusedPane, restFocusedPane...)
}

func (uis *UIState) PaneFocusNext() {
	uis.FocusedPane = uis.paneFocusNext(uis.Tabs[uis.CurrentTab], uis.FocusedPane)
}

func (uis *UIState) paneFocusNext(pane *Pane, focusedPane []int) []int {
	if len(focusedPane) == 1 {
		return []int{(focusedPane[0] + 1) % len(pane.Panes)}
	}

	restFocusedPane := uis.paneFocusNext(pane.Panes[focusedPane[0]], focusedPane[1:])
	newFocusedPane := []int{focusedPane[0]}
	return append(newFocusedPane, restFocusedPane...)
}

func (uis *UIState) PaneFocusPrev() {
	uis.FocusedPane = uis.paneFocusPrev(uis.Tabs[uis.CurrentTab], uis.FocusedPane)
}

func (uis *UIState) paneFocusPrev(pane *Pane, focusedPane []int) []int {
	if len(focusedPane) == 1 {
		return []int{(focusedPane[0] + len(pane.Panes) - 1) % len(pane.Panes)}
	}

	restFocusedPane := uis.paneFocusPrev(pane.Panes[focusedPane[0]], focusedPane[1:])
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
}

func (uis *UIState) TabFocusPrev() {
	uis.CurrentTab = (uis.CurrentTab + len(uis.Tabs) - 1) % len(uis.Tabs)
	uis.FocusedPane = []int{0}
}
