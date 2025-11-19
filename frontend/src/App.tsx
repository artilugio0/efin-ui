import {useState, useEffect} from 'react';
import logo from './assets/images/logo-universal.png';
import './App.css';
import {Mode} from './Modes';
import {CommandInputBar} from './CommandInputBar';
import {GenericTable} from './GenericTable';
import {RequestResponseDetail} from './RequestResponseDetail';
import {EvalCommand, RowAction} from "../wailsjs/go/main/App";
import {main as models} from "../wailsjs/go/models";

function App() {
    const [mode, setMode] = useState<Mode>('normal');
    const [activeView, setActiveView] = useState('start');
    const [commandResult, setCommandResult] = useState<models.CommandResult | null>(null);
    const [commandInputValue, setCommandInputValue] = useState('');
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
        if (!cmd.trim()) return;
        setLoading(true);
        try {
            const result = await EvalCommand(cmd);
            setActiveView(result.result_type);
            setCommandResult(result);
            console.log(JSON.stringify(result));
        } catch (err) {
            console.error(err);
            setCommandResult(null);
        } finally {
            setLoading(false);
            setMode('normal');
        }
    };

    useEffect(() => {
      const handler = (e: KeyboardEvent) => {
        if (e.key === 'Escape') {
          e.preventDefault();
          setMode('normal');
        }

        if (e.key === 'i' && mode === 'normal') {
          e.preventDefault();
          setMode('command');
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
                        <RequestResponseDetail data={commandResult.request_response_detail} />
                    ) : activeView === 'request_response_table' ? (
                        <GenericTable
                            headers={commandResult?.request_response_table[0] || []}
                            rows={commandResult?.request_response_table.slice(1) || []}
                            enableKeybindings={mode === 'normal'}
                            onRowAction={handleRowAction}
                        />
                    ) : null
                    }
                </div>
            </div>

            <CommandInputBar
                value={commandInputValue}
                onChange={setCommandInputValue}
                mode={mode}
                loading={loading}
                onSubmit={evalCommand}
            />
        </div>
    )
}

export default App
