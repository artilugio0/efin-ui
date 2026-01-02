package main

import (
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

const (
	TableMessageCopyRow = "table_copy_row"
)

type Table struct {
	widget.BaseWidget

	table *widget.Table

	selectedRow    int
	selectedColumn int

	contentsIndex int
	headers       []string
	rows          [][]string

	searchResults      [][]int
	searchResultsIndex int

	searchResultsBox *SearchResultsCountBox

	keyBindings *KeyBindings

	ShowToastMessageFunc func(string)

	OnSubmit func([]string)
}

func NewTable(rows [][]string) *Table {
	var headers []string
	if len(rows) > 0 {
		headers = rows[0]
	}

	var dataRows [][]string
	if len(rows) > 0 {
		dataRows = rows[1:]
	}

	t := &Table{
		rows:    dataRows,
		headers: headers,
	}
	t.table = widget.NewTable(t.length, t.create, t.update)

	t.table.ShowHeaderRow = true
	t.table.CreateHeader = t.createHeader
	t.table.UpdateHeader = t.updateHeader

	t.ExtendBaseWidget(t)
	t.table.Select(widget.TableCellID{0, 0})

	for i := range len(t.headers) {
		t.table.SetColumnWidth(i, 200)
	}

	t.searchResultsBox = NewSearchResultsCountBox()

	return t
}

func (t *Table) length() (int, int) {
	if len(t.rows) == 0 {
		return 0, 0
	}

	return len(t.rows), len(t.rows[0])
}

func (t *Table) create() fyne.CanvasObject {
	l := widget.NewLabel("")
	l.Wrapping = fyne.TextTruncate
	return l
}

func (t *Table) update(i widget.TableCellID, o fyne.CanvasObject) {
	o.(*widget.Label).SetText(t.rows[i.Row][i.Col])
}

func (t *Table) createHeader() fyne.CanvasObject {
	return widget.NewLabel("")
}

func (t *Table) updateHeader(i widget.TableCellID, o fyne.CanvasObject) {
	if i.Row != -1 || len(t.headers) == 0 {
		return
	}

	o.(*widget.Label).SetText(t.headers[i.Col])
}

func (t *Table) FocusGained() {
	t.table.OnSelected = func(i widget.TableCellID) {
		t.selectedRow = i.Row
		t.selectedColumn = i.Col
	}

	t.table.FocusGained()
}

func (t *Table) FocusLost() {
	t.table.OnSelected = nil
	t.table.FocusLost()
}

func (t *Table) Search(search string, caseSensitive bool) {
	t.searchResultsIndex = 0
	t.searchResults = [][]int{}

	for i, r := range t.rows {
		for j, col := range r {
			if caseSensitive && strings.Contains(col, search) ||
				!caseSensitive && strings.Contains(strings.ToLower(col), strings.ToLower(search)) {

				t.searchResults = append(t.searchResults, []int{i, j})
			}
		}
	}

	if len(t.searchResults) > 0 {
		t.selectedRow = t.searchResults[0][0]
		t.selectedColumn = t.searchResults[0][1]

		t.updateSelectedRow()
	}

	t.searchResultsBox.ShowResults(1, len(t.searchResults))
}

func (t *Table) SearchClear() {
	t.searchResults = nil
	t.searchResultsIndex = 0

	t.searchResultsBox.Hide()
	t.searchResultsBox.Refresh()
}

func (t *Table) SearchPrev() {
	searchLen := len(t.searchResults)
	if searchLen == 0 {
		return
	}
	t.searchResultsIndex = (t.searchResultsIndex - 1 + searchLen) % searchLen

	t.selectedRow = t.searchResults[t.searchResultsIndex][0]
	t.selectedColumn = t.searchResults[t.searchResultsIndex][1]

	t.searchResultsBox.ShowResults(t.searchResultsIndex+1, searchLen)
	t.updateSelectedRow()
}

func (t *Table) SearchNext() {
	searchLen := len(t.searchResults)
	if searchLen == 0 {
		return
	}

	t.searchResultsIndex = (t.searchResultsIndex + 1) % searchLen

	t.selectedRow = t.searchResults[t.searchResultsIndex][0]
	t.selectedColumn = t.searchResults[t.searchResultsIndex][1]

	t.searchResultsBox.ShowResults(t.searchResultsIndex+1, searchLen)
	t.updateSelectedRow()
}

func (t *Table) CreateRenderer() fyne.WidgetRenderer {
	w := container.NewStack(
		t.table,
		container.NewHBox(layout.NewSpacer(), t.searchResultsBox),
	)

	return widget.NewSimpleRenderer(w)
}

func (t *Table) MoveUp() {
	t.selectedRow = max(0, t.selectedRow-1)
	t.updateSelectedRow()
}

func (t *Table) MoveDown() {
	rows, _ := t.table.Length()
	t.selectedRow = min(rows-1, t.selectedRow+1)

	t.updateSelectedRow()
}

func (t *Table) MoveLeft() {
	t.selectedColumn = max(0, t.selectedColumn-1)

	t.updateSelectedRow()
}

func (t *Table) MoveRight() {
	_, cols := t.table.Length()
	t.selectedColumn = min(cols-1, t.selectedColumn+1)

	t.updateSelectedRow()
}

func (t *Table) updateSelectedRow() {
	tcid := widget.TableCellID{
		Row: t.selectedRow,
		Col: t.selectedColumn,
	}
	t.table.ScrollTo(tcid)
	t.table.Select(tcid)
	t.table.Refresh()
}

func (t *Table) Submit() {
	if t.OnSubmit != nil {
		t.OnSubmit(t.rows[t.selectedRow])
	}
}

func (t *Table) SetKeyBindings(kbs *KeyBindings) {
	t.keyBindings = kbs
}

func (t *Table) WidgetName() string {
	return "table"
}

func (t *Table) TypedKey(ev *fyne.KeyEvent) {
	if ok := t.keyBindings.OnTypedKey(ev); ok {
		return
	}

	t.table.TypedKey(ev)
}

func (t *Table) TypedRune(rune) {
}

func (t *Table) TypedShortcut(sc fyne.Shortcut) {
	if ok := t.keyBindings.OnTypedShortcut(sc); ok {
		return
	}
}

func (t *Table) MessageHandle(m Message) {
	messageStr, ok := m.(string)
	if !ok || len(t.rows) == 0 {
		return
	}

	switch messageStr {
	case TableMessageCopyRow:
		rowStr := strings.Join(t.rows[t.selectedRow], "\t")
		err := copyToClipboard(rowStr)
		if err != nil {
			log.Printf("could not copy row to clipboard: %v", err)
			return
		}

		if t.ShowToastMessageFunc != nil {
			t.ShowToastMessageFunc("Row copied to clipboard")
		}
	}
}
