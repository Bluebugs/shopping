package main

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func showProgressBarInfinite(cancel context.CancelFunc,
	title, text string,
	blocking func() error,
	parent fyne.Window) {

	displayCode := container.NewHBox(layout.NewSpacer(), widget.NewLabel(text), layout.NewSpacer())
	content := container.NewVBox(displayCode, widget.NewProgressBarInfinite())

	cancelled := false

	d := dialog.NewCustom(title, "Cancel", content, parent)
	d.SetOnClosed(func() {
		cancelled = true
		cancel()
	})

	go func() {
		err := blocking()
		if err != nil && !cancelled {
			d.Hide()
			dialog.ShowError(err, parent)
		} else {
			d.Hide()
		}
	}()

	d.Show()
}
