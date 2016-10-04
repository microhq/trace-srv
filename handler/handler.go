package handler

import (
	"github.com/micro/go-micro/errors"

	proto2 "github.com/micro/go-os/trace/proto"
	"github.com/micro/trace-srv/db"
	proto "github.com/micro/trace-srv/proto/trace"
	"golang.org/x/net/context"
)

type Trace struct{}

func (t *Trace) Read(ctx context.Context, req *proto.ReadRequest, rsp *proto.ReadResponse) error {
	if len(req.Id) == 0 {
		return errors.BadRequest("go.micro.srv.trace.Trace.Read", "invalid trace id")
	}
	spans, err := db.Read(req.Id)
	if err != nil {
		return errors.InternalServerError("go.micro.srv.trace.Trace.Read", err.Error())
	}

	// collapse spans
	spanMap := make(map[string]*proto2.Span)

	for _, span := range spans {
		sp, ok := spanMap[span.Id]
		if !ok {
			spanMap[span.Id] = span
			continue
		}

		if span.Timestamp < sp.Timestamp {
			spanMap[span.Id] = span
		}
	}

	for _, span := range spanMap {
		rsp.Spans = append(rsp.Spans, span)
	}

	return nil
}

func (t *Trace) Create(ctx context.Context, req *proto.CreateRequest, rsp *proto.CreateResponse) error {
	if req.Span == nil {
		return errors.BadRequest("go.micro.srv.trace.Trace.Create", "invalid span")
	}
	err := db.Create(req.Span)
	if err != nil {
		return errors.InternalServerError("go.micro.srv.trace.Trace.Create", err.Error())
	}
	return nil
}

func (t *Trace) Delete(ctx context.Context, req *proto.DeleteRequest, rsp *proto.DeleteResponse) error {
	if len(req.Id) == 0 {
		return errors.BadRequest("go.micro.srv.trace.Trace.Delete", "invalid trace id")
	}
	if err := db.Delete(req.Id); err != nil {
		return errors.InternalServerError("go.micro.srv.trace.Trace.Delete", err.Error())
	}
	return nil
}

func (t *Trace) Search(ctx context.Context, req *proto.SearchRequest, rsp *proto.SearchResponse) error {
	if req.Limit <= 0 {
		req.Limit = 10
	}

	if req.Offset < 0 {
		req.Offset = 0
	}

LOOP:
	for {
		spans, err := db.Search(req.Name, req.Limit, req.Offset, req.Reverse)
		if err != nil {
			return errors.InternalServerError("go.micro.srv.trace.Trace.Search", err.Error())
		}

		// collapse spans
		spanMap := make(map[string]*proto2.Span)

		for _, span := range spans {
			sp, ok := spanMap[span.Id]
			if !ok {
				spanMap[span.Id] = span
				continue
			}

			if span.Timestamp < sp.Timestamp {
				spanMap[span.Id] = span
			}
		}

		for _, span := range spanMap {
			if len(rsp.Spans) == int(req.Limit) {
				break LOOP
			}
			rsp.Spans = append(rsp.Spans, span)
		}

		if len(rsp.Spans) >= int(req.Limit) || len(spans) < int(req.Limit) {
			break
		}
	}

	return nil
}
