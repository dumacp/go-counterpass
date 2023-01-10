package listen

import (
	"fmt"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/eventstream"
	"github.com/dumacp/go-counterpass/internal/device"
	"github.com/dumacp/go-counterpass/messages"
	"github.com/dumacp/go-logs/pkg/logs"
)

//ListenActor actor to listen events
type ListenActor struct {
	context     actor.Context
	device      interface{}
	test        bool
	typeCounter int
	quit        chan int

	sendConsole    bool
	evts           *eventstream.EventStream
	isListennerOk  bool
	lastTryConnect time.Time
}

//NewListen create listen actor
func NewListen(typeCounter int) actor.Actor {
	act := &ListenActor{}
	act.typeCounter = typeCounter
	act.evts = eventstream.NewEventStream()
	return act
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
		case *MsgListenError, *MsgListenStarted,
			*messages.Event, *device.CloseDevice, *device.StartDevice, *device.StopDevice:
			return true
		}
		return false
	})
}

func (act *ListenActor) SendToConsole(send bool) {
	act.sendConsole = send
}

func (act *ListenActor) Test(test bool) {
	act.test = test
}

//Receive func Receive in actor
func (a *ListenActor) Receive(ctx actor.Context) {
	fmt.Printf("actor \"%s\", message: %v, %T\n", ctx.Self().GetId(),
		ctx.Message(), ctx.Message())
	a.context = ctx
	switch msg := ctx.Message().(type) {
	case *actor.Stopped:
		logs.LogInfo.Printf("actor stopped, reason: %s", msg)
	case *actor.Started:
		// act.initlogs.Logs()
		logs.LogInfo.Printf("actor started \"%s\"", ctx.Self().Id)
		a.lastTryConnect = time.Now().Add(-10 * time.Minute)

	case *actor.Stopping:
		// case *messages.CountingActor:
		// 	act.countingActor = actor.NewPID(msg.Address, msg.ID)
	case *device.MsgDevice:
		if a.quit != nil {
			select {
			case _, ok := <-a.quit:
				if ok {
					close(a.quit)
				}
			default:
				close(a.quit)
			}
			time.Sleep(300 * time.Millisecond)
		}

		a.quit = make(chan int)

		if err := Listen(msg.Device, a.quit, ctx, a.typeCounter, a.sendConsole); err != nil {
			logs.LogError.Println(err)
			time.Sleep(1 * time.Second)
			id := a.typeCounter >> 1
			ctx.Send(ctx.Self(), &MsgListenError{ID: id})
			break
		}
		fmt.Printf("listen(typeCounter=%d) started\n", a.typeCounter)
		a.device = msg.Device
	case *MsgListenError:
		if a.evts != nil {
			a.evts.Publish(&device.StopDevice{})
			a.evts.Publish(&device.StartDevice{})
		}
		if a.isListennerOk {
			a.isListennerOk = false
			if a.evts != nil {
				a.evts.Publish(msg)
			}
		} else if time.Since(a.lastTryConnect) > 3*time.Minute {
			a.lastTryConnect = time.Now()
			if a.evts != nil {
				a.evts.Publish(msg)
			}
		}
		fmt.Printf("listen(id=%d) error\n", msg.ID)
	case *MsgToTest:
		logs.LogBuild.Printf("test frame: %s", msg.Data)
	case *MsgLogRequest:
	case *messages.Event:
		if !a.isListennerOk {
			a.isListennerOk = true
		}
		if a.evts != nil {
			a.evts.Publish(msg)
		}
	case *Subscribe:
		if ctx.Sender() == nil {
			break
		}
		subscribe(ctx, a.evts)
	}
}
