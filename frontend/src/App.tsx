import {useState, useEffect} from 'react';
import logo from './assets/images/logo-universal.png';
import './App.css';
import {Mode} from './Modes';
import {SuggestionInput} from './SuggestionInput';
import {CommandInputBar} from './CommandInputBar';
import {SearchableGenericTable} from './SearchableGenericTable';
import {RequestResponseDetail} from './RequestResponseDetail';
import {EvalCommand, RowAction, SuggestCommand} from "../wailsjs/go/main/App";
import {main as models} from "../wailsjs/go/models";

function App() {
    const [mode, setMode] = useState<Mode>('normal');
    const [activeView, setActiveView] = useState('start');
    const [commandResult, setCommandResult] = useState<models.CommandResult | null>(null);
    const [search, setSearch] = useState<string>('');
    const [loading, setLoading] = useState(false);

    const handleRowAction = async (rowObject: Record<string, string>) => {
        setLoading(true);
        try {
            const result = await RowAction(rowObject);
            setActiveView(result.result_type);
            setCommandResult(result);
        } catch (err) {
            console.error("RowAction failed", err);
            setCommandResult(null);
        } finally {
            setLoading(false);
            setMode('normal');
        }
    };

    const evalCommand = async (cmd: string) => {
        setLoading(true);
        try {
            const result = await EvalCommand(cmd);
            setActiveView(result.result_type);
            setCommandResult(result);
        } catch (err) {
            console.error(err);
            setCommandResult(null);
        } finally {
            setLoading(false);
            setMode('normal');
        }
    };

    const handleSubmit = (submitted: string) => {
        if (!submitted.trim()) return;
        switch (mode) {
        case 'command':
            evalCommand(submitted);
            setSearch('');
            return;
        case 'search':
            setSearch(submitted);
            setMode('normal');
            return;
        }
    };

    useEffect(() => {
      const handler = (e: KeyboardEvent) => {
        if (e.key === 'Escape') {
          e.preventDefault();
          if (mode == 'normal') {
              setSearch('');
          }
          setMode('normal');
        }

        if (e.key === 'i' && mode === 'normal') {
          const active = document.activeElement;
          if (active && (active.tagName === 'INPUT' || active.tagName === 'TEXTAREA')) {
            return;
          }

          e.preventDefault();
          setMode('command');
        }

        if (e.key === '/' && mode === 'normal') {
          const active = document.activeElement;
          if (active && (active.tagName === 'INPUT' || active.tagName === 'TEXTAREA')) {
            return;
          }

          e.preventDefault();
          setMode('search');
        }
      };

      window.addEventListener('keydown', handler);

      return () => {
        window.removeEventListener('keydown', handler);
      };
    }, [mode, setMode]);

    return (
        <div className="app-container">
            <div className="results-area">
                <div className="content-wrapper">
                    {loading ? (
                        <div className="loading-placeholder">Loading data...</div>
                    ) : activeView === 'request_response_detail' && commandResult !== null ? (
                        <RequestResponseDetail
                            data={commandResult.request_response_detail}
                            search={search}
                        />
                    ) : activeView === 'request_response_table' ? (
                        <SearchableGenericTable
                            headers={commandResult?.request_response_table[0] || []}
                            rows={commandResult?.request_response_table.slice(1) || []}
                            enableKeybindings={mode === 'normal'}
                            onRowAction={handleRowAction}
                            search={search}
                        />
                    ) : null
                    }
                </div>
            </div>

            <CommandInputBar
                mode={mode}
                onSubmit={handleSubmit}
                suggest={SuggestCommand}
            />
        </div>
    )
}

export default App
