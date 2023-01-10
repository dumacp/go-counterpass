module github.com/dumacp/go-counterpass

go 1.16

replace github.com/dumacp/go-levis => ../go-levis

replace github.com/dumacp/go-actors => ../go-actors

replace github.com/dumacp/go-logs => ../go-logs

replace github.com/dumacp/gpsnmea => ../gpsnmea

replace github.com/dumacp/pubsub => ../pubsub

replace github.com/dumacp/go-ingnovus => ../go-ingnovus

replace github.com/dumacp/go-optocontrol => ../go-optocontrol

replace github.com/dumacp/go-logirastreo => ../go-logirastreo

replace github.com/dumacp/sonar => ../sonar

require (
	github.com/AsynkronIT/protoactor-go v0.0.0-20220214042420-fcde2cd4013e
	github.com/brian-armstrong/gpio v0.0.0-20181227042754-72b0058bbbcb
	github.com/dumacp/go-actors v0.0.0-20210923182122-b64616cc9d17
	github.com/dumacp/go-ingnovus v0.0.0-00010101000000-000000000000
	github.com/dumacp/go-levis v0.0.0-00010101000000-000000000000
	github.com/dumacp/go-logirastreo v0.0.0-00010101000000-000000000000
	github.com/dumacp/go-logs v0.0.0-20211122205852-dfcc5685457f
	github.com/dumacp/go-optocontrol v0.0.0-00010101000000-000000000000
	github.com/dumacp/gpsnmea v0.0.0-00010101000000-000000000000
	github.com/dumacp/pubsub v0.0.0-20200115200904-f16f29d84ee0
	github.com/dumacp/sonar v0.0.0-00010101000000-000000000000
	github.com/eclipse/paho.mqtt.golang v1.3.5
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.5.2
	github.com/looplab/fsm v0.3.0
	golang.org/x/net v0.0.0-20220520000938-2e3eb7b945c2 // indirect
)
