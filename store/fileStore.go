package store

import (
	"Akso/meta"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

type Store struct {
	FileLocation string
}

func NewStore() (*Store, error) {
	return &Store{FileLocation: "./store.json"}, nil
}

func (s *Store) ReadStore() (*meta.MaterialStore, error) {
	var ms meta.MaterialStore
	raw, err := ioutil.ReadFile(s.FileLocation)
	if err != nil {
		return &ms, err
	}
	if err := json.Unmarshal(raw, &ms); err != nil {
		return &meta.MaterialStore{}, nil
	}
	return &ms, nil
}

func (s *Store) WriteStore(ms *meta.MaterialStore) error {
	jsonFile, err := os.Create(s.FileLocation)
	defer jsonFile.Close()
	if err != nil {
		return err
	}
	msB, err := json.MarshalIndent(ms, "", "    ")
	if err != nil {
		return err
	}
	if _, err := jsonFile.Write(msB); err != nil {
		return err
	}
	return nil
}

func (s *Store) GetMaterial(materialName string) (*meta.Material, error) {
	ms, err := s.ReadStore()
	if err != nil {
		return nil, err
	}
	for _, material := range ms.Materials {
		if materialName == material.Name {
			return material, nil
		}
	}
	return nil, errors.New("Not Found: material" + materialName)
}

func (s *Store) CreateMaterial(material *meta.Material) error {
	ms, err := s.ReadStore()
	if err != nil {
		ms = &meta.MaterialStore{}
	}
	for _, value := range ms.Materials {
		if material.Name == value.Name {
			return errors.New("Material " + material.Name + " already exist")
		}
	}

	ms.Materials = append(ms.Materials, material)
	err = s.WriteStore(ms)

	return err
}
