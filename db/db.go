package db

import (
	proto "github.com/micro/go-platform/trace/proto"
)

var (
	db DB
)

type DB interface {
	Init() error
	Create(span *proto.Span) error
	Read(traceId string) ([]*proto.Span, error)
	Delete(traceId string) error
	Search(name string, limit, offset int64, reverse bool) ([]*proto.Span, error)
}

func Register(backend DB) {
	db = backend
}

func Init() error {
	return db.Init()
}

func Create(span *proto.Span) error {
	return db.Create(span)
}

func Read(traceId string) ([]*proto.Span, error) {
	return db.Read(traceId)
}

func Delete(traceId string) error {
	return db.Delete(traceId)
}

func Search(name string, limit, offset int64, reverse bool) ([]*proto.Span, error) {
	return db.Search(name, limit, offset, reverse)
}
