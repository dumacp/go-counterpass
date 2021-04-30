package listen

import (
	"container/list"
	"fmt"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-counterpass/business/events"
	"github.com/dumacp/go-counterpass/business/ping"
	"github.com/dumacp/go-counterpass/logs"
	"github.com/dumacp/go-counterpass/messages"
)

type MsgToTest struct {
	Data []byte
}

type MsgLogRequest struct{}
type MsgLogResponse struct {
	Value []byte
}

//ListenActor actor to listen events
type ListenActor struct {
	// *logs.Logger
	context actor.Context
	test    bool
	// countingActor *actor.PID
	enters0Before int64
	exits0Before  int64
	locks0Before  int64
	enters1Before int64
	exits1Before  int64
	locks1Before  int64

	quit chan int
	dev  Dev

	socket   string
	baudRate int
	// dev         *contador.Device
	timeFailure int
	// counterType int
	sendConsole bool
	countersMem []int64
	queue       *list.List
}

//NewListen create listen actor
func NewListen(socket string, baudRate int) *ListenActor {
	act := &ListenActor{}
	// act.countingActor = countingActor
	act.socket = socket
	act.baudRate = baudRate
	// act.logs.Logger = &logs.Logger{}
	act.quit = make(chan int, 0)
	act.timeFailure = 3
	act.countersMem = make([]int64, 0)
	act.queue = list.New()
	return act
}

func (act *ListenActor) SendToConsole(send bool) {
	act.sendConsole = send
}

func (act *ListenActor) Test(test bool) {
	act.test = test
}

type Dev interface {
	Listen(quit chan int) chan *Event
	SendData(data []byte)
}

type Event struct {
	Type  messages.Event_EventType
	ID    int
	Value int64
}

//Receive func Receive in actor
func (act *ListenActor) Receive(ctx actor.Context) {
	act.context = ctx
	switch msg := ctx.Message().(type) {
	case *actor.Stopped:
		logs.LogInfo.Printf("actor stopped, reason: %s", msg)
	case *actor.Started:
		// act.initlogs.Logs()
		logs.LogInfo.Printf("actor started \"%s\"", ctx.Self().Id)
		dev, err := NewDev(act.socket, act.baudRate)
		if err != nil {
			time.Sleep(3 * time.Second)
			panic(err)
		}
		act.dev = dev
		go act.runListen(act.quit)
	case *actor.Stopping:
		logs.LogWarn.Println("stopped actor")
		select {
		case act.quit <- 1:
		case <-time.After(3 * time.Second):
		}
	// case *messages.CountingActor:
	// 	act.countingActor = actor.NewPID(msg.Address, msg.ID)
	case *msgListenError:
		ctx.Send(ctx.Parent(), &ping.MsgPingError{})
		time.Sleep(time.Duration(act.timeFailure) * time.Second)
		act.timeFailure = 2 * act.timeFailure
		logs.LogError.Panicln("listen error")
	case *MsgToTest:
		logs.LogBuild.Printf("test frame: %s", msg.Data)
		if act.sendConsole {
			logs.LogBuild.Printf("send to console: %q", msg.Data)
			act.dev.SendData(msg.Data)
		}
	case *MsgLogRequest:
		if act.queue.Len() > 0 {
			e := act.queue.Front()
			ev := e
			for ev != nil {
				if v, ok := ev.Value.([]byte); ok {
					ctx.Send(ctx.Sender(), &MsgLogResponse{Value: v})
				}
				ev = e.Next()
			}
		}
	case *events.MsgGPS:
		// if !act.sendGPS {
		// 	break
		// }
		data := make([]byte, 0)
		data = append(data, []byte(">S;0")...)
		data = append(data, msg.Data[0:len(msg.Data)-3]...)
		data = append(data, []byte(";*")...)
		csum := byte(0)
		for _, v := range data {
			csum ^= v
		}
		data = append(data, []byte(fmt.Sprintf("%02X<", csum))...)
		data = append(data, []byte("\r\n")...)
		logs.LogBuild.Printf("send to console ->, %q", data)
		act.dev.SendData(data)
	}
}

type msgListenError struct{}

func (act *ListenActor) runListen(quit chan int) {
	events := act.dev.Listen(quit)

	if err := func() (err error) {
		defer func() {
			r := recover()
			if r != nil {
				err = fmt.Errorf(fmt.Sprintf("recover Listen -> %v", r))
			}
		}()
		for event := range events {
			logs.LogBuild.Printf("listen event: %#v\n", event)
			switch event.Type {
			case messages.INPUT:
				id := event.ID
				if id == 0x81 {
					enters := event.Value
					if diff := enters - act.enters0Before; diff > 0 {
						act.context.Send(act.context.Parent(), &messages.Event{Id: 0, Type: messages.INPUT, Value: enters})
					}
					act.enters0Before = enters
				} else {
					enters := event.Value
					if diff := enters - act.enters1Before; diff > 0 {
						act.context.Send(act.context.Parent(), &messages.Event{Id: 1, Type: messages.INPUT, Value: enters})
					}
					act.enters1Before = enters
				}
			case messages.OUTPUT:
				id := event.ID
				if id == 0x82 {
					enters := event.Value
					if diff := enters - act.exits0Before; diff > 0 {
						act.context.Send(act.context.Parent(), &messages.Event{Id: 0, Type: messages.OUTPUT, Value: enters})
					}
					act.exits0Before = enters
				} else {
					enters := event.Value
					if diff := enters - act.exits1Before; diff > 0 {
						act.context.Send(act.context.Parent(), &messages.Event{Id: 1, Type: messages.OUTPUT, Value: enters})
					}
					act.exits1Before = enters
				}
			case messages.TAMPERING:
				id := event.ID
				if id == 0x82 {
					enters := event.Value
					if diff := enters - act.locks0Before; diff > 0 {
						act.context.Send(act.context.Parent(), &messages.Event{Id: 0, Type: messages.TAMPERING, Value: enters})
					}
					act.locks0Before = enters
				} else {
					enters := event.Value
					if diff := enters - act.locks1Before; diff > 0 {
						act.context.Send(act.context.Parent(), &messages.Event{Id: 1, Type: messages.TAMPERING, Value: enters})
					}
					act.locks1Before = enters
				}
			}

		}
		return nil
	}(); err != nil {
		logs.LogError.Println(err)
	}

	act.context.Send(act.context.Self(), &msgListenError{})
}
