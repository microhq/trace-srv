package mysql

var (
	spanSchema = `
CREATE TABLE IF NOT EXISTS spans (
span_id varchar(36) not null,
trace_id varchar(36),
parent_id varchar(36),
timestamp bigint,
duration bigint,
debug boolean,
source text,
destination text,
name varchar(255),
index (trace_id, timestamp),
index (parent_id, timestamp, name),
index(span_id));
`

	annSchema = `
CREATE TABLE IF NOT EXISTS annotations (
span_id varchar(36) not null,
trace_id varchar(36),
timestamp bigint,
type tinyint(1),
akey varchar(255),
value blob,
debug text,
service text,
index(span_id));
`
)
