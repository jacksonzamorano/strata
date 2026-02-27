package core

type KVStorageProvider interface {
	GetValue(namespace, key string) (string, bool)
	InsertValue(namespace, key, value string) error
	UpdateValue(namespace, key, value string) error
}

func (s *SQLiteStorage) GetValue(namespace, key string) (string, bool) {
	row, err := GetStorageRow(s.db, namespace, key)
	if err != nil {
		panic(err)
	}
	if row == nil {
		return "", false
	}
	return row.Value, true
}
func (s *SQLiteStorage) InsertValue(namespace, key, value string) error {
	_, err := CreateStorageRow(s.db, namespace, key, value)
	return err
}
func (s *SQLiteStorage) UpdateValue(namespace, key, value string) error {
	_, err := UpdateStorageRow(s.db, namespace, key, value)
	return err
}
