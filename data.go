package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"

	"fyne.io/fyne/v2/storage"
	"go.etcd.io/bbolt"
)

func (a *appData) saveShoppingList(sl *shoppingList) error {
	return a.db.Update(func(tx *bbolt.Tx) error {
		return saveShoppingListInTx(tx, sl)
	})
}

func saveShoppingListInTx(tx *bbolt.Tx, sl *shoppingList) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(sl)
	if err != nil {
		return err
	}
	b := tx.Bucket([]byte("shoppingLists"))
	if b == nil {
		return fmt.Errorf("bucket not found")
	}
	return b.Put(binary.BigEndian.AppendUint64([]byte{}, sl.key), buf.Bytes())
}

func (a *appData) loadShoppingLists() error {
	dbURI, err := storage.Child(a.app.Storage().RootURI(), "shopping_list.boltdb")
	if err != nil {
		return err
	}

	a.db, err = bbolt.Open(dbURI.Path(), 0600, nil)
	if err != nil {
		return err
	}

	a.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("shoppingLists"))
		if b == nil {
			return nil
		}

		return b.ForEach(func(k, v []byte) error {
			var sl shoppingList

			buf := bytes.NewBuffer(v)
			dec := gob.NewDecoder(buf)
			err := dec.Decode(&sl)
			if err != nil {
				return err
			}

			sl.key = binary.BigEndian.Uint64(k)

			a.shoppingLists = append(a.shoppingLists, &sl)
			return nil
		})
	})

	return nil
}

func (a *appData) Close() error {
	return a.db.Close()
}

func (a *appData) newShoppingList(name string) (*shoppingList, error) {
	newShoppingList := &shoppingList{Name: name}

	var key uint64
	err := a.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("shoppingLists"))
		if err != nil {
			return err
		}

		key, err = b.NextSequence()
		if err != nil {
			return err
		}

		newShoppingList.key = key
		a.shoppingLists = append(a.shoppingLists, newShoppingList)

		return saveShoppingListInTx(tx, newShoppingList)
	})
	if err != nil {
		return nil, err
	}
	return newShoppingList, nil
}

func (a *appData) deleteShoppingList(index int, sl *shoppingList) error {
	if index < len(a.shoppingLists)-1 {
		a.shoppingLists[index] = a.shoppingLists[len(a.shoppingLists)-1]
	}
	a.shoppingLists = a.shoppingLists[:len(a.shoppingLists)-1]

	return a.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("shoppingLists"))
		if b == nil {
			return fmt.Errorf("bucket not found")
		}
		return b.Delete(binary.BigEndian.AppendUint64([]byte{}, sl.key))
	})
}
