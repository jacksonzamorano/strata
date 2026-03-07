package core

type EntityStorageProvider interface {
	Get(ns, k string, id int64) (*EntityStorageRecord, error)
	All(ns, k string) ([]EntityStorageRecord, error)
	Insert(ns, k string, payload []byte) (int64, error)
	Update(ns, k string, id int64, payload []byte) error
	Delete(ns, k string, id int64) error
}

type EntityStorageRecord struct {
	Id    int64
	Value []byte
}

func (s *SQLiteStorage) Get(ns, k string, id int64) (*EntityStorageRecord, error) {
	entity, err := GetEntityRow(s.db, ns, k, id)
	if err != nil {
		return nil, err
	}
	if entity == nil {
		return nil, nil
	}
	return &EntityStorageRecord{
		Value: []byte(entity.Value),
		Id:    entity.Id,
	}, nil
}
func (s *SQLiteStorage) All(ns, k string) ([]EntityStorageRecord, error) {
	allRes := []EntityStorageRecord{}

	entities, err := GetInNamespace(s.db, ns, k)
	if err != nil {
		return allRes, err
	}

	for r := range entities {
		allRes = append(allRes, EntityStorageRecord{
			Value: []byte(entities[r].Value),
			Id:    entities[r].Id,
		})
	}

	return allRes, nil
}
func (s *SQLiteStorage) Insert(ns, k string, payload []byte) (int64, error) {
	r, err := CreateEntityRow(s.db, ns, k, string(payload))
	if err != nil {
		return 0, err
	}
	return r.Id, nil
}
func (s *SQLiteStorage) Update(ns, k string, id int64, payload []byte) error {
	_, err := UpdateEntityRow(s.db, ns, k, id, string(payload))
	return err
}
func (s *SQLiteStorage) Delete(ns, k string, id int64) error {
	_, err := DeleteEntityRow(s.db, ns, k, id)
	return err
}
