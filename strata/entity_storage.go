package strata

import (
	"encoding/json"

	"github.com/jacksonzamorano/tasks/strata/core"
)

type ContainerEntityStorage[T any] struct {
	storage   core.EntityStorageProvider
	namespace string
	kind      string
}

type FilterFn[T any] = func(v T) bool

func (s *ContainerEntityStorage[T]) Get(id int64) *T {
	entity, err := s.storage.Get(s.namespace, s.kind, id)
	if err != nil {
		panic(err)
	}
	if entity == nil {
		return nil
	}
	var newVal T
	err = json.Unmarshal([]byte(entity.Value), &newVal)
	if err != nil {
		return nil
	}
	return &newVal
}

func (s *ContainerEntityStorage[T]) Find(filter FilterFn[T]) []T {
	entities, err := s.storage.All(s.namespace, s.kind)
	if err != nil {
		panic(err)
	}

	newEntities := []T{}
	for e := range entities {
		var newVal T
		err := json.Unmarshal([]byte(entities[e].Value), &newVal)
		if err != nil {
			continue
		}
		if filter != nil && filter(newVal) {
			newEntities = append(newEntities, newVal)
		}
	}
	return newEntities
}

func (s *ContainerEntityStorage[T]) Insert(record T) int64 {
	encoded, err := json.Marshal(record)
	if err != nil {
		return 0
	}
	newVal, err := s.storage.Insert(s.namespace, s.kind, encoded)
	return newVal
}

func (s *ContainerEntityStorage[T]) Update(id int64, record T) {
	encoded, err := json.Marshal(record)
	if err != nil {
		return
	}
	err = s.storage.Update(s.namespace, s.kind, id, encoded)
	if err != nil {
		panic(err)
	}
}

func (s *ContainerEntityStorage[T]) Delete(id int64) {
	s.storage.Delete(s.namespace, s.kind, id)
}

func (s *ContainerEntityStorage[T]) DeleteWhere(filter FilterFn[T]) {
	entities, err := s.storage.All(s.namespace, s.kind)
	if err != nil {
		panic(err)
	}

	for e := range entities {
		var newVal T
		err := json.Unmarshal([]byte(entities[e].Value), &newVal)
		if err != nil {
			continue
		}
		if filter != nil && filter(newVal) {
			s.storage.Delete(s.namespace, s.kind, entities[e].Id)
		}
	}
}
