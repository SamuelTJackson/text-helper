package main

import (
	"bytes"
	"context"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/container"
	"fyne.io/fyne/driver/desktop"
	"fyne.io/fyne/widget"
)

type VimEntry struct {
	widget.Entry
	insertMode bool
	okCallback func()
}

func NewVimEntry() *VimEntry {
	v := &VimEntry{
		insertMode: true,
	}

	v.ExtendBaseWidget(v)
	v.Wrapping = fyne.TextWrapWord
	v.MultiLine = true

	return v
}

func (v *VimEntry) TypedRune(r rune) {
	if v.insertMode {
		v.Entry.TypedRune(r)
	} else {
		switch r {
		case 'i':
			v.insertMode = true
		case 'h':
			v.CursorColumn--
			if v.CursorColumn < 0 {
				v.CursorColumn = 0
			}
			v.Refresh()
		case 'l':
			v.CursorColumn++
			if v.CursorColumn > len(v.Text) {
				v.CursorColumn = len(v.Text)
			}
			v.Refresh()
		case 'j':
			v.CursorRow++
			v.Refresh()
		case 'k':
			if v.CursorRow > 0 {
				v.CursorRow--
			}
			v.Refresh()
		}
	}
}

func (v *VimEntry) TypedKey(ev *fyne.KeyEvent) {
	switch ev.Name {
	case fyne.KeyEscape:
		v.insertMode = false
	default:
		if v.insertMode {
			v.Entry.TypedKey(ev)
		}
	}
}
func (v *VimEntry) TypedShortcut(s fyne.Shortcut) {
	if _, ok := s.(*desktop.CustomShortcut); !ok {
		v.Entry.TypedShortcut(s)

		return
	}

	shortCut, ok := s.(*desktop.CustomShortcut)
	if !ok {
		return
	}

	if shortCut.KeyName == fyne.KeyReturn && shortCut.Modifier == desktop.ControlModifier {
		v.okCallback()
	}
}

type gui struct {
	app        fyne.App
	mainWindow fyne.Window
}

func newGui(ctx context.Context, geminiClient *geminiClient) (*gui, error) {
	app := app.NewWithID("texthelper")
	window := app.NewWindow("Text Helper")
	window.SetPadded(false)

	textArea := NewVimEntry()
	textArea.Wrapping = fyne.TextWrapWord

	okCallback := func() {
		c := geminiClient.SendMessage(ctx, textArea.Text)
		textArea.Text = ""
		textArea.Refresh()

		buf := bytes.NewBufferString("")
		for t := range c {
			buf.WriteString(t)

			textArea.Text = buf.String()
			textArea.Refresh()
		}

		text := buf.String()
		window.Clipboard().SetContent(text)
	}

	textArea.okCallback = okCallback

	scroll := container.NewVScroll(textArea)
	scroll.SetMinSize(fyne.NewSize(0, 300))

	clearButton := widget.NewButton("Clear", func() {
		textArea.Text = ""
		textArea.Refresh()
	})
	submit := widget.NewButton("Ok", func() {
		okCallback()
	})

	buttonBar := widget.NewHBox(clearButton, submit)
	buttonBarFullWidth := container.NewMax(buttonBar)

	window.SetContent(container.NewBorder(
		nil,
		buttonBarFullWidth,
		nil,
		nil,
		scroll,
	))

	window.Canvas().Focus(textArea)

	return &gui{
		app:        app,
		mainWindow: window,
	}, nil
}

func (g *gui) Show(ctx context.Context) {
	go func() {
		<-ctx.Done()
		g.mainWindow.Close()

	}()
	g.mainWindow.ShowAndRun()
}
