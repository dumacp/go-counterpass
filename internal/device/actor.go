package device

import (
	"fmt"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/eventstream"
	"github.com/dumacp/go-logs/pkg/logs"
	"github.com/looplab/fsm"
)

type Actor struct {
	ctx        actor.Context
	portSerial string
	speedBaud  int
	fmachinae  *fsm.FSM
	evts       *eventstream.EventStream
}

func NewActor(port string, speed int, readTimeout time.Duration) actor.Actor {

	a := &Actor{
		portSerial: port,
		speedBaud:  speed,
	}
	a.evts = &eventstream.EventStream{}
	a.Fsm()
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
	evs.SubscribeWithPredicate(fn,
		func(evt interface{}) bool {
			switch evt.(type) {
			case *MsgDevice:
				return true
			}
			return false
		})
}

func (a *Actor) Receive(ctx actor.Context) {
	fmt.Printf("actor \"%s\", message: %v, %T\n", ctx.Self().GetId(),
		ctx.Message(), ctx.Message())
	a.ctx = ctx

	switch msg := ctx.Message().(type) {
	case *actor.Started:
		ctx.Send(ctx.Self(), &StartDevice{})
	case *actor.Stopping:
		a.fmachinae.Event(eError)
	case *StartDevice:
		if err := a.fmachinae.Event(eStarted); err != nil {
			logs.LogError.Printf("open device error: %s", err)
		}
	case *MsgDevice:
		a.fmachinae.Event(eOpenned)
		a.evts.Publish(msg)
	case *StopDevice:
		a.fmachinae.Event(eStop)
	case *Subscribe:
		if ctx.Sender() == nil {
			break
		}
		subscribe(ctx, a.evts)
	}
}
