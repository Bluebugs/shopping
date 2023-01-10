package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"

	"fyne.io/fyne/v2/storage"
	"go.etcd.io/bbolt"
)

func (a *appData) saveShoppingList(key uint64) error {
	return a.db.Update(func(tx *bbolt.Tx) error {
		return saveShoppingListInTx(tx, key, a.shoppingLists[key])
	})
}

func saveShoppingListInTx(tx *bbolt.Tx, key uint64, sl *shoppingList) error {
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
	return b.Put(binary.BigEndian.AppendUint64([]byte{}, key), buf.Bytes())
}

func (a *appData) deleteShoppingList(key uint64) error {
	return a.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("shoppingLists"))
		if b == nil {
			return fmt.Errorf("bucket not found")
		}
		return b.Delete(binary.BigEndian.AppendUint64([]byte{}, key))
	})
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

			a.shoppingLists[binary.BigEndian.Uint64(k)] = &sl
			return nil
		})
	})

	return nil
}

func (a *appData) Close() error {
	return a.db.Close()
}

func (a *appData) newTuppleKeyShoppingList(name string) (*shoppingList, uint64, error) {
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

		a.shoppingLists[key] = newShoppingList

		return saveShoppingListInTx(tx, key, newShoppingList)
	})
	if err != nil {
		return nil, 0, err
	}
	return newShoppingList, key, nil
}
