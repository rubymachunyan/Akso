package store

import (
	"Akso/meta"
	//"os"
	"testing"
)

func TestWriteReadStore(t *testing.T) {
	store, err := NewStore()
	if err != nil {
		t.Fatal("fail to new file store:", err)
	}

	material := meta.Material{Name: "apple", Type: "Fruit", Description: "a kind of fruit", Tags: "vc, fruit"}
	err = store.CreateMaterial(&material)
	if err != nil {
		t.Fatal("fail to write service:", err)
	}
	m, errr := store.GetMaterial("apple")
	if errr != nil {
		t.Fatal("fail to read service:", err)
	}
	if "apple" != m.Name {
		t.Error("app name should be apple")
	}

	material = meta.Material{Name: "orange", Type: "Fruit", Description: "a kind of fruit", Tags: "vc, fruit"}
	err = store.CreateMaterial(&material)

}
