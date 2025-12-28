package main

import "fyne.io/fyne/v2"

// FlexGrid lays out children in groups along a primary direction (default: horizontal groups = rows).
// It can optionally switch to vertical groups (columns) via WithColumnMode().
// Supports percentage-based sizing and uniform padding between cells.
type FlexGrid struct {
	// Percentages along primary direction (height in row mode, width in column mode)
	PrimarySizes []float32

	// Per group: percentages along secondary direction
	SecondarySizes [][]float32

	// Number of items in each primary group
	ItemsPerGroup []int

	// Uniform padding between cells
	Padding float32

	columnMode bool
}

// NewFlexGrid creates a flexible grid layout in row mode by default.
// itemsPerGroup defines how many widgets belong to each primary group.
func NewFlexGrid(itemsPerGroup []int) *FlexGrid {
	return &FlexGrid{
		ItemsPerGroup: itemsPerGroup,
	}
}

// WithColumnMode switches the layout to column mode (vertical primary direction).
func (g *FlexGrid) WithColumnMode() *FlexGrid {
	g.columnMode = true
	return g
}

// WithPrimarySizes sets percentage sizes along the primary direction.
// In row mode: heights of each group.
// In column mode: widths of each group.
func (g *FlexGrid) WithPrimarySizes(sizes ...float32) *FlexGrid {
	g.PrimarySizes = sizes
	return g
}

// WithSecondarySizes sets percentage sizes within each group along the secondary direction.
// In row mode: column widths inside each row.
// In column mode: row heights inside each column.
func (g *FlexGrid) WithSecondarySizes(sizesPerGroup ...[]float32) *FlexGrid {
	if g.SecondarySizes == nil || len(g.SecondarySizes) < len(sizesPerGroup) {
		tmp := make([][]float32, len(sizesPerGroup))
		copy(tmp, g.SecondarySizes)
		g.SecondarySizes = tmp
	}
	for i, s := range sizesPerGroup {
		g.SecondarySizes[i] = append(g.SecondarySizes[i][:0], s...)
	}
	return g
}

// WithPadding sets uniform padding (in pixels) between all cells.
func (g *FlexGrid) WithPadding(padding float32) *FlexGrid {
	g.Padding = padding
	return g
}

func (g *FlexGrid) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) == 0 || len(g.ItemsPerGroup) == 0 {
		return
	}

	numGroups := len(g.ItemsPerGroup)
	padding := g.Padding
	isColumnMode := g.columnMode

	totalWidth := float32(size.Width)
	totalHeight := float32(size.Height)

	// Primary direction: horizontal for rows, vertical for columns
	var primaryTotal, secondaryTotal float32
	var primaryAvailable float32
	var primaryPadTotal float32

	if isColumnMode {
		// Primary = horizontal (rows), Secondary = vertical
		primaryTotal = totalWidth
		secondaryTotal = totalHeight
		primaryPadTotal = padding * float32(numGroups-1)
	} else {
		// Primary = vertical (columns), Secondary = horizontal
		primaryTotal = totalHeight
		secondaryTotal = totalWidth
		primaryPadTotal = padding * float32(numGroups-1)
	}
	primaryPadTotal = max(0, primaryPadTotal)
	primaryAvailable = primaryTotal - primaryPadTotal

	// Compute primary group sizes
	groupSizes := make([]float32, numGroups)
	specifiedSum := float32(0)
	if len(g.PrimarySizes) > 0 {
		for i, p := range g.PrimarySizes {
			if i < numGroups {
				groupSizes[i] = primaryAvailable * p / 100
				specifiedSum += p
			}
		}
	}
	remainingPerc := 100 - specifiedSum
	if remainingPerc > 0 && len(g.PrimarySizes) < numGroups {
		equalPerc := remainingPerc / float32(numGroups-len(g.PrimarySizes))
		for i := len(g.PrimarySizes); i < numGroups; i++ {
			groupSizes[i] = primaryAvailable * equalPerc / 100
		}
	} else if specifiedSum == 0 {
		equal := primaryAvailable / float32(numGroups)
		for i := range groupSizes {
			groupSizes[i] = equal
		}
	}

	// Precompute secondary padding per group
	secondaryPadPerGroup := make([]float32, numGroups)
	for i, count := range g.ItemsPerGroup {
		if count > 1 {
			secondaryPadPerGroup[i] = padding * float32(count-1)
		}
	}

	posPrimary := float32(0) // x in row mode, y in column mode
	objIdx := 0

	for groupIdx, itemCount := range g.ItemsPerGroup {
		if objIdx >= len(objects) || groupSizes[groupIdx] <= 0 {
			break
		}
		groupSize := groupSizes[groupIdx]

		secondaryAvailable := secondaryTotal - secondaryPadPerGroup[groupIdx]
		if secondaryAvailable < 0 {
			secondaryAvailable = 0
		}

		// Secondary sizes for this group
		var secPercents []float32
		if g.SecondarySizes != nil && groupIdx < len(g.SecondarySizes) && len(g.SecondarySizes[groupIdx]) > 0 {
			secPercents = g.SecondarySizes[groupIdx]
		}

		itemSizes := make([]float32, itemCount)
		specifiedSec := float32(0)
		for i, p := range secPercents {
			if i < itemCount {
				itemSizes[i] = secondaryAvailable * p / 100
				specifiedSec += p
			}
		}
		remSec := 100 - specifiedSec
		if remSec > 0 && len(secPercents) < itemCount {
			equalSec := remSec / float32(itemCount-len(secPercents))
			for i := len(secPercents); i < itemCount; i++ {
				itemSizes[i] = secondaryAvailable * equalSec / 100
			}
		} else if specifiedSec == 0 {
			equalSec := secondaryAvailable / float32(itemCount)
			for i := range itemSizes {
				itemSizes[i] = equalSec
			}
		}

		posSecondary := float32(0)

		for item := 0; item < itemCount && objIdx < len(objects); item++ {
			obj := objects[objIdx]
			itemSize := itemSizes[item]

			var pos fyne.Position
			var sz fyne.Size
			if isColumnMode {
				// Primary = X, Secondary = Y
				pos = fyne.NewPos(posPrimary, posSecondary)
				sz = fyne.NewSize(groupSize, itemSize)
			} else {
				// Primary = Y, Secondary = X
				pos = fyne.NewPos(posSecondary, posPrimary)
				sz = fyne.NewSize(itemSize, groupSize)
			}

			obj.Move(pos)
			obj.Resize(sz)

			posSecondary += itemSize + padding
			objIdx++
		}

		posPrimary += groupSize + padding
	}
}

func (g *FlexGrid) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(0, 0)
}
