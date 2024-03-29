package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"

	"fyne.io/fyne/v2/storage"
	"github.com/psanford/wormhole-william/wormhole"
	"go.etcd.io/bbolt"
	"gopkg.in/yaml.v2"
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

func (a *appData) loadShoppingListsFromDB() error {
	return a.db.View(func(tx *bbolt.Tx) error {
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

	return a.loadShoppingListsFromDB()
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

func (sl *shoppingList) exportYaml(writer io.WriteCloser) error {
	defer writer.Close()

	// Export shopping list sl to a yaml file
	return yaml.NewEncoder(writer).Encode(sl)
}

func (sl *shoppingList) uploadYaml(ctx context.Context) (string, chan wormhole.SendResult, error) {
	var writer bytes.Buffer
	err := sl.exportYaml(nopWriteCloser{Writer: &writer})
	if err != nil {
		return "", nil, err
	}

	var c wormhole.Client
	code, status, err := c.SendText(ctx, writer.String())
	if err != nil {
		return "", nil, err
	}

	return code, status, err
}

func (sl *shoppingList) importYaml(reader io.ReadCloser) error {
	defer reader.Close()

	// Import shopping list sl from a yaml file
	return yaml.NewDecoder(reader).Decode(sl)
}

func (sl *shoppingList) downloadYaml(ctx context.Context, code string) error {
	var c wormhole.Client

	msg, err := c.Receive(ctx, code)
	if err != nil {
		return err
	}

	if msg.Type != wormhole.TransferText {
		return fmt.Errorf("expected a text message but got type %s", msg.Type)
	}

	err = sl.importYaml(io.NopCloser(msg))
	if err != nil {
		return err
	}
	return nil
}
