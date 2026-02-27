package core

import (
	"encoding/json"
	"time"
)

type TaskHistoryEntry struct {
	Successful   bool
	Input        []byte
	Output       []byte
	QueryParams  map[string][]string
	HeaderParams map[string][]string
	Start        time.Time
	End          time.Time
}
type TaskHistoryProvider interface {
	SaveTaskRun(taskRun TaskHistoryEntry)
}

func (s *SQLiteStorage) SaveTaskRun(taskRun TaskHistoryEntry) {
	query_param, err := json.Marshal(taskRun.QueryParams)
	if err != nil {
		return
	}
	header_param, err := json.Marshal(taskRun.HeaderParams)
	if err != nil {
		return
	}

	op := string(taskRun.Output)

	CreateTaskRun(s.db,
		taskRun.Successful,
		string(taskRun.Input),
		string(query_param),
		string(header_param),
		&op,
		taskRun.Start,
		taskRun.End,
	)
}
