import {useState, useEffect} from 'react';
import logo from './assets/images/logo-universal.png';
import './App.css';
import {Mode} from './Modes';
import {SuggestionInput} from './SuggestionInput';
import {CommandInputBar} from './CommandInputBar';
import {SearchableGenericTable} from './SearchableGenericTable';
import {RequestResponseDetail} from './RequestResponseDetail';
import {EvalUIAction} from "../wailsjs/go/main/App";
import {main as models} from "../wailsjs/go/models";
import {Pane} from "./Pane";

function App() {
    const [mode, setMode] = useState<Mode>('normal');
    const [search, setSearch] = useState<string>('');
    const [loading, setLoading] = useState(false);

    const [currentTab, setCurrentTab] = useState<number>(0);
    const [focusedPane, setFocusedPane] = useState<number[]>([0]);

    const [contents, setContents] = useState<any>([{result_type: 'start'}]);
    const [tabs, setTabs] = useState([{
        layout: 'vsplit',
        panes: [
            {
                content: 0,
                layout: 'single',
            },
        ],
    }]);

    function updateFocusedPaneContent(currentPane: any, focusedPane: number[], newContent: number) {
        if (focusedPane.length === 1) {
            currentPane.panes[focusedPane[0]].content = newContent;
            return;
        }

        updateFocusedPaneContent(
            currentPane.panes[focusedPane[0]],
            focusedPane.slice(1),
            newContent,
        );
    }

    function renderPane(pane: any, focusedPane: number[]) {
        return (
            <Pane layout='vsplit'>
                {pane.panes.map((p: any, i: number) => (
                    contents[p.content].result_type === 'request_response_detail' ? (
                        <RequestResponseDetail
                            data={contents[p.content].request_response_detail}
                            search={search}
                        />
                    ) : contents[p.content].result_type === 'request_response_table' ? (
                        <SearchableGenericTable
                            headers={contents[p.content].request_response_table[0] || []}
                            rows={contents[p.content].request_response_table.slice(1) || []}
                            enableKeybindings={mode === 'normal' && focusedPane.length === 1 && focusedPane[0] === i}
                            onRowAction={handleRowAction}
                            search={search}
                        />
                    ) : contents[p.content].result_type === 'start' ? (
                        <p>Start page</p>
                    ) : null
                ))}
            </Pane>
        );
    }

    // Add this state
    const [theme, setTheme] = useState<'light' | 'dark'>('dark');

    const handleRowAction = async (rowObject: Record<string, string>) => {
        setLoading(true);
        try {
            const result = await EvalUIAction({
                action_type: "row_submitted",
                row_submitted: rowObject,
            });
            setContents((prev: any[]) => [...prev, result]);
            setTabs(prev => {
                updateFocusedPaneContent(prev[currentTab], focusedPane, contents.length)
                return {...prev};
            });
        } catch (err) {
            console.error("RowAction failed", err);
        } finally {
            setLoading(false);
            setMode('normal');
        }
    };

    const updateUIState = (uiState: any) => {
        console.log("before", currentTab, tabs, focusedPane);
        console.log("after", uiState);

        setCurrentTab(uiState.current_tab);
        setTabs(uiState.tabs);
        setFocusedPane(uiState.focused_pane);
    };

    const evalCommand = async (cmd: string) => {
        setLoading(true);
        try {
            const result = await EvalUIAction({
                action_type: "command_submitted",
                command_submitted: cmd,
            });

            if (result.result_type !== "ui_state_updated") {
                setContents((prev: any[]) => [...prev, result]);
            }

            if (result.ui_state) {
                updateUIState(result.ui_state);
            }

        } catch (err) {
            console.error(err);
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
        EvalUIAction({
            action_type: "ui_state_requested",
        }).then(result => {;
            updateUIState(result.ui_state);
        });
    }, []);

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
        <div className="app-container" data-theme={theme}>
            <div className="results-area">
                {renderPane(tabs[currentTab], focusedPane)}
            </div>

            <CommandInputBar
                mode={mode}
                onSubmit={handleSubmit}
                suggest={async (cmd: string) => (await EvalUIAction({
                    action_type: "command_suggestion_requested",
                    command_suggestion_requested: cmd,
                })).command_suggestion}
            />
        </div>
    )
}

export default App
