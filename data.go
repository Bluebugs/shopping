package main

func (a *appData) newShoppingList(name string) (*shoppingList, error) {
	newShoppingList := &shoppingList{Name: name}
	a.shoppingLists = append(a.shoppingLists, newShoppingList)
	return newShoppingList, nil
}

func (a *appData) deleteShoppingList(index int, sl *shoppingList) error {
	if index < len(a.shoppingLists)-1 {
		a.shoppingLists[index] = a.shoppingLists[len(a.shoppingLists)-1]
	}
	a.shoppingLists = a.shoppingLists[:len(a.shoppingLists)-1]
	return nil
}
