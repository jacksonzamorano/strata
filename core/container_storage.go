package core

import (
	"strconv"
	"time"
)

type ContainerStorage struct {
	db        KVStorageProvider
	namespace string
	keys      map[string]struct{}
}

func (cs *ContainerStorage) getValue(key string) string {
	val, ok := cs.db.GetValue(cs.namespace, key)
	if ok {
		cs.keys[key] = struct{}{}
	}
	return val
}

func (cs *ContainerStorage) setValue(key string, value string) error {
	var err error
	if _, ok := cs.keys[key]; ok {
		err = cs.db.UpdateValue(cs.namespace, key, value)
	} else {
		err = cs.db.InsertValue(cs.namespace, key, value)
		if err == nil {
			cs.keys[key] = struct{}{}
		}
	}
	return err
}

func (cs *ContainerStorage) GetString(key string) string {
	return cs.getValue(key)
}

func (cs *ContainerStorage) SetString(key string, value string) error {
	return cs.setValue(key, value)
}

func (cs *ContainerStorage) GetInt(key string) int {
	val := cs.getValue(key)
	pVal, err := strconv.Atoi(val)
	if err != nil {
		return 0
	}
	return pVal
}
func (cs *ContainerStorage) SetInt(key string, value int) error {
	return cs.setValue(key, strconv.Itoa(value))
}

func (cs *ContainerStorage) GetFloat(key string) float64 {
	val := cs.getValue(key)
	pVal, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0
	}
	return pVal
}

func (cs *ContainerStorage) GetBool(key string) bool {
	val := cs.getValue(key)
	pVal, err := strconv.ParseBool(val)
	if err != nil {
		return false
	}
	return pVal
}

func (cs *ContainerStorage) GetDate(key string) time.Time {
	val := cs.getValue(key)
	pVal, err := time.Parse(time.RFC3339, val)
	if err != nil {
		return time.Time{}
	}
	return pVal
}

func (cs *ContainerStorage) SetDate(key string, value time.Time) error {
	return cs.setValue(key, value.Format(time.RFC3339))
}
