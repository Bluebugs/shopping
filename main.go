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

	items := []*container.TabItem{}
	for k := range myApp.shoppingLists {
		items = append(items, myApp.buildTabItem(myApp.shoppingLists[k]))
	}
	myApp.tabs = container.NewDocTabs(items...)

	myApp.tabs.CreateTab = myApp.createTab
	myApp.tabs.OnClosed = func(item *container.TabItem) {
		for index, value := range myApp.shoppingLists {
			if value.Name == item.Text {
				myApp.deleteShoppingList(index, value)
				return
			}
		}
	}
	myApp.tabs.SetTabLocation(container.TabLocationLeading)

	myApp.win.SetContent(container.NewBorder(nil, nil, nil, nil, myApp.tabs))
	myApp.win.Resize(fyne.NewSize(800, 600))
	myApp.win.SetOnClosed(func() {
		myApp.Close()
	})
	myApp.win.ShowAndRun()
}
