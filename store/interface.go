package store

import (
	"Akso/meta"
)

type storeManager interface {
	ReadStore() (*meta.MaterialStore, error)
	WriteStore(*meta.MaterialStore) error
}

type materialManager interface {
	GetMaterial(materialName string) (*meta.Material, error)
	CreateMaterial(meterial *meta.Material) error
}

type FoodStore interface {
	storeManager
	materialManager
}
