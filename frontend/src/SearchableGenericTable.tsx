// frontend/src/components/SearchableGenericTable.tsx
import { useState, useEffect } from 'react';
import { GenericTable, GenericTableProps } from './GenericTable';
import './SearchableGenericTable.css';

export interface SearchableGenericTableProps extends GenericTableProps {
    search?: string;
}

export function SearchableGenericTable({
  headers,
  rows,
  onRowAction,
  enableKeybindings = true,
  search,
}: SearchableGenericTableProps) {

  let filteredRows = rows;
  if (search && search.trim()) {
      const term = search.toLowerCase();
      filteredRows = rows.filter((row) =>
        row.some((cell) => {
          if (cell === null || cell === undefined) return false;
          return String(cell).toLowerCase().includes(term);
        })
      );
  }

  const showResultCount = search && search.length > 0;

  return (
    <div className="searchable-table-wrapper">
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
