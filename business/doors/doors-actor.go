package doors

import (
	"log"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-counterpass/utils"
)

const (
	GpioPuerta1 uint = 162
	GpioPuerta2 uint = 163
)

//DoorsActor type
type DoorsActor struct {
	ctx  actor.Context
	quit chan int
}

//Receive function to actor Receive messages
func (act *DoorsActor) Receive(ctx actor.Context) {
	act.ctx = ctx
	switch ctx.Message().(type) {
	case *actor.Started:
		log.Printf("actor started \"%s\"", ctx.Self().Id)
		act.quit = make(chan int, 0)
		act.listenGpio(act.quit)
	case *msgGpioError:
		panic("error gpio")
	case *actor.Stopping:
		select {
		case act.quit <- 1:
		case <-time.After(1 * time.Second):
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
		for v := range chPuertas {
			act.ctx.Send(act.ctx.Parent(), &MsgDoor{ID: v[0], Value: v[1]})
		}

		act.ctx.Send(act.ctx.Self(), &msgGpioError{})
	}()
}
