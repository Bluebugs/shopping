package main

import (
	"io/ioutil"
	"os"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/bbolt"
)

func setupAppData() (*appData, func()) {
	tempFile, err := ioutil.TempFile("", "test.db")
	if err != nil {
		panic(err)
	}
	tempFile.Close()
	os.Remove(tempFile.Name())
	db, err := bbolt.Open(tempFile.Name(), 0600, nil)
	if err != nil {
		panic(err)
	}

	app := test.NewApp()
	win := test.NewWindow(nil)

	return &appData{
			shoppingLists: []*shoppingList{
				{
					Name: "test",
					Items: []item{
						{
							What:    "unchecked",
							Checked: false,
						},
						{
							What:    "checked",
							Checked: true,
						},
					},
				},
			},
			db:  db,
			app: app,
			win: win,
		}, func() {
			db.Close()
			os.Remove(tempFile.Name())
		}
}

func Test_CheckUncheck(t *testing.T) {
	a, done := setupAppData()
	defer done()
	assert.NotNil(t, a)

	a.createUI()
	assert.NotNil(t, a.tabs)

	test.AssertRendersToImage(t, "create_ui.png", a.win.Canvas())
	test.AssertRendersToMarkup(t, "create_ui.markup", a.win.Canvas())

	test.TapCanvas(a.win.Canvas(), fyne.NewPos(100, 60))
	test.AssertRendersToMarkup(t, "check_unchecked.markup", a.win.Canvas())
	assert.True(t, a.shoppingLists[0].Items[0].Checked)

	test.TapCanvas(a.win.Canvas(), fyne.NewPos(100, 60))
	test.AssertRendersToMarkup(t, "uncheck_checked.markup", a.win.Canvas())
	assert.False(t, a.shoppingLists[0].Items[0].Checked)
}

func Test_AddCancelShoppingList(t *testing.T) {
	a, done := setupAppData()
	defer done()
	assert.NotNil(t, a)

	a.createUI()
	assert.NotNil(t, a.tabs)

	test.AssertRendersToMarkup(t, "create_ui.markup", a.win.Canvas())

	test.TapCanvas(a.win.Canvas(), fyne.NewPos(21, 578))
	test.AssertRendersToMarkup(t, "add_shopping_list.markup", a.win.Canvas())

	test.Type(a.win.Canvas().Focused(), "new shopping list")
	test.AssertRendersToMarkup(t, "new_shopping_list.markup", a.win.Canvas())

	test.TapCanvas(a.win.Canvas(), fyne.NewPos(360, 336))
	test.AssertRendersToMarkup(t, "removed.markup", a.win.Canvas())

	assert.Len(t, a.shoppingLists, 1)
}

func Test_AddModifyRemoveShoppingList(t *testing.T) {
	a, done := setupAppData()
	defer done()
	assert.NotNil(t, a)

	a.createUI()
	assert.NotNil(t, a.tabs)

	test.AssertRendersToMarkup(t, "create_ui.markup", a.win.Canvas())

	test.TapCanvas(a.win.Canvas(), fyne.NewPos(21, 578))
	test.AssertRendersToMarkup(t, "add_shopping_list.markup", a.win.Canvas())

	test.Type(a.win.Canvas().Focused(), "new shopping list")
	test.AssertRendersToMarkup(t, "new_shopping_list.markup", a.win.Canvas())

	test.TapCanvas(a.win.Canvas(), fyne.NewPos(452, 336))
	test.AssertRendersToMarkup(t, "shopping_list_added.markup", a.win.Canvas())

	test.TapCanvas(a.win.Canvas(), fyne.NewPos(180, 579))
	test.AssertRendersToMarkup(t, "shopping_list_new_item.markup", a.win.Canvas())

	test.Type(a.win.Canvas().Focused(), "new item")
	test.AssertRendersToMarkup(t, "shopping_list_new_item_typed.markup", a.win.Canvas())

	test.TapCanvas(a.win.Canvas(), fyne.NewPos(452, 336))
	test.AssertRendersToMarkup(t, "shopping_list_new_item_added.markup", a.win.Canvas())

	assert.Len(t, a.shoppingLists, 2)
	assert.Len(t, a.shoppingLists[0].Items, 2)
	assert.Len(t, a.shoppingLists[1].Items, 1)
	assert.Equal(t, "new shopping list", a.shoppingLists[1].Name)
	assert.Equal(t, "new item", a.shoppingLists[1].Items[0].What)
	assert.False(t, a.shoppingLists[1].Items[0].Checked)

	test.TapCanvas(a.win.Canvas(), fyne.NewPos(61, 580))
	test.AssertRendersToMarkup(t, "shopping_list_menu.markup", a.win.Canvas())

	test.TapCanvas(a.win.Canvas(), fyne.NewPos(0, 0))
	test.AssertRendersToMarkup(t, "shopping_list_menu_disapear.markup", a.win.Canvas())

	test.MoveMouse(a.win.Canvas(), fyne.NewPos(92, 60))
	test.AssertRendersToMarkup(t, "shopping_list_hover.markup", a.win.Canvas())

	test.TapCanvas(a.win.Canvas(), fyne.NewPos(144, 60))
	test.AssertRendersToMarkup(t, "shopping_list_delete.markup", a.win.Canvas())

	assert.Len(t, a.shoppingLists, 1)
	assert.Len(t, a.shoppingLists[0].Items, 2)
}
