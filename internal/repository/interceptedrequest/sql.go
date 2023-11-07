package interceptedrequest

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"time"

	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/entity"
)

func init() {
	gob.Register(entity.InterceptedRequestRelevantContent{})
}

type SQLInterceptedRequestRepository struct {
	conn *sql.DB
}

func SQL(db *sql.DB) entity.InterceptedRequestRepository {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS intercepted_request(id TEXT, solved_at TIMESTAMP(3), solved BOOL, req BYTEA, version INTEGER)")
	if err != nil {
		panic(err)
	}
	return &SQLInterceptedRequestRepository{
		conn: db,
	}
}

func (r *SQLInterceptedRequestRepository) Save(req *entity.InterceptedRequest) error {
	tx, err := r.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Commit()

	var request bytes.Buffer
	encoder := gob.NewEncoder(&request)
	if err := encoder.Encode(req.Request); err != nil {
		return err
	}
	query := "INSERT INTO intercepted_request(id, solved_at, solved, req, version) VALUES($1, $2, $3, $4, $5)"
	_, err = tx.Exec(query, req.ID, nil, req.Solved, request.Bytes(), req.Version)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return nil
}

func (r *SQLInterceptedRequestRepository) SetSolved(reqID string, solvedAt time.Time, solved bool) error {
	tx, err := r.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Commit()

	query := "UPDATE intercepted_request SET solved_at=$1, solved=$2 WHERE id=$3"
	_, err = tx.Exec(query, solvedAt, solved, reqID)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return nil
}

func (r *SQLInterceptedRequestRepository) GetLastRequestSolved() (*entity.InterceptedRequest, error) {
	stmt, err := r.conn.Prepare("SELECT id, solved_at, solved, req FROM intercepted_request, version ORDER BY solved_at DESC LIMIT 1")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var req entity.InterceptedRequest
	row := stmt.QueryRow()
	if err := row.Scan(&req.ID, &req.SolvedAt, &req.Solved, &req.Request, &req.Version); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &req, nil
}

func (r *SQLInterceptedRequestRepository) GetAll() ([]*entity.InterceptedRequest, error) {
	stmt, err := r.conn.Prepare("SELECT id, solved_at, solved, req, version FROM intercepted_request ORDER BY solved_at")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var requests []*entity.InterceptedRequest
	rows, err := stmt.Query()
	if err != nil {
		if err == sql.ErrNoRows {
			return []*entity.InterceptedRequest{}, nil
		}
		return nil, err
	}

	for rows.Next() {
		var req entity.InterceptedRequest
		var request []byte
		if err := rows.Scan(&req.ID, &req.SolvedAt, &req.Solved, &request, &req.Version); err != nil {
			return nil, err
		}
		buffer := bytes.NewBuffer(request)
		var decodedRequest entity.InterceptedRequestRelevantContent
		dec := gob.NewDecoder(buffer)
		if err := dec.Decode(&decodedRequest); err == nil {
			req.Request = decodedRequest
		}

		requests = append(requests, &req)
	}

	return requests, nil
}

func (r *SQLInterceptedRequestRepository) GetLastVersion() (int, error) {
	var version int

	stmt, err := r.conn.Prepare("SELECT version FROM intercepted_request ORDER BY version DESC")
	if err != nil {
		return version, err
	}
	defer stmt.Close()

	row := stmt.QueryRow()
	err = row.Scan(&version)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return version, err
}

func (r *SQLInterceptedRequestRepository) GetAllFromLastVersion(version int) ([]*entity.InterceptedRequest, error) {
	stmt, err := r.conn.Prepare("SELECT id, solved_at, solved, req, version FROM intercepted_request WHERE version >= $1 ORDER BY version ASC")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var requests []*entity.InterceptedRequest
	rows, err := stmt.Query(version)
	if err != nil {
		if err == sql.ErrNoRows {
			return []*entity.InterceptedRequest{}, nil
		}
		return nil, err
	}

	for rows.Next() {
		var req entity.InterceptedRequest
		var request []byte
		if err := rows.Scan(&req.ID, &req.SolvedAt, &req.Solved, &request, &req.Version); err != nil {
			return nil, err
		}
		buffer := bytes.NewBuffer(request)
		var decodedRequest entity.InterceptedRequestRelevantContent
		dec := gob.NewDecoder(buffer)
		if err := dec.Decode(&decodedRequest); err == nil {
			req.Request = decodedRequest
		}
		requests = append(requests, &req)
	}

	return requests, nil
}
