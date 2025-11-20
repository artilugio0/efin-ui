// frontend/src/components/SearchBox.tsx
import { useState, useEffect } from 'react';
import './SearchBox.css';

type SearchBoxProps = {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  debounceMs?: number;
};

export function SearchBox({
  value,
  onChange,
  placeholder = 'Search...',
  debounceMs = 250,
}: SearchBoxProps) {
  const [localValue, setLocalValue] = useState(value);

  useEffect(() => {
    setLocalValue(value);
  }, [value]);

  useEffect(() => {
    const timer = setTimeout(() => {
      onChange(localValue);
    }, debounceMs);
    return () => clearTimeout(timer);
  }, [localValue, debounceMs, onChange]);

  return (
    <div className="search-box-container">
      <svg
        className="search-box-icon"
        viewBox="0 0 24 24"
      >
        <path d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
      </svg>

      <input
        type="text"
        value={localValue}
        onChange={(e) => setLocalValue(e.target.value)}
        placeholder={placeholder}
        className="search-box-input"
      />

      {localValue && (
        <button
          onClick={() => setLocalValue('')}
          className="search-box-clear"
          aria-label="Clear search"
        >
          <svg viewBox="0 0 24 24">
            <path d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      )}
    </div>
  );
}
