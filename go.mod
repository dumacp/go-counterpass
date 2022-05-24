module github.com/dumacp/go-counterpass

go 1.16

replace github.com/dumacp/go-levis => ../go-levis

replace github.com/dumacp/go-actors => ../go-actors

replace github.com/dumacp/go-logs => ../go-logs

replace github.com/dumacp/gpsnmea => ../gpsnmea

replace github.com/dumacp/pubsub => ../pubsub

replace github.com/dumacp/go-ingnovus => ../go-ingnovus

replace github.com/dumacp/go-optocontrol => ../go-optocontrol

replace github.com/dumacp/sonar => ../sonar

require (
	github.com/AsynkronIT/protoactor-go v0.0.0-20220214042420-fcde2cd4013e
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/brian-armstrong/gpio v0.0.0-20181227042754-72b0058bbbcb
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/containerd/continuity v0.3.0 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/dumacp/go-actors v0.0.0-20210923182122-b64616cc9d17
	github.com/dumacp/go-ingnovus v0.0.0-00010101000000-000000000000
	github.com/dumacp/go-levis v0.0.0-00010101000000-000000000000
	github.com/dumacp/go-logs v0.0.0-20211122205852-dfcc5685457f
	github.com/dumacp/go-optocontrol v0.0.0-00010101000000-000000000000
	github.com/dumacp/gpsnmea v0.0.0-00010101000000-000000000000
	github.com/dumacp/pubsub v0.0.0-20200115200904-f16f29d84ee0
	github.com/dumacp/sonar v0.0.0-20211118203617-623170a5174d
	github.com/dumacp/turnstilene v0.0.0-20220118160154-be26dbc3b0cc
	github.com/eclipse/paho.mqtt.golang v1.3.5
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.5.2
	github.com/gotestyourself/gotestyourself v2.2.0+incompatible // indirect
	github.com/lib/pq v0.0.0-20180327071824-d34b9ff171c2 // indirect
	github.com/looplab/fsm v0.3.0
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/opencontainers/runc v1.1.2 // indirect
	github.com/ory/dockertest v3.3.5+incompatible
	github.com/tarm/serial v0.0.0-20180830185346-98f6abe2eb07
	golang.org/x/net v0.0.0-20220520000938-2e3eb7b945c2 // indirect
	golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a // indirect
	gotest.tools v2.2.0+incompatible // indirect
)
