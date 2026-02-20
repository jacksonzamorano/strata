package tasklib

import (
	"database/sql"
	"encoding/json"
	"reflect"
)

type ContainerEntityStorage[T any] struct {
	db        *sql.DB
	namespace string
	kind      string
}

func NewEntityStorage[T any](f *ContainerStorage) *ContainerEntityStorage[T] {
	var zero T
	t := reflect.TypeOf(zero)
	return &ContainerEntityStorage[T]{
		db:        f.db,
		namespace: f.namespace,
		kind:      t.String(),
	}
}

type FilterFn[T any] = func(v T) bool

func (s *ContainerEntityStorage[T]) Get(id int64) *T {
	entity, err := GetEntityRow(s.db, s.namespace, s.kind, id)
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
	entities, err := GetInNamespace(s.db, s.namespace, s.kind)
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
	newVal, err := CreateEntityRow(s.db, s.namespace, s.kind, string(encoded))
	return newVal.Id
}

func (s *ContainerEntityStorage[T]) Update(id int64, record T) {
	encoded, err := json.Marshal(record)
	if err != nil {
		return
	}
	_, err = UpdateEntityRow(s.db, s.namespace, s.kind, id, string(encoded))
	if err != nil {
		panic(err)
	}
}

func (s *ContainerEntityStorage[T]) Delete(id int64) {
	DeleteEntityRow(s.db, s.namespace, s.kind, id)
}

func (s *ContainerEntityStorage[T]) DeleteWhere(filter FilterFn[T]) int {
	if filter == nil {
		return 0
	}
	entities, err := GetInNamespace(s.db, s.namespace, s.kind)
	if err != nil {
		panic(err)
	}

	count := 0
	for e := range entities {
		var newVal T
		err := json.Unmarshal([]byte(entities[e].Value), &newVal)
		if err != nil {
			continue
		}
		if filter(newVal) {
			DeleteEntityRow(s.db, s.namespace, s.kind, entities[e].Id)
			count++
		}
	}
	return count
}
