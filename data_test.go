package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.etcd.io/bbolt"
)

func Test_saveLoad(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test.db")
	if err != nil {
		panic(err)
	}
	tempFile.Close()
	os.Remove(tempFile.Name())
	db, err := bbolt.Open(tempFile.Name(), 0600, nil)
	if err != nil {
		panic(err)
	}
	defer os.Remove(tempFile.Name())

	a := &appData{db: db}
	defer a.Close()

	sl, err := a.newShoppingList("test")
	assert.Nil(t, err)
	assert.NotNil(t, sl)
	assert.Equal(t, "test", sl.Name)

	sl.Items = []item{
		{
			What:    "unchecked",
			Checked: false,
		},
		{
			What:    "checked",
			Checked: true,
		},
	}

	err = a.saveShoppingList(sl)
	assert.Nil(t, err)

	slEmpty, err := a.newShoppingList("test empty")
	assert.Nil(t, err)
	assert.NotNil(t, slEmpty)
	assert.Equal(t, "test empty", slEmpty.Name)

	a.shoppingLists = []*shoppingList{}

	err = a.loadShoppingListsFromDB()
	assert.Nil(t, err)

	assert.Equal(t, 2, len(a.shoppingLists))
	assert.Equal(t, "test", a.shoppingLists[0].Name)
	assert.Equal(t, "test empty", a.shoppingLists[1].Name)
	assert.Equal(t, 2, len(a.shoppingLists[0].Items))
	assert.Equal(t, 0, len(a.shoppingLists[1].Items))
	assert.Equal(t, "unchecked", a.shoppingLists[0].Items[0].What)
	assert.Equal(t, false, a.shoppingLists[0].Items[0].Checked)
	assert.Equal(t, "checked", a.shoppingLists[0].Items[1].What)
	assert.Equal(t, true, a.shoppingLists[0].Items[1].Checked)
}
