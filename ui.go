package main

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"gopkg.in/yaml.v2"
)

func (a *appData) buildTabItem(key uint64, sl *shoppingList) *container.TabItem {
	displayCheckedItem := true
	filter := ""

	sl.list = widget.NewList(func() int {
		count := 0
		if filter == "" && displayCheckedItem {
			return len(sl.Items)
		}

		for _, i := range sl.Items {
			if i.shouldFilter(filter, displayCheckedItem) {
				continue
			}
			count++
		}
		return count
	}, func() fyne.CanvasObject {
		return widget.NewCheck("test", func(b bool) {})
	}, func(lii widget.ListItemID, co fyne.CanvasObject) {
		a.setFilteredItem(key, sl, filter, displayCheckedItem, lii, co)
	})

	var toolbar *widget.Toolbar
	var visibilityAction *widget.ToolbarAction

	visibilityAction = widget.NewToolbarAction(theme.VisibilityIcon(), func() {
		if displayCheckedItem {
			visibilityAction.SetIcon(theme.VisibilityOffIcon())
			displayCheckedItem = false
		} else {
			visibilityAction.SetIcon(theme.VisibilityIcon())
			displayCheckedItem = true
		}
		toolbar.Refresh()
		sl.list.Refresh()
	})

	toolbar = widget.NewToolbar(
		widget.NewToolbarAction(theme.ContentAddIcon(), a.addItem(key, sl)),
		visibilityAction,
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.DownloadIcon(), a.importYaml(key, sl)),
		widget.NewToolbarAction(theme.UploadIcon(), a.exportYaml(key, sl)),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.ContentClearIcon(), func() {
			keepItem := []item{}
			for _, item := range sl.Items {
				if item.Checked {
					keepItem = append(keepItem, item)
				}
			}
			sl.Items = keepItem
			a.saveShoppingList(key)
			sl.list.Refresh()
		}),
	)

	sl.filterEntry = widget.NewEntry()
	sl.filterEntry.OnChanged = func(s string) {
		filter = s
		sl.list.Refresh()
	}

	return container.NewTabItem(sl.Name, container.NewBorder(
		container.NewBorder(nil, nil, widget.NewLabel("Filter"), nil, sl.filterEntry),
		toolbar,
		nil, nil,
		sl.list),
	)
}

func (a *appData) createTab() *container.TabItem {
	minimalPlaceIndex := a.minimalPlaceIndex()
	newShoppingList, key, err := a.newTuppleKeyShoppingList(fmt.Sprintf("Unknown place %d", minimalPlaceIndex))
	if err != nil {
		return nil
	}

	newDocItem := a.buildTabItem(key, a.shoppingLists[key])

	newShoppingLocationEntry := widget.NewEntry()
	dialog.ShowForm("New shopping place", "Create", "Cancel",
		[]*widget.FormItem{{Text: "Name", Widget: newShoppingLocationEntry}}, func(confirm bool) {
			if confirm {
				newShoppingList.Name = newShoppingLocationEntry.Text
				a.saveShoppingList(key)

				newDocItem.Text = newShoppingList.Name
				a.tabs.Refresh()
			} else {
				a.tabs.Remove(newDocItem)
				delete(a.shoppingLists, key)
				a.deleteShoppingList(key)
			}
		}, a.win)
	a.win.Canvas().Focus(newShoppingLocationEntry)

	return newDocItem
}

func (a *appData) setFilteredItem(key uint64, sl *shoppingList, filter string, displayCheckedItem bool, index widget.ListItemID, co fyne.CanvasObject) {
	var pos widget.ListItemID
	for realIndex, i := range sl.Items {
		if i.shouldFilter(filter, displayCheckedItem) {
			continue
		}

		pos++

		if pos-1 == index {
			c := co.(*widget.Check)
			c.Text = i.What
			c.Checked = i.Checked
			c.OnChanged = func(b bool) {
				sl.Items[realIndex].Checked = b
				a.saveShoppingList(key)
			}
			c.Refresh()
			return
		}
	}
}

func (a *appData) addItem(key uint64, sl *shoppingList) func() {
	return func() {
		newItemEntry := widget.NewEntry()
		dialog.ShowForm("New shopping item", "Create", "Cancel",
			[]*widget.FormItem{{Text: "Name", Widget: newItemEntry}}, func(confirm bool) {
				if confirm {
					sl.Items = append(sl.Items, item{What: newItemEntry.Text})
					a.saveShoppingList(key)
					sl.list.Refresh()
				}
			}, a.win)
		a.win.Canvas().Focus(newItemEntry)
	}
}

func (a *appData) importYaml(key uint64, sl *shoppingList) func() {
	return func() {
		go func() {
			dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
				if reader == nil || err != nil {
					return
				}

				defer reader.Close()

				// Import shopping list from a yaml file to the selected shopping list sl
				err = yaml.NewDecoder(reader).Decode(sl)
				if err != nil {
					return
				}

				a.saveShoppingList(key)
			}, a.win)
		}()
	}
}

func (a *appData) exportYaml(key uint64, sl *shoppingList) func() {
	return func() {
		go func() {
			dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
				if writer == nil || err != nil {
					return
				}
				defer writer.Close()

				// Export shopping list sl to a yaml file
				err = yaml.NewEncoder(writer).Encode(sl)
				if err != nil {
					return
				}
			}, a.win)
		}()
	}
}

func (i *item) shouldFilter(filter string, displayCheckedItem bool) bool {
	if filter != "" {
		if !strings.Contains(i.What, filter) {
			return true
		}
	}

	if !displayCheckedItem && i.Checked {
		return true
	}
	return false
}

func (a *appData) minimalPlaceIndex() int {
	index := 1
	for _, sl := range a.shoppingLists {
		if !strings.HasPrefix(sl.Name, "Unknown place ") {
			continue
		}

		n, err := strconv.Atoi(strings.TrimPrefix(sl.Name, "Unknown place "))
		if err != nil {
			continue
		}

		if n >= index {
			index = n + 1
		}
	}

	return index
}
