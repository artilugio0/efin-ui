import './GenericTable.css';
import { useState, useEffect, useRef } from 'react';

export interface GenericTableProps {
  /** Array of column header names */
  headers: string[];
  /** 2D array of row data — rows[i][j] corresponds to column j in row i */
  rows: (string | number | null | undefined)[][];
  /** Enable Vim-like navigation (h/j/k/l, Enter, Tab for expand) */
  enableKeybindings?: boolean;
  /** Called when user presses Enter on a data row */
  onRowAction?: (rowData: Record<string, string>) => void;
}

export function GenericTable({
  headers,
  rows,
  enableKeybindings = true,
  onRowAction,
}: GenericTableProps) {
  const [selectedRow, setSelectedRow] = useState(0); // 0 = header row, 1+ = data rows
  const [selectedCol, setSelectedCol] = useState<number | null>(null);
  const [expandedCell, setExpandedCell] = useState<{ row: number; col: number } | null>(null);

  const tableWrapperRef = useRef<HTMLDivElement>(null);

  const totalRows = rows.length; // number of data rows
  const totalCols = headers.length;

  // Reset selection when data changes
  useEffect(() => {
    setSelectedRow(totalRows > 0 ? 1 : 0);
    setSelectedCol(null);
    setExpandedCell(null);
  }, [headers, rows]);

  // Keyboard navigation
  useEffect(() => {
    if (!enableKeybindings || totalRows === 0 || totalCols === 0) return;


    const handleKey = (e: KeyboardEvent) => {
      if (!['h', 'j', 'k', 'l', 'Enter', 'Tab'].includes(e.key)) return;
        const active = document.activeElement;
        if (active && (active.tagName === 'INPUT' || active.tagName === 'TEXTAREA')) {
          return;
        }
      e.preventDefault();

      let newRow = selectedRow;
      let newCol = selectedCol ?? 0;

      switch (e.key) {
        case 'j': // down
          newRow = Math.min(totalRows + 1, selectedRow + 1); // +1 because header is 0
          setSelectedCol(null);
          break;
        case 'k': // up
          newRow = Math.max(0, selectedRow - 1);
          setSelectedCol(null);
          break;
        case 'h': // left
          newCol = selectedCol === null ? 0 : Math.max(0, selectedCol - 1);
          setSelectedCol(newCol);
          break;
        case 'l': // right
          newCol = selectedCol === null ? 0 : Math.min(totalCols - 1, selectedCol + 1);
          setSelectedCol(newCol);
          break;
        case 'Tab': {
          const row = selectedRow;
          const col = selectedCol ?? 0;
          const cell = tableWrapperRef.current?.querySelector(
            `.table-cell[data-row="${row}"][data-col="${col}"]`
          ) as HTMLElement | null;

          if (cell && cell.scrollWidth > cell.clientWidth) {
            setExpandedCell(prev =>
              prev?.row === row && prev?.col === col ? null : { row, col }
            );
          }
          return;
        }
        case 'Enter':
          if (selectedRow > 0 && onRowAction) {
            const rowData = rows[selectedRow - 1];
            const rowObject = Object.fromEntries(
              headers.map((header, idx) => [header, String(rowData[idx]) ?? null])
            );
            onRowAction(rowObject);
          }
          return;
      }

      setSelectedRow(newRow);

      // Scroll into view
      setTimeout(() => {
        const targetRow = newRow;
        const targetCol = selectedCol ?? 0;
        const cell = tableWrapperRef.current?.querySelector(
          `.table-cell[data-row="${targetRow}"][data-col="${targetCol}"]`
        ) as HTMLElement | null;

        if (cell) {
          if (newRow === 0) {
            tableWrapperRef.current?.scrollTo({ top: 0, behavior: 'smooth' });
          } else {
            cell.scrollIntoView({ behavior: 'smooth', block: 'center', inline: 'nearest' });
          }
        }
      }, 0);
    };

    window.addEventListener('keydown', handleKey);
    return () => window.removeEventListener('keydown', handleKey);
  }, [
    enableKeybindings,
    selectedRow,
    selectedCol,
    expandedCell,
    totalRows,
    totalCols,
    headers,
    rows,
    onRowAction,
  ]);

  const isRowSelected = (r: number) => selectedRow === r;
  const isCellSelected = (r: number, c: number) => selectedCol !== null && selectedRow === r && selectedCol === c;
  const isExpanded = (r: number, c: number) => expandedCell?.row === r && expandedCell?.col === c;

  return (
    <div className="generic-table-container">
      <div className="table-wrapper" ref={tableWrapperRef}>
        <table className="generic-table">
          <thead>
            <tr className={isRowSelected(0) ? 'row-selected' : ''}>
              {headers.map((header, colIdx) => (
                <th
                  key={colIdx}
                  className={`table-cell ${isCellSelected(0, colIdx) ? 'cell-selected' : ''} ${
                    isExpanded(0, colIdx) ? 'hover-expand' : ''
                  }`}
                  data-row="0"
                  data-col={colIdx}
                >
                  {header}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {rows.map((row, rowIdx) => {
              const visualRow = rowIdx + 1; // 1-based for data rows
              return (
                <tr
                  key={rowIdx}
                  className={`${isRowSelected(visualRow) ? 'row-selected' : ''} ${
                    rowIdx % 2 === 0 ? 'generic-table-row-even' : 'generic-table-row-odd'
                  }`}
                >
                  {row.map((cell, colIdx) => (
                    <td
                      key={colIdx}
                      className={`table-cell ${isCellSelected(visualRow, colIdx) ? 'cell-selected' : ''} ${
                        isExpanded(visualRow, colIdx) ? 'hover-expand' : ''
                      }`}
                      data-row={visualRow}
                      data-col={colIdx}
                    >
                      {cell == null ? '' : String(cell)}
                    </td>
                  ))}
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}
