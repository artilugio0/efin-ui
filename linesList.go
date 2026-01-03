package main

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type LinesList struct {
	widget.BaseWidget

	list *widget.List

	charWidth float32

	originalLines []string
	viewLines     []string

	search              string
	searchCaseSensitive bool
	searchResults       []int
	searchResultsIndex  int

	container        *fyne.Container
	searchResultsBox *SearchResultsCountBox
}

func NewLinesList(text string) *LinesList {
	ll := &LinesList{
		originalLines: strings.Split(text, "\n"),
		charWidth:     0,

		searchResultsBox: NewSearchResultsCountBox(),
	}

	var list *widget.List
	list = widget.NewList(
		func() int { return len(ll.viewLines) },
		func() fyne.CanvasObject {
			return canvas.NewText("", list.Theme().Color(theme.ColorNameForeground, theme.VariantDark))
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			txt := obj.(*canvas.Text)
			txt.Text = ll.viewLines[id]
			txt.Color = list.Theme().Color(theme.ColorNameForeground, theme.VariantDark)
			txt.Refresh()
		},
	)
	list.HideSeparators = true
	ll.list = list

	ll.ExtendBaseWidget(ll)

	ll.container = container.NewBorder(
		ll.searchResultsBox,
		nil,
		nil,
		nil,
		ll.list,
	)

	return ll
}

func (ll *LinesList) refreshWrappedContent() {
	// Get current list width (it should be laid out at this point)
	width := ll.list.Size().Width

	// Small safety margin (scrollbar, padding, etc.)
	const margin = 20.0
	width -= margin

	if width <= 0 {
		return // not yet laid out
	}

	// Calculate char width only once (or on theme change)
	if ll.charWidth <= 0 {
		// Use "W" or "M" - widest common chars in monospace
		size := fyne.MeasureText("W", theme.TextSize(), fyne.TextStyle{Monospace: true})
		ll.charWidth = size.Width
	}

	maxChars := int(width / ll.charWidth)

	// Safety - don't allow ridiculous values
	if maxChars < 20 {
		maxChars = 90
	}

	ll.viewLines = wrapLines(ll.originalLines, maxChars)

	if ll.search != "" {
		ll.Search(ll.search, ll.searchCaseSensitive)
	}

	ll.list.Refresh()
}

func wrapLines(original []string, maxChars int) []string {
	var result []string
	for _, line := range original {
		if len(line) <= maxChars {
			result = append(result, line)
			continue
		}
		// Hard split (you can improve with word-aware splitting if wanted)
		for i := 0; i < len(line); i += maxChars {
			end := i + maxChars
			if end > len(line) {
				end = len(line)
			}
			result = append(result, line[i:end])
		}
	}
	return result
}

func (ll *LinesList) Resize(s fyne.Size) {
	ll.list.Resize(s)
	ll.refreshWrappedContent()
}

func (ll *LinesList) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(ll.container)
}

func (ll *LinesList) Search(search string, caseSensitive bool) {
	ll.search = search
	ll.searchCaseSensitive = caseSensitive

	ll.searchResultsIndex = 0
	ll.searchResults = []int{}
	for i, l := range ll.viewLines {
		if caseSensitive && strings.Contains(l, search) ||
			!caseSensitive && strings.Contains(strings.ToLower(l), strings.ToLower(search)) {

			ll.searchResults = append(ll.searchResults, i)
		}
	}

	ll.updateSearchResultsBox()
}

func (ll *LinesList) SearchClear() {
	ll.search = ""

	ll.searchResultsIndex = 0
	ll.searchResults = nil

	ll.searchResultsBox.Hide()
	ll.searchResultsBox.Refresh()
	ll.container.Refresh()
}

func (ll *LinesList) SearchPrev() {
	lenRes := len(ll.searchResults)
	if lenRes == 0 {
		return
	}

	ll.searchResultsIndex = (ll.searchResultsIndex - 1 + lenRes) % lenRes
	ll.updateSearchResultsBox()
}

func (ll *LinesList) SearchNext() {
	lenRes := len(ll.searchResults)
	if lenRes == 0 {
		return
	}

	ll.searchResultsIndex = (ll.searchResultsIndex + 1) % lenRes
	ll.updateSearchResultsBox()
}

func (ll *LinesList) updateSearchResultsBox() {
	if len(ll.searchResults) > 0 {
		ll.list.ScrollTo(ll.searchResults[ll.searchResultsIndex])
		ll.list.Select(ll.searchResults[ll.searchResultsIndex])
		ll.list.Refresh()
	}

	ll.searchResultsBox.ShowResults(ll.searchResultsIndex+1, len(ll.searchResults))
}
