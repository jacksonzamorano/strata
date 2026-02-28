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
	GetRecentTaskRuns(limit int) []TaskRun
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

func (s *SQLiteStorage) GetRecentTaskRuns(limit int) []TaskRun {
	if limit <= 0 {
		limit = 200
	}

	rows, err := s.db.Query(
		`SELECT task_run.id,
		        task_run.succeeded,
		        task_run.input_body,
		        task_run.input_query,
		        task_run.input_headers,
		        task_run.output,
		        task_run.task_start_date,
		        task_run.task_end_date
		   FROM task_run
		  ORDER BY task_run.task_start_date DESC
		  LIMIT ?1`,
		limit,
	)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	out := make([]TaskRun, 0)
	for rows.Next() {
		var entry TaskRun
		if scanErr := ScanTaskRun(rows, &entry); scanErr != nil {
			panic(scanErr)
		}
		out = append(out, entry)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		panic(rowsErr)
	}
	return out
}
