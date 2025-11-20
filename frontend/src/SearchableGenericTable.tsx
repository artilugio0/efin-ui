// frontend/src/components/SearchableGenericTable.tsx
import { useState, useMemo, useEffect } from 'react';
import { GenericTable, GenericTableProps } from './GenericTable';
import { SearchBox } from './SearchBox';
import './SearchableGenericTable.css';

export function SearchableGenericTable({
  headers,
  rows,
  onRowAction,
  enableKeybindings = true,
}: GenericTableProps) {
  const [searchTerm, setSearchTerm] = useState('');

  const filteredRows = useMemo(() => {
    if (!searchTerm.trim()) return rows;

    const term = searchTerm.toLowerCase();
    return rows.filter((row) =>
      row.some((cell) => {
        if (cell === null || cell === undefined) return false;
        return String(cell).toLowerCase().includes(term);
      })
    );
  }, [rows, searchTerm]);

  const showResultCount = searchTerm.length > 0;

  return (
    <div className="searchable-table-wrapper">
      {/* Search Box - Top Right */}
      <div className="searchable-table-search-container">
        <SearchBox
          value={searchTerm}
          onChange={setSearchTerm}
          placeholder={`Search ${rows.length} rows...`}
        />
      </div>

      {/* Optional: Show result count */}
      {showResultCount && (
        <div className="searchable-table-result-count">
          {filteredRows.length === rows.length
            ? 'No results'
            : `${filteredRows.length} of ${rows.length} rows`}
        </div>
      )}


      {/* The actual table */}
      <GenericTable
        headers={headers}
        rows={filteredRows}
        onRowAction={onRowAction}
        enableKeybindings={enableKeybindings}
      />
    </div>
  );
}
