package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (a *appData) createUI() {
	items := []*container.TabItem{}
	for k := range a.shoppingLists {
		items = append(items, a.buildTabItem(a.shoppingLists[k]))
	}
	a.tabs = container.NewDocTabs(items...)

	a.tabs.CreateTab = a.createTab
	a.tabs.OnClosed = func(item *container.TabItem) {
		for index, value := range a.shoppingLists {
			if value.Name == item.Text {
				a.deleteShoppingList(index, value)
				return
			}
		}
	}
	a.tabs.SetTabLocation(container.TabLocationLeading)

	a.win.SetContent(container.NewMax(a.tabs))
	a.win.Resize(fyne.NewSize(800, 600))
}

func (a *appData) buildTabItem(sl *shoppingList) *container.TabItem {
	displayCheckedItem := true
	filter := ""

	sl.list = widget.NewList(func() int {
		if filter == "" && displayCheckedItem {
			return len(sl.Items)
		}

		count := 0
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
		a.setFilteredItem(sl, filter, displayCheckedItem, lii, co)
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
		widget.NewToolbarAction(theme.ContentAddIcon(), a.addItem(sl)),
		visibilityAction,
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.FileTextIcon(), a.importYaml(sl)),
		widget.NewToolbarAction(theme.DownloadIcon(), a.downloadYaml(sl)),
		widget.NewToolbarAction(theme.DocumentSaveIcon(), a.exportYaml(sl)),
		widget.NewToolbarAction(theme.UploadIcon(), a.uploadYaml(sl)),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.ContentClearIcon(), func() {
			keepItem := []item{}
			for _, item := range sl.Items {
				if item.Checked {
					keepItem = append(keepItem, item)
				}
			}
			sl.Items = keepItem
			a.saveShoppingList(sl)
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
	newShoppingList, err := a.newShoppingList(fmt.Sprintf("Unknown place %d", minimalPlaceIndex))
	if err != nil {
		return nil
	}

	newDocItem := a.buildTabItem(newShoppingList)

	newShoppingLocationEntry := widget.NewEntry()
	dialog.ShowForm("New shopping place", "Create", "Cancel",
		[]*widget.FormItem{{Text: "Name", Widget: newShoppingLocationEntry}}, func(confirm bool) {
			if confirm {
				newShoppingList.Name = newShoppingLocationEntry.Text
				a.saveShoppingList(newShoppingList)

				newDocItem.Text = newShoppingList.Name
				a.tabs.Refresh()
			} else {
				a.tabs.Remove(newDocItem)
				for index, value := range a.shoppingLists {
					if value == newShoppingList {
						a.deleteShoppingList(index, value)
					}
				}
			}
		}, a.win)
	a.win.Canvas().Focus(newShoppingLocationEntry)

	return newDocItem
}

func (a *appData) setFilteredItem(sl *shoppingList, filter string, displayCheckedItem bool, index widget.ListItemID, co fyne.CanvasObject) {
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
				a.saveShoppingList(sl)
			}
			c.Refresh()
			return
		}
	}
}

func (a *appData) addItem(sl *shoppingList) func() {
	return func() {
		newItemEntry := widget.NewEntry()
		dialog.ShowForm("New shopping item", "Create", "Cancel",
			[]*widget.FormItem{{Text: "Name", Widget: newItemEntry}}, func(confirm bool) {
				if confirm {
					sl.Items = append(sl.Items, item{What: newItemEntry.Text})
					a.saveShoppingList(sl)
					sl.list.Refresh()
				}
			}, a.win)
		a.win.Canvas().Focus(newItemEntry)
	}
}

func (a *appData) importYaml(sl *shoppingList) func() {
	return func() {
		go func() {
			dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
				if reader == nil || err != nil {
					return
				}
				err = sl.importYaml(reader)
				if err != nil {
					dialog.ShowError(err, a.win)
					return
				}

				a.saveShoppingList(sl)
			}, a.win)
		}()
	}
}

func (sl *shoppingList) downloadYamlDialog(ctx context.Context, cancel context.CancelFunc, win fyne.Window) (dialog.Dialog, *widget.Entry) {
	code := widget.NewEntry()

	d := dialog.NewForm("Download shopping list", "Download", "Cancel",
		[]*widget.FormItem{
			{Text: "Code", Widget: code},
		}, func(confirm bool) {
			if !confirm {
				return
			}

			showProgressBarInfinite(cancel, "Downloading", "Downloading shopping list", func() error {
				return sl.downloadYaml(ctx, code.Text)
			}, win)
		}, win)
	return d, code
}

func (a *appData) downloadYaml(sl *shoppingList) func() {
	return func() {
		ctx, cancel := context.WithCancel(context.Background())

		d, _ := sl.downloadYamlDialog(ctx, cancel, a.win)
		d.Show()
	}
}

func (a *appData) exportYaml(sl *shoppingList) func() {
	return func() {
		go func() {
			dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
				if writer == nil || err != nil {
					return
				}
				err = sl.exportYaml(writer)
				if err != nil {
					dialog.ShowError(err, a.win)
				}
			}, a.win)
		}()
	}
}

func (sl *shoppingList) uploadYamlDialog(ctx context.Context, cancel context.CancelFunc, win fyne.Window) string {
	code, status, err := sl.uploadYaml(ctx)
	if err != nil {
		dialog.ShowError(err, win)
		cancel()
		return ""
	}

	showProgressBarInfinite(cancel, "Wormhole code", "Wormhole code: "+code, func() error {
		s := <-status

		if !s.OK {
			return s.Error
		}
		return nil
	}, win)
	return code
}

func (a *appData) uploadYaml(sl *shoppingList) func() {
	return func() {
		ctx, cancel := context.WithCancel(context.Background())
		sl.uploadYamlDialog(ctx, cancel, a.win)
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
