package doors

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/eventstream"
	"github.com/dumacp/go-counterpass/internal/pubsub"
	"github.com/dumacp/go-doors/api"
	"github.com/dumacp/go-doors/doorsmsg"
	"github.com/dumacp/go-logs/pkg/logs"
)

// Actor actor to listen events
type Actor struct {
	ctx  actor.Context
	evts *eventstream.EventStream
}

func NewActor() actor.Actor {
	a := &Actor{}
	a.evts = eventstream.NewEventStream()
	return a
}

func subscribe(ctx actor.Context, evs *eventstream.EventStream) {
	if evs == nil {
		return
	}
	rootctx := ctx.ActorSystem().Root
	pid := ctx.Sender()
	self := ctx.Self()

	fn := func(evt interface{}) {
		rootctx.RequestWithCustomSender(pid, evt, self)
	}
	evs.SubscribeWithPredicate(fn, func(evt interface{}) bool {
		switch evt.(type) {
		case *MsgDoor:
			return true
		}
		return false
	})
}

func parseEvents(msg []byte) (interface{}, error) {

	val := &api.ValuePubSubEvent{}

	event := new(api.PubSubEvent)
	event.Value = val

	if err := json.Unmarshal(msg, event); err != nil {
		return nil, err
	}

	result := new(MsgDoor)

	// resVal, ok := event.Value.(*api.ValuePubSubEvent)
	// if !ok {
	// 	return nil, fmt.Errorf("empty value in door message: %v", event)
	// }

	result.ID = uint(val.Id)
	result.Value = uint(val.State)

	return result, nil
}

func parseEvents2(msg []byte) (interface{}, error) {

	val := &api.DoorState{}

	if err := json.Unmarshal(msg, val); err != nil {
		return nil, fmt.Errorf("parse Doorstate error: %v", val)
	}

	result := new(MsgDoor)
	result.ID = uint(val.Id)
	result.Value = uint(val.Value)

	return result, nil
}

// Receive func Receive in actor
func (a *Actor) Receive(ctx actor.Context) {
	logs.LogBuild.Printf("Message arrived in doorsActor: %T, %s",
		ctx.Message(), ctx.Sender())

	a.ctx = ctx
	switch msg := ctx.Message().(type) {
	case *actor.Started:
		logs.LogInfo.Printf("started \"%s\", %v", ctx.Self().GetId(), ctx.Self())
		if err := pubsub.Subscribe("EVENTS/doors", ctx.Self(), parseEvents); err != nil {
			time.Sleep(3 * time.Second)
			logs.LogError.Panic(err)
		}
		if err := pubsub.Subscribe("counterpass/doors", ctx.Self(), parseEvents2); err != nil {
			time.Sleep(3 * time.Second)
			logs.LogError.Panic(err)
		}
		req := doorsmsg.DoorsStateRequest{
			TopicResponse: "counterpass/doors",
		}
		data, err := json.Marshal(req)
		if err != nil {
			logs.LogError.Printf("subscribe error: %s", err)
		} else {
			pubsub.Publish("DOORS", data)
		}
	case *actor.Stopping:
		logs.LogWarn.Printf("\"%s\" - Stopped actor, reason -> %v", ctx.Self(), msg)
	case *actor.Restarting:
		logs.LogWarn.Printf("\"%s\" - Restarting actor, reason -> %v", ctx.Self(), msg)
	case *actor.Terminated:
		logs.LogWarn.Printf("\"%s\" - Terminated actor, reason -> %v", ctx.Self(), msg)
	case *MsgDoor:
		if a.evts != nil {
			a.evts.Publish(msg)
		}
	case *Subscribe:
		if ctx.Sender() != nil {
			subscribe(ctx, a.evts)
		}
	}
}
