import { useEffect, useState, useCallback} from 'react';
import './CommandInputBar.css';
import {SuggestionInput} from './SuggestionInput';
import {Mode} from './Modes';

interface CommandInputBarProps {
  mode: Mode;
  onSubmit: (value: string) => void;
  suggest: (value: string) => Promise<string[]>;
}

export function CommandInputBar({
  mode,
  onSubmit,
  suggest,
}: CommandInputBarProps) {
  const [value, setValue] = useState('');
  const [suggestions, setSuggestions] = useState<string[]>([]);
  const [selectedIndex, setSelectedIndex] = useState<number>(-1);

  const fetchSuggestions = useCallback(async (text: string) => {
    if (!suggest) {
      setSuggestions([]);
      setSelectedIndex(-1);
      return;
    }

    try {
      const results = await suggest(text);
      const filtered = results.filter(s =>
        s.toLowerCase().startsWith(text.toLowerCase()) && s !== text
      );
      setSuggestions(filtered);
      setSelectedIndex(filtered.length > 0 ? 0 : -1);
    } catch (err) {
      console.error('Suggestion error:', err);
      setSuggestions([]);
      setSelectedIndex(-1);
    }
  }, [suggest]);

  // Debounced fetch
  useEffect(() => {
    const id = setTimeout(() => fetchSuggestions(value), 100);
    return () => clearTimeout(id);
  }, [value, fetchSuggestions]);


  return (
    <div className="input-bar">
      <div className="content-wrapper">
        <SuggestionInput 
            value={value}
            onChange={setValue}
            suggestions={suggestions}
            onSubmit={onSubmit}
            disabled={mode === 'normal'}
            focus={mode === 'command' || mode === 'search'}
        />

        <div className="mode-indicator">
          {mode.toUpperCase()}
        </div>
      </div>
    </div>
  );
}
