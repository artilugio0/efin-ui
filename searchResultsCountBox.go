package efinui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type SearchResultsCountBox struct {
	widget.BaseWidget

	label *widget.Label
}

func NewSearchResultsCountBox() *SearchResultsCountBox {
	s := &SearchResultsCountBox{}
	s.ExtendBaseWidget(s)
	s.label = widget.NewLabel("")

	s.Hide()

	return s
}

func (s *SearchResultsCountBox) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(s.label)
}

func (s *SearchResultsCountBox) ShowResults(current, total int) {
	if total > 0 {
		s.label.SetText(fmt.Sprintf("Search: %d of %d", current, total))
	} else {
		s.label.SetText("No search results")
	}

	s.Show()
	s.Refresh()
}
