package main

import (
	"fmt"
	"net/http"
	"strings"
	"unicode"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// RequestResponseViewer displays a Request and Response side-by-side in a clean, read-only view
type RequestResponseViewer struct {
	widget.BaseWidget

	request  *Request
	response *Response

	leftLines  []string
	rightLines []string

	leftSearchResults       []int
	leftSearchResultsIndex  int
	rightSearchResults      []int
	rightSearchResultsIndex int

	leftList  *widget.List
	rightList *widget.List

	leftSearchResultsBox  *SearchResultsCountBox
	rightSearchResultsBox *SearchResultsCountBox

	reqLabel  *widget.Label
	respLabel *widget.Label

	rightSelected bool
}

// NewRequestResponseViewer creates a new viewer widget
func NewRequestResponseViewer() *RequestResponseViewer {
	v := &RequestResponseViewer{}

	v.leftLines = []string{}
	v.rightLines = []string{}

	v.leftList = widget.NewList(
		func() int { return len(v.leftLines) },
		func() fyne.CanvasObject {
			return widget.NewLabel("") // Monospace will be applied via style
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			label.SetText(v.leftLines[id])
			label.TextStyle = fyne.TextStyle{Monospace: true}
			label.Wrapping = fyne.TextWrapOff
			label.Selectable = true
		},
	)
	v.leftList.HideSeparators = true

	v.rightList = widget.NewList(
		func() int { return len(v.rightLines) },
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			label.SetText(v.rightLines[id])
			label.TextStyle = fyne.TextStyle{Monospace: true}
			label.Wrapping = fyne.TextWrapOff
			label.Selectable = true
		},
	)
	v.rightList.HideSeparators = true

	v.leftSearchResultsBox = NewSearchResultsCountBox()
	v.rightSearchResultsBox = NewSearchResultsCountBox()

	v.reqLabel = widget.NewLabel("Request")
	v.respLabel = widget.NewLabel("Response")

	v.ExtendBaseWidget(v)
	return v
}

// SetData updates the displayed request and response
func (v *RequestResponseViewer) SetData(req *Request, resp *Response) {
	v.request = req
	v.response = resp
	v.updateContent()
	v.Refresh()
}

func (v *RequestResponseViewer) updateContent() {
	maxLineLength := 90

	if v.request == nil {
		v.leftLines = []string{"No request selected"}
		v.rightLines = []string{}
	} else {
		// Format Request
		var reqLines []string
		reqLines = append(reqLines, fmt.Sprintf("%s %s HTTP/1.1", v.request.Method, v.request.URL))
		reqLines = append(reqLines, fmt.Sprintf("Host: %s", v.request.Host))

		for _, h := range v.request.Headers {
			reqLines = append(reqLines, fmt.Sprintf("%s: %s", h.Name, h.Value))
		}

		if v.request.Body != "" {
			reqLines = append(reqLines, "", replaceUnprintableHex(v.request.Body))
		} else {
			reqLines = append(reqLines, "")
		}

		reqLines = splitLongStrings(reqLines, maxLineLength)
		v.leftLines = reqLines
	}

	if v.response == nil {
		v.rightLines = []string{"No response received"}
	} else {
		// Format Response
		statusReason := http.StatusText(v.response.StatusCode)
		var respLines []string
		respLines = append(respLines, fmt.Sprintf("HTTP/1.1 %d %s", v.response.StatusCode, statusReason))

		for _, h := range v.response.Headers {
			respLines = append(respLines, fmt.Sprintf("%s: %s", h.Name, h.Value))
		}

		if v.response.Body != "" {
			respLines = append(respLines, "", replaceUnprintableHex(v.response.Body))
		} else {
			respLines = append(respLines, "")
		}

		respLines = splitLongStrings(respLines, maxLineLength)
		v.rightLines = respLines
	}

	// Trigger list refresh
	v.leftList.Refresh()
	v.rightList.Refresh()
}

// CreateRenderer implements the Fyne renderer
func (v *RequestResponseViewer) CreateRenderer() fyne.WidgetRenderer {
	// Labels with icons for better visual appeal
	v.reqLabel.Alignment = fyne.TextAlignCenter
	v.reqLabel.TextStyle = fyne.TextStyle{Bold: true}
	reqHeader := container.NewBorder(nil, nil, nil, nil, v.reqLabel)

	v.respLabel.Alignment = fyne.TextAlignCenter
	v.respLabel.TextStyle = fyne.TextStyle{Bold: false}
	respHeader := container.NewBorder(nil, nil, nil, nil, v.respLabel)

	// Scrollable lists with padding
	leftScroll := container.NewVScroll(container.NewPadded(v.leftList))
	rightScroll := container.NewVScroll(container.NewPadded(v.rightList))

	// Add headers on top
	var left fyne.CanvasObject = container.NewBorder(reqHeader, nil, nil, nil, leftScroll)
	var right fyne.CanvasObject = container.NewBorder(respHeader, nil, nil, nil, rightScroll)

	left = container.NewStack(
		left,
		container.NewHBox(layout.NewSpacer(), v.leftSearchResultsBox),
	)

	right = container.NewStack(
		right,
		container.NewHBox(layout.NewSpacer(), v.rightSearchResultsBox),
	)

	// Horizontal split
	split := container.NewHSplit(left, right)
	split.Offset = 0.5

	return widget.NewSimpleRenderer(split)
}

func replaceUnprintableHex(s string) string {
	var builder strings.Builder
	for _, r := range s {
		if unicode.IsPrint(r) {
			builder.WriteRune(r)
		} else {
			if r <= 0xFF {
				builder.WriteString(fmt.Sprintf("<0x%02X>", r))
			} else {
				builder.WriteString(fmt.Sprintf("<0x%04X>", r))
			}
		}
	}
	return builder.String()
}

func splitLongStrings(lines []string, maxLen int) []string {
	if maxLen <= 0 {
		panic("maxLen must be positive")
	}

	var result []string

	for _, line := range lines {
		if len(line) <= maxLen {
			result = append(result, line)
			continue
		}

		for i := 0; i < len(line); i += maxLen {
			end := i + maxLen
			if end > len(line) {
				end = len(line)
			}
			result = append(result, line[i:end])
		}
	}

	return result
}

func (v *RequestResponseViewer) Search(search string, caseSensitive bool) {
	v.leftSearchResultsIndex = 0
	v.rightSearchResultsIndex = 0
	v.leftSearchResults = []int{}
	v.rightSearchResults = []int{}
	for i, l := range v.leftLines {
		if caseSensitive && strings.Contains(l, search) ||
			!caseSensitive && strings.Contains(strings.ToLower(l), strings.ToLower(search)) {

			v.leftSearchResults = append(v.leftSearchResults, i)
		}
	}

	for i, l := range v.rightLines {
		if caseSensitive && strings.Contains(l, search) ||
			!caseSensitive && strings.Contains(strings.ToLower(l), strings.ToLower(search)) {

			v.rightSearchResults = append(v.rightSearchResults, i)
		}
	}

	v.updateSearchResultsBox()
}

func (v *RequestResponseViewer) SearchClear() {
	v.leftSearchResultsIndex = 0
	v.rightSearchResultsIndex = 0
	v.leftSearchResults = nil
	v.rightSearchResults = nil

	fyne.Do(func() {
		v.rightSearchResultsBox.Hide()
		v.rightSearchResultsBox.Refresh()

		v.leftSearchResultsBox.Hide()
		v.leftSearchResultsBox.Refresh()
	})
}

func (v *RequestResponseViewer) SearchPrev() {
	if v.rightSelected {
		lenRight := len(v.rightSearchResults)
		if lenRight == 0 {
			return
		}

		v.rightSearchResultsIndex = (v.rightSearchResultsIndex - 1 + lenRight) % lenRight
	} else {
		lenLeft := len(v.leftSearchResults)
		if lenLeft == 0 {
			return
		}

		v.leftSearchResultsIndex = (v.leftSearchResultsIndex - 1 + lenLeft) % lenLeft
	}

	v.updateSearchResultsBox()
}

func (v *RequestResponseViewer) SearchNext() {
	if v.rightSelected {
		lenRight := len(v.rightSearchResults)
		if lenRight == 0 {
			return
		}

		v.rightSearchResultsIndex = (v.rightSearchResultsIndex + 1) % lenRight
	} else {
		lenLeft := len(v.leftSearchResults)
		if lenLeft == 0 {
			return
		}

		v.leftSearchResultsIndex = (v.leftSearchResultsIndex + 1) % lenLeft
	}

	v.updateSearchResultsBox()
}

/*
	func (v *RequestResponseViewer) Update(ev any) {
		switch ev := ev.(type) {
		case EventSearchResult:

			if v.rightSelected {
				lenRight := len(v.rightSearchResults)
				if lenRight == 0 {
					break
				}

				switch ev {
				case EventSearchResultNext:
					v.rightSearchResultsIndex = (v.rightSearchResultsIndex + 1) % lenRight

				case EventSearchResultPrev:
					v.rightSearchResultsIndex = (v.rightSearchResultsIndex - 1 + lenRight) % lenRight
				}
			} else {
				lenLeft := len(v.leftSearchResults)
				if lenLeft == 0 {
					break
				}

				switch ev {
				case EventSearchResultNext:
					v.leftSearchResultsIndex = (v.leftSearchResultsIndex + 1) % lenLeft

				case EventSearchResultPrev:
					v.leftSearchResultsIndex = (v.leftSearchResultsIndex - 1 + lenLeft) % lenLeft
				}
			}

			v.updateSearchResultsBox()

		case EventMovement:
			if ev == EventMovementLeft && v.rightSelected {
				fyne.Do(func() {
					v.rightSelected = false
					v.reqLabel.TextStyle = fyne.TextStyle{Bold: true}
					v.respLabel.TextStyle = fyne.TextStyle{Bold: false}
					v.reqLabel.Refresh()
					v.respLabel.Refresh()
				})
			}

			if ev == EventMovementRight && !v.rightSelected {
				fyne.Do(func() {
					v.rightSelected = true
					v.reqLabel.TextStyle = fyne.TextStyle{Bold: false}
					v.respLabel.TextStyle = fyne.TextStyle{Bold: true}
					v.reqLabel.Refresh()
					v.respLabel.Refresh()
				})
			}
		}
	}
*/
func (v *RequestResponseViewer) MoveUp() {
}

func (v *RequestResponseViewer) MoveDown() {
}

func (v *RequestResponseViewer) MoveLeft() {
	v.rightSelected = false
	v.reqLabel.TextStyle = fyne.TextStyle{Bold: true}
	v.respLabel.TextStyle = fyne.TextStyle{Bold: false}
	v.reqLabel.Refresh()
	v.respLabel.Refresh()
}

func (v *RequestResponseViewer) MoveRight() {
	v.rightSelected = true
	v.reqLabel.TextStyle = fyne.TextStyle{Bold: false}
	v.respLabel.TextStyle = fyne.TextStyle{Bold: true}
	v.reqLabel.Refresh()
	v.respLabel.Refresh()
}

func (v *RequestResponseViewer) updateSearchResultsBox() {
	if len(v.leftSearchResults) > 0 {
		v.leftList.ScrollTo(v.leftSearchResults[v.leftSearchResultsIndex])
		v.leftList.Select(v.leftSearchResults[v.leftSearchResultsIndex])
		v.leftList.Refresh()
	}
	v.leftSearchResultsBox.ShowResults(v.leftSearchResultsIndex+1, len(v.leftSearchResults))

	if len(v.rightSearchResults) > 0 {
		v.rightList.ScrollTo(v.rightSearchResults[v.rightSearchResultsIndex])
		v.rightList.Select(v.rightSearchResults[v.rightSearchResultsIndex])
	}
	v.rightSearchResultsBox.ShowResults(v.rightSearchResultsIndex+1, len(v.rightSearchResults))
}
