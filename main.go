package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"go.etcd.io/bbolt"
)

type item struct {
	What    string
	Checked bool
}

type shoppingList struct {
	Name  string
	Items []item

	key uint64

	list        *widget.List
	filterEntry *widget.Entry
}

type appData struct {
	shoppingLists []*shoppingList

	db *bbolt.DB

	app  fyne.App
	win  fyne.Window
	tabs *container.DocTabs
}

func main() {
	a := app.NewWithID("github.com.bluebugs.shopping")

	myApp := &appData{shoppingLists: []*shoppingList{}, app: a, win: a.NewWindow("Shopping List")}

	if err := myApp.loadShoppingLists(); err != nil {
		log.Panic(err)
	}

	myApp.createUI()
	myApp.win.SetOnClosed(func() {
		myApp.Close()
	})
	myApp.win.ShowAndRun()
}
