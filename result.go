package scylla

type Result struct {
	lastInsertId int64
	rowsAffected int64
}

func (r Result) LastInsertId() int64 {
	return r.lastInsertId
}

func (r Result) RowsAffected() int64 {
	return r.rowsAffected
}
