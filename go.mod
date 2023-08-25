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

replace github.com/dumacp/go-doors => ../go-doors

require (
	github.com/AsynkronIT/protoactor-go v0.0.0-20211103040926-5998dd70e1c4
	github.com/brian-armstrong/gpio v0.0.0-20181227042754-72b0058bbbcb
	github.com/dumacp/go-actors v0.0.0-20210923182122-b64616cc9d17
	github.com/dumacp/go-doors v0.0.0-00010101000000-000000000000
	github.com/dumacp/go-ingnovus v0.0.0-00010101000000-000000000000
	github.com/dumacp/go-levis v0.0.0-00010101000000-000000000000
	github.com/dumacp/go-logirastreo v0.0.0-00010101000000-000000000000
	github.com/dumacp/go-logs v0.0.1
	github.com/dumacp/go-optocontrol v0.0.0-00010101000000-000000000000
	github.com/dumacp/gpsnmea v0.0.0-00010101000000-000000000000
	github.com/dumacp/pubsub v0.0.0-20200115200904-f16f29d84ee0
	github.com/dumacp/sonar v0.0.0-00010101000000-000000000000
	github.com/eclipse/paho.mqtt.golang v1.4.2
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.5.2
	github.com/looplab/fsm v1.0.1
)
