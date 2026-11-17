module github.com/dumacp/go-counterpass

go 1.25.0

require (
	github.com/asynkron/protoactor-go v0.0.0-20240413045429-76c172a71a16
	github.com/brian-armstrong/gpio v0.0.0-20181227042754-72b0058bbbcb
	github.com/dumacp/go-actors v0.1.0
	github.com/dumacp/go-doors v1.2.1
	github.com/dumacp/go-gwiot v1.2.24-0.20261101194921-f2d0862b04ef
	github.com/dumacp/go-ingnovus v0.0.0-20240615163652-30e11db662ad
	github.com/dumacp/go-levis v0.0.0-20241119224207-a91cacdf55e3
	github.com/dumacp/go-logirastreo v0.0.0-00010101000000-000000000000
	github.com/dumacp/go-logs v0.0.2-0.20241119230451-ac5d49be15ce
	github.com/dumacp/go-optocontrol v0.1.0
	github.com/dumacp/go-params v1.0.4
	github.com/dumacp/gpsnmea v0.0.0-20260401150321-db2499a38f8c
	github.com/dumacp/pubsub v0.0.0-20200115200904-f16f29d84ee0
	github.com/dumacp/sonar v0.1.0
	github.com/eclipse/paho.mqtt.golang v1.4.3
	github.com/looplab/fsm v1.0.2
	google.golang.org/protobuf v1.36.8
)

require (
	github.com/Workiva/go-datastructures v1.1.3 // indirect
	github.com/asynkron/gofun v0.0.0-20220329210725-34fed760f4c2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/boltdb/bolt v1.3.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/coreos/go-oidc v2.2.1+incompatible // indirect
	github.com/dumacp/keycloak v0.0.0-20191212174805-9e9a5c3da24f // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/go-logr/logr v1.3.0 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/goburrow/modbus v0.1.0 // indirect
	github.com/goburrow/serial v0.1.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/lithammer/shortuuid/v4 v4.0.0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/nats-io/nats.go v1.44.0 // indirect
	github.com/nats-io/nkeys v0.4.11 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/orcaman/concurrent-map v1.0.0 // indirect
	github.com/pquerna/cachecontrol v0.1.0 // indirect
	github.com/prometheus/client_golang v1.17.0 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.11.1 // indirect
	github.com/serialx/hashring v0.0.0-20200727003509-22c0c7ab6b1b // indirect
	github.com/tarm/serial v0.0.0-20180830185346-98f6abe2eb07 // indirect
	github.com/twmb/murmur3 v1.1.8 // indirect
	go.etcd.io/bbolt v1.3.10 // indirect
	go.opentelemetry.io/otel v1.21.0 // indirect
	go.opentelemetry.io/otel/exporters/prometheus v0.44.0 // indirect
	go.opentelemetry.io/otel/metric v1.21.0 // indirect
	go.opentelemetry.io/otel/sdk v1.21.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.21.0 // indirect
	go.opentelemetry.io/otel/trace v1.21.0 // indirect
	golang.org/x/crypto v0.41.0 // indirect
	golang.org/x/exp v0.0.0-20231110203233-9a3e6036ecaa // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/oauth2 v0.13.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231002182017-d307bd883b97 // indirect
	google.golang.org/grpc v1.60.1 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
)

replace github.com/asynkron/protoactor-go => github.com/dumacp/protoactor-go v0.0.0-20261101194920-b85f35a0f362

replace github.com/nats-io/nats.go => github.com/dumacp/nats.go v1.44.1-0.20261101194920-9a28c3fb2b88

// Pending an accessible repository for this driver.
replace github.com/dumacp/go-logirastreo => ../go-logirastreo
