package doors

import (
	"log"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/eventstream"
	"github.com/dumacp/go-counterpass/utils"
)

const (
	GpioPuerta1 uint = 162
	GpioPuerta2 uint = 163
)

//DoorsActor type
type DoorsActor struct {
	ctx  actor.Context
	evts *eventstream.EventStream
	quit chan int
}

func New() *DoorsActor {
	a := &DoorsActor{}
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

//Receive function to actor Receive messages
func (act *DoorsActor) Receive(ctx actor.Context) {
	act.ctx = ctx
	switch ctx.Message().(type) {
	case *actor.Started:
		log.Printf("actor started \"%s\"", ctx.Self().Id)
		if act.quit != nil {
			select {
			case _, ok := <-act.quit:
				if ok {
					close(act.quit)
				}
			default:
				close(act.quit)
			}
			time.Sleep(300 * time.Millisecond)
		}
		act.quit = make(chan int)
		act.listenGpio(act.quit)
	case *msgGpioError:
		time.Sleep(3 * time.Second)
		panic("error gpio")
	case *actor.Stopping:
		select {
		case act.quit <- 1:
		case <-time.After(1 * time.Second):
		}
	case *Subscribe:
		if ctx.Sender() != nil {
			subscribe(ctx, act.evts)
		}
	}
}

type MsgDoor struct {
	ID    uint
	Value uint
}

type msgGpioError struct{}

func (act *DoorsActor) listenGpio(quit chan int) {
	chPuertas, err := utils.GpioNewWatcher(quit, GpioPuerta1, GpioPuerta2)
	if err != nil {
		log.Println(err)
		log.Panic(err)
	}
	go func() {
		defer func() {
			act.ctx.Send(act.ctx.Self(), &msgGpioError{})
		}()
		for {
			select {
			case <-quit:
				return
			case v := <-chPuertas:
				if act.evts != nil {
					act.evts.Publish(&MsgDoor{ID: v[0], Value: v[1]})
				}
			}
		}

	}()
}
