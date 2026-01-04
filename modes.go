package efinui

type Mode int

func (m Mode) String() string {
	switch m {
	case ModeNormal:
		return "normal"
	case ModeCommand:
		return "command"
	case ModeSearch:
		return "search"
	case ModeHelp:
		return "help"
	}

	return "unknown"
}

const (
	ModeNormal Mode = iota
	ModeCommand
	ModeSearch
	ModeHelp
)
