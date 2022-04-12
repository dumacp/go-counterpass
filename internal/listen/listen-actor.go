package listen

import (
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/eventstream"
	"github.com/dumacp/go-counterpass/internal/device"
	"github.com/dumacp/go-counterpass/messages"
	"github.com/dumacp/go-logs/pkg/logs"
)

//ListenActor actor to listen events
type ListenActor struct {
	// *logs.Logger
	context     actor.Context
	device      interface{}
	test        bool
	typeCounter int

	quit chan int

	// socket   string
	// baudRate int
	// timeFailure int
	// counterType int
	sendConsole bool
	// countersMem []int64
	// queue       *list.List
	evts *eventstream.EventStream
}

//NewListen create listen actor
func NewListen(typeCounter int) *ListenActor {
	act := &ListenActor{}
	act.typeCounter = typeCounter
	// act.countingActor = countingActor
	// act.socket = socket
	// act.baudRate = baudRate
	// act.logs.Logger = &logs.Logger{}
	// act.quit = make(chan int, 0)
	// act.timeFailure = 3
	// act.countersMem = make([]int64, 0)
	// act.queue = list.New()
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
		case *MsgListenError, *messages.Event, *device.CloseDevice, *device.StartDevice:
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
	// fmt.Printf("actor \"%s\", message: %v, %T\n", ctx.Self().GetId(),
	// 	ctx.Message(), ctx.Message())
	a.context = ctx
	switch msg := ctx.Message().(type) {
	case *actor.Stopped:
		logs.LogInfo.Printf("actor stopped, reason: %s", msg)
	case *actor.Started:
		// act.initlogs.Logs()
		logs.LogInfo.Printf("actor started \"%s\"", ctx.Self().Id)

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

		if err := Listen(msg.Device, a.quit, ctx, a.typeCounter); err != nil {
			logs.LogError.Println(err)
			break
		}
		a.device = msg.Device
	case *MsgListenError:
		if a.evts != nil {
			a.evts.Publish(msg)
			a.evts.Publish(&device.StopDevice{})
			a.evts.Publish(&device.StartDevice{})
		}
		// ctx.Send(ctx.Parent(), &msgPingError{})
		// time.Sleep(3 * time.Second)
		logs.LogError.Println("listen error")
	case *MsgToTest:
		logs.LogBuild.Printf("test frame: %s", msg.Data)
	case *MsgLogRequest:
	case *messages.Event:
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
