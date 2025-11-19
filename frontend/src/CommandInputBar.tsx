import { useRef, useEffect } from 'react';
import './CommandInputBar.css';
import {Mode} from './Modes';

interface CommandInputBarProps {
  value: string;
  onChange: (value: string) => void;
  mode: Mode;
  loading: boolean;
  onSubmit: (value: string) => void;
}

export function CommandInputBar({
  value,
  onChange,
  mode,
  loading,
  onSubmit,
}: CommandInputBarProps) {
  const inputRef = useRef<HTMLInputElement>(null);

  // Handle Enter key in command mode
  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter' && mode === 'command') {
      e.preventDefault(); // prevent newline if needed
      onSubmit(value);
    }
  };

  // Auto-focus the input when switching to command mode
  useEffect(() => {
    if (mode === 'command') {
      const timeoutId = setTimeout(() => {
        inputRef.current?.focus();
      }, 0);
      return () => clearTimeout(timeoutId);
    } else {
      inputRef.current?.blur();
    }
  }, [mode]);

  return (
    <div className="input-bar">
      <div className="content-wrapper">
        <div className="input-wrapper">
          <input
            ref={inputRef}
            className={`command-input ${mode === 'normal' ? 'mode-normal' : 'mode-command'}`}
            type="text"
            value={value}
            onChange={(e) => onChange(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={
              mode === 'command'
                ? 'Enter Command...'
                : "Normal mode — press 'i' to edit"
            }
            disabled={mode === 'normal'}
            readOnly={mode === 'normal'}
          />
          {loading && <span className="loading-indicator">Running…</span>}
        </div>

        <div className="mode-indicator">
          {mode === 'command' ? 'COMMAND' : 'NORMAL'}
        </div>
      </div>
    </div>
  );
}
