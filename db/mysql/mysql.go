package mysql

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	proto "github.com/micro/go-platform/trace/proto"
	"github.com/micro/trace-srv/db"

	_ "github.com/go-sql-driver/mysql"
)

var (
	Url = "root@tcp(127.0.0.1:3306)/trace"

	spanQ = map[string]string{
		"createSpan": `INSERT INTO %s.%s (trace_id, parent_id, span_id, timestamp, duration, debug, source, destination, name)
				values (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"readSpan": `SELECT trace_id, parent_id, span_id, timestamp, duration, debug, source, destination, name 
				from %s.%s where trace_id = ? and timestamp > 0 limit 1`,
		"deleteSpan": "DELETE FROM %s.%s where trace_id = ?",
		"searchAsc": `select trace_id, parent_id, span_id, timestamp, duration, debug, source, destination, name 
				from %s.%s where parent_id = '0' and timestamp > 0 order by timestamp asc limit ? offset ?`,
		"searchDesc": `select trace_id, parent_id, span_id, timestamp, duration, debug, source, destination, name 
				from %s.%s where parent_id = '0' and timestamp > 0 order by timestamp desc limit ? offset ?`,
		"searchNameAsc": `select trace_id, parent_id, span_id, timestamp, duration, debug, source, destination, name 
				from %s.%s where parent_id = '0' and timestamp > 0 and name = ? order by timestamp asc limit ? offset ?`,
		"searchNameDesc": `select trace_id, parent_id, span_id, timestamp, duration, debug, source, destination, name 
				from %s.%s where parent_id = '0' and timestamp > 0 and name = ? order by timestamp desc limit ? offset ?`,
	}

	annQ = map[string]string{
		"createAnn": `INSERT INTO %s.%s (span_id, trace_id, timestamp, type, akey, value, debug, service) 
				values (?, ?, ?, ?, ?, ?, ?, ?)`,
		"readAnn":   `SELECT span_id, trace_id, timestamp, type, akey, value, debug, service from %s.%s where span_id = ?`,
		"deleteAnn": "DELETE FROM %s.%s where trace_id = ?",
	}

	st = map[string]*sql.Stmt{}
)

type mysql struct {
	db *sql.DB
}

func init() {
	db.Register(new(mysql))
}

func (m *mysql) Init() error {
	var d *sql.DB
	var err error

	parts := strings.Split(Url, "/")
	if len(parts) != 2 {
		return errors.New("Invalid database url")
	}

	if len(parts[1]) == 0 {
		return errors.New("Invalid database name")
	}

	url := parts[0]
	database := parts[1]

	if d, err = sql.Open("mysql", url+"/"); err != nil {
		return err
	}
	if _, err := d.Exec("CREATE DATABASE IF NOT EXISTS " + database); err != nil {
		return err
	}
	d.Close()
	if d, err = sql.Open("mysql", Url); err != nil {
		return err
	}
	if _, err = d.Exec(spanSchema); err != nil {
		return err
	}
	if _, err = d.Exec(annSchema); err != nil {
		return err
	}

	for query, statement := range spanQ {
		prepared, err := d.Prepare(fmt.Sprintf(statement, database, "spans"))
		if err != nil {
			return err
		}
		st[query] = prepared
	}

	for query, statement := range annQ {
		prepared, err := d.Prepare(fmt.Sprintf(statement, database, "annotations"))
		if err != nil {
			return err
		}
		st[query] = prepared
	}

	m.db = d

	return nil
}

func (m *mysql) Create(span *proto.Span) error {
	var source, destination string
	b, _ := json.Marshal(span.Source)
	source = string(b)
	b, _ = json.Marshal(span.Destination)
	destination = string(b)

	_, err := st["createSpan"].Exec(span.TraceId, span.ParentId, span.Id, span.Timestamp, span.Duration, span.Debug, source, destination, span.Name)
	if err != nil {
		return err
	}

	for _, ann := range span.Annotations {
		var service, debug string
		b, _ := json.Marshal(ann.Service)
		service = string(b)
		b, _ = json.Marshal(ann.Debug)
		debug = string(b)

		_, err := st["createAnn"].Exec(span.Id, span.TraceId, ann.Timestamp, ann.Type, ann.Key, ann.Value, debug, service)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *mysql) Delete(traceId string) error {
	if len(traceId) == 0 {
		return errors.New("Invalid trace id")
	}

	_, err := st["deleteSpan"].Exec(traceId)
	if err != nil {
		return err
	}

	_, err = st["deleteAnn"].Exec(traceId)
	return err
}

func (m *mysql) Read(traceId string) ([]*proto.Span, error) {
	if len(traceId) == 0 {
		return nil, errors.New("Invalid trace id")
	}

	r, err := st["readSpan"].Query(traceId)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var spans []*proto.Span

	for r.Next() {
		span := &proto.Span{}
		var source, dest string
		if err := r.Scan(&span.TraceId, &span.ParentId, &span.Id, &span.Timestamp, &span.Duration, &span.Debug, &source, &dest, &span.Name); err != nil {
			if err == sql.ErrNoRows {
				return spans, nil
			}
			return nil, err
		}
		if err := json.Unmarshal([]byte(source), &span.Source); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(dest), &span.Destination); err != nil {
			return nil, err
		}

		anns, err := m.ReadAnnotations(span.Id)
		if err != nil {
			return nil, err
		}
		span.Annotations = anns
		spans = append(spans, span)

	}
	if r.Err() != nil {
		return nil, err
	}
	return spans, nil
}

func (m *mysql) ReadAnnotations(spanId string) ([]*proto.Annotation, error) {
	if len(spanId) == 0 {
		return nil, errors.New("Invalid span id")
	}

	r, err := st["readAnn"].Query(spanId)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var anns []*proto.Annotation

	for r.Next() {
		ann := &proto.Annotation{}
		var debug, service string
		var id, tid string
		if err := r.Scan(&id, &tid, &ann.Timestamp, &ann.Type, &ann.Key, &ann.Value, &debug, &service); err != nil {
			if err == sql.ErrNoRows {
				return anns, nil
			}
			return nil, err
		}
		if err := json.Unmarshal([]byte(debug), &ann.Debug); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(service), &ann.Service); err != nil {
			return nil, err
		}

		anns = append(anns, ann)

	}
	if r.Err() != nil {
		return nil, err
	}
	return anns, nil
}

func (m *mysql) Search(name string, limit, offset int64, reverse bool) ([]*proto.Span, error) {
	var r *sql.Rows
	var err error
	query := "search"
	order := "Asc"

	if reverse {
		order = "Desc"
	}

	if len(name) > 0 {
		query = "searchName"
		r, err = st[query+order].Query(name, limit, offset)
	} else {
		r, err = st[query+order].Query(limit, offset)
	}
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var spans []*proto.Span

	for r.Next() {
		span := &proto.Span{}
		var source, dest string
		if err := r.Scan(&span.TraceId, &span.ParentId, &span.Id, &span.Timestamp, &span.Duration, &span.Debug, &source, &dest, &span.Name); err != nil {
			if err == sql.ErrNoRows {
				return spans, nil
			}
			return nil, err
		}
		if err := json.Unmarshal([]byte(source), &span.Source); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(dest), &span.Destination); err != nil {
			return nil, err
		}
		spans = append(spans, span)

	}
	if r.Err() != nil {
		return nil, err
	}
	return spans, nil
}
