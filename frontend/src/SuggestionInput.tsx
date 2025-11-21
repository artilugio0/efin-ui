import React from 'react';
import './SuggestionInput.css';
import {useEffect, useState, useRef} from 'react';

interface SuggestionInputProps {
  value: string;
  suggestions: string[];
  onSubmit: (value: string) => void;
  onChange: (value: string) => void;
  disabled: boolean;
  focus: boolean;
}

export function SuggestionInput({
    value,
    suggestions,
    onSubmit,
    onChange,
    disabled,
    focus,
}: SuggestionInputProps) {
  const inputRef = useRef<HTMLInputElement>(null);

  const [suggestionIndex, setSuggestionIndex] = useState<number>(0);

  const validSuggestions = suggestions.filter(s =>
    s.toLowerCase().startsWith(value.toLowerCase()) && s !== value
  );

  let currentIndex = suggestionIndex;
  if (currentIndex >= validSuggestions.length) {
      currentIndex = 0;
      if (suggestionIndex > 0) {
          setSuggestionIndex(0);
      }
  }

  const suggestion = validSuggestions.length > 0 ? validSuggestions[currentIndex].slice(value.length) : '';

  // Handle Enter key in command mode
  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter' && !disabled) {
      e.preventDefault(); // prevent newline if needed
      e.stopPropagation()
      onSubmit(value);
      onChange('');
    }

    if (e.key === 'Tab') {
        e.preventDefault();
      }

    if (validSuggestions.length > 0) {
      if (e.key === 'ArrowDown') {
        e.preventDefault();
        setSuggestionIndex(prev => (prev + 1) % validSuggestions.length);
        return;
      }
      if (e.key === 'ArrowUp') {
        e.preventDefault();
        setSuggestionIndex(prev => (prev - 1 + validSuggestions.length) % validSuggestions.length);
        return;
      }
      if ((e.key === 'Tab' || e.key === 'ArrowRight') && currentIndex >= 0) {
        console.log('suggestion accepted')
        e.preventDefault();
        onChange(validSuggestions[currentIndex]);
        setSuggestionIndex(0);
        return;
      }
    }
  }

  useEffect(() => {
    if (focus) {
      const timeoutId = setTimeout(() => {
        inputRef.current?.focus();
      }, 0);
      return () => clearTimeout(timeoutId);
    } else {
      inputRef.current?.blur();
    }
  }, [focus]);

  return (
      <div className="autocomplete-wrapper">
          {/* Hidden real input for proper selection behavior */}
          <input
              type="text"
              ref={inputRef}
              value={value}
              onChange={(e) => onChange(e.target.value)}
              className="autocomplete-input"
              onKeyDown={handleKeyDown}
              disabled={disabled}
          />

          {/* Visual representation */}
          <div className="autocomplete-visual" aria-hidden="true">
              <span className="autocomplete-hidden-text">{value || " "}</span>
              <span className="autocomplete-suggestion">{suggestion || " "}</span>
          </div>
      </div>
  );
};
