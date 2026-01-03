package efinui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

type KeyBindings struct {
	keyBindings  map[string]bool
	onKeyBinding func(string)
}

func NewKeyBindings(kbs []string, onKeyBinding func(string)) *KeyBindings {
	keyBindings := map[string]bool{}
	for _, k := range kbs {
		keyBindings[k] = true
	}

	return &KeyBindings{
		keyBindings:  keyBindings,
		onKeyBinding: onKeyBinding,
	}
}

func (kbs *KeyBindings) OnTypedKey(ev *fyne.KeyEvent) bool {
	if kbs == nil {
		return false
	}

	kb := strings.ToLower(string(ev.Name))

	if _, ok := kbs.keyBindings[kb]; ok {
		kbs.onKeyBinding(kb)
		return true
	}

	return false
}

func (kbs *KeyBindings) OnTypedShortcut(sc fyne.Shortcut) bool {
	if kbs == nil {
		return false
	}

	if sc, ok := sc.(*desktop.CustomShortcut); ok {
		kb := ""
		if (sc.Modifier & fyne.KeyModifierControl) > 0 {
			kb += "ctrl "
		}
		if (sc.Modifier & fyne.KeyModifierShift) > 0 {
			kb += "shift "
		}
		if (sc.Modifier & fyne.KeyModifierAlt) > 0 {
			kb += "alt "
		}
		if (sc.Modifier & fyne.KeyModifierSuper) > 0 {
			kb += "super "
		}

		kb += strings.ToLower(string(sc.KeyName))

		if _, ok := kbs.keyBindings[kb]; ok {
			kbs.onKeyBinding(kb)
			return true
		}
	}

	return false
}
