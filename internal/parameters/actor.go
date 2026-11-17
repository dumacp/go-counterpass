package parameters

import (
	"time"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/dumacp/go-counterpass/internal/pubsub"
	"github.com/dumacp/go-gwiot/pkg/gwiot"
	"github.com/dumacp/go-params/pkg/params"
)

func NewActor(id, databasepath string, timeout time.Duration) actor.Actor {
	a := params.Actor(id, databasepath, false, params.NatsActor(id, gwiot.NewDiscoveryActor("counterpassdiscovery/params",
		pubsub.Subscribe,
		pubsub.Publish)))
	return a
}
