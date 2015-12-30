# Trace Server

Trace server is a distributed tracing storage system. It's built for the micro ecosystem and to be used with [go-platform/trace](https:/github.com/micro/go-platform/trace)

## Getting started

1. Install Consul

	Consul is the default registry/discovery for go-micro apps. It's however pluggable.
	[https://www.consul.io/intro/getting-started/install.html](https://www.consul.io/intro/getting-started/install.html)

2. Run Consul
	```
	$ consul agent -server -bootstrap-expect 1 -data-dir /tmp/consul
	```

3. Download and start the service

	```shell
	go get github.com/micro/trace-srv
	trace-srv --database_url="root:root@tcp(192.168.99.100:3306)/trace"
	```

	OR as a docker container

	```shell
	docker run microhq/trace-srv --database_url="root:root@tcp(192.168.99.100:3306)/user" --registry_address=YOUR_REGISTRY_ADDRESS
	```

## The Trace API

### Read Trace
```shell
micro query go.micro.srv.trace Trace.Read '{"id": "c45ab444-ae8e-11e5-b22a-68a86d0d36b6"}'
{
	"spans": [
		{
			"annotations": [
				{
					"timestamp": 1.451436390757642e+15,
					"type": 7
				},
				{
					"timestamp": 1.451436390757735e+15,
					"type": 8
				},
				{
					"timestamp": 1.451436390753609e+15,
					"type": 1
				},
				{
					"timestamp": 1.451436390753621e+15,
					"type": 4
				},
				{
					"timestamp": 1.451436390758566e+15,
					"type": 5
				},
				{
					"timestamp": 1.451436390758568e+15,
					"type": 2
				}
			],
			"debug": true,
			"duration": 4964,
			"id": "c45ab444-ae8e-11e5-b22a-68a86d0d36b6",
			"name": "go.micro.srv.example.Example.Call",
			"parent_id": "0",
			"timestamp": 1.451436390753604e+15,
			"trace_id": "c45ab48a-ae8e-11e5-b22a-68a86d0d36b6"
		}
	]
}
```

## Sending to Trace

The trace server consumes messages from the broker topic micro.trace.span. Traces can be generated and sent with the go-platform.
