package efinui

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"text/template"
	"unicode"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/artilugio0/efin-ui/templates"
)

const (
	RequestResponseViewerMessageCopyRequestScript = "request_response_viewer_copy_request_script"
	RequestResponseViewerMessageCopyRequest       = "request_response_viewer_copy_request"
	RequestResponseViewerMessageCopyResponse      = "request_response_viewer_copy_response"
)

// RequestResponseViewer displays a Request and Response side-by-side in a clean, read-only view
type RequestResponseViewer struct {
	widget.BaseWidget

	request  *Request
	response *Response

	leftLinesList  *LinesList
	rightLinesList *LinesList

	keyBindings *KeyBindings

	reqLabel  *widget.Label
	respLabel *widget.Label

	rightSelected bool

	ShowToastMessageFunc func(string)
}

// NewRequestResponseViewer creates a new viewer widget
func NewRequestResponseViewer(req *Request, resp *Response) *RequestResponseViewer {
	v := &RequestResponseViewer{
		request:  req,
		response: resp,
	}

	v.leftLinesList = NewLinesList(string(req.Raw()))
	v.rightLinesList = NewLinesList(string(resp.Raw()))

	v.reqLabel = widget.NewLabel("Request")
	v.respLabel = widget.NewLabel("Response")

	v.ExtendBaseWidget(v)
	return v
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
	leftScroll := container.NewVScroll(container.NewPadded(v.leftLinesList))
	rightScroll := container.NewVScroll(container.NewPadded(v.rightLinesList))

	// Add headers on top
	var left fyne.CanvasObject = container.NewBorder(reqHeader, nil, nil, nil, leftScroll)
	var right fyne.CanvasObject = container.NewBorder(respHeader, nil, nil, nil, rightScroll)

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

func (v *RequestResponseViewer) Search(search string, caseSensitive bool) {
	v.rightLinesList.Search(search, caseSensitive)
	v.leftLinesList.Search(search, caseSensitive)
}

func (v *RequestResponseViewer) SearchClear() {
	v.rightLinesList.SearchClear()
	v.leftLinesList.SearchClear()
}

func (v *RequestResponseViewer) SearchPrev() {
	if v.rightSelected {
		v.rightLinesList.SearchPrev()
		return
	}

	v.leftLinesList.SearchPrev()
	return
}

func (v *RequestResponseViewer) SearchNext() {
	if v.rightSelected {
		v.rightLinesList.SearchNext()
		return
	}

	v.leftLinesList.SearchNext()
	return
}

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

func (v *RequestResponseViewer) SetKeyBindings(kbs *KeyBindings) {
	v.keyBindings = kbs
}

func (v *RequestResponseViewer) WidgetName() string {
	return "request_response_viewer"
}

func (v *RequestResponseViewer) TypedKey(ev *fyne.KeyEvent) {
	if ok := v.keyBindings.OnTypedKey(ev); ok {
		return
	}
}

func (v *RequestResponseViewer) TypedRune(rune) {
}

func (v *RequestResponseViewer) TypedShortcut(sc fyne.Shortcut) {
	if ok := v.keyBindings.OnTypedShortcut(sc); ok {
		return
	}
}

func (v *RequestResponseViewer) MessageHandle(m Message) {
	messageStr, ok := m.(string)
	if !ok {
		return
	}

	switch messageStr {
	case RequestResponseViewerMessageCopyRequestScript:
		funcs := map[string]any{
			"contains": strings.Contains,
			"contains_bytes": func(s []byte, c string) bool {
				return bytes.Contains(s, []byte(c))
			},
		}

		scriptTpl := templates.GetRequestTestifierScript()
		t, err := template.New("make_request").Funcs(funcs).Parse(scriptTpl)
		if err != nil {
			log.Printf("could not copy request script: %v", err)
			return
		}

		f := &strings.Builder{}

		if err := t.Execute(f, v.request); err != nil {
			log.Printf("could execute request script template: %v", err)
			return
		}

		if err := copyToClipboard(f.String()); err != nil {
			log.Printf("could not copy request script to clipboard: %v", err)
			return
		}

		if v.ShowToastMessageFunc != nil {
			v.ShowToastMessageFunc("Request script copied to clipboard")
		}

	case RequestResponseViewerMessageCopyRequest:
		reqBytes := v.request.Raw()
		copyToClipboard(string(reqBytes))

		if v.ShowToastMessageFunc != nil {
			v.ShowToastMessageFunc("Request copied to clipboard")
		}

	case RequestResponseViewerMessageCopyResponse:
		respBytes := v.response.Raw()
		copyToClipboard(string(respBytes))

		if v.ShowToastMessageFunc != nil {
			v.ShowToastMessageFunc("Response copied to clipboard")
		}
	}
}

func (v *RequestResponseViewer) FocusGained() {}

func (v *RequestResponseViewer) FocusLost() {}
