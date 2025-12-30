package main

import "fyne.io/fyne/v2"

type Searcher interface {
	Search(string, bool)
	SearchNext()
	SearchPrev()
	SearchClear()
}

type Mover interface {
	MoveUp()
	MoveDown()
	MoveLeft()
	MoveRight()
}

type Submitter interface {
	Submit()
}

type KeyBinder interface {
	fyne.Focusable

	SetKeyBindings(kbs *KeyBindings)
	WidgetName() string
}

type Message = any

type MessageHandler interface {
	MessageHandle(Message)
}

type Themer interface {
	SetTheme(fyne.Theme, fyne.ThemeVariant)
}
