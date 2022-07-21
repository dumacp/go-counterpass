package app

import (
	"container/list"
	"fmt"
	"strings"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-counterpass/internal/pubsub"
	"github.com/dumacp/go-counterpass/messages"
	"github.com/dumacp/go-logs/pkg/logs"
)

const (
	timeoutQueue = 2 * time.Second
)

type QueueActor struct {
	queue *list.List
	quit  chan int
}

func NewQueueActor() actor.Actor {
	return &QueueActor{}
}

type MsgQueueEvent struct {
	Event     []byte
	Type      messages.Event_EventType
	ID        int32
	Timestamp time.Time
}

func (a *QueueActor) Receive(ctx actor.Context) {
	// fmt.Printf("actor \"%s\", message: %v, %T\n", ctx.Self().GetId(), ctx.Message(), ctx.Message())
	switch msg := ctx.Message().(type) {
	case *actor.Started:
		a.queue = list.New()
		if a.quit != nil {
			select {
			case <-a.quit:
			default:
				close(a.quit)
			}
		}
		a.quit = make(chan int)
		go funcTick(ctx.ActorSystem().Root, ctx.Self(), 500*time.Millisecond, a.quit)
		logs.LogInfo.Printf("actor started \"%s\"", ctx.Self().Id)
	case *actor.Stopping:
		if a.quit != nil {
			select {
			case <-a.quit:
			default:
				close(a.quit)
			}
		}
	case *MsgTickQueue:
		el := a.queue.Back()
		for {
			if el == nil {
				break
			}
			elm := el
			el = elm.Prev()
			switch evt := elm.Value.(type) {
			case *MsgQueueEvent:
				if time.Since(evt.Timestamp) >= 2*time.Second {
					pubsub.Publish(topicCounterEvent, evt.Event)
					fmt.Println("remove 0 message")
					a.queue.Remove(elm)
				}
			}
		}
	case *MsgQueueEvent:
		// fmt.Printf("actor \"%s\", message: %s, %T\n", ctx.Self().GetId(), ctx.Message(), ctx.Message())
		if !strings.Contains(VendorCounter, "beane") {
			pubsub.Publish(topicCounterEvent, msg.Event)
			break
		}
		if err := func() error {
			el := a.queue.Back()
			if el == nil {
				fmt.Println("pushback 0 message")
				a.queue.PushBack(msg)
				return nil
			}
			switch evt := el.Value.(type) {
			case *MsgQueueEvent:
				if evt.ID == msg.ID &&
					msg.Timestamp.Sub(evt.Timestamp) < timeoutQueue &&
					evt.Type != msg.Type {
					fmt.Println("remove 1 message")
					a.queue.Remove(el)
					return fmt.Errorf("event before 2 secs, prev: %s; current: %s", evt.Event, msg.Event)
				} else {
					fmt.Println("pushback 1 message")
					a.queue.PushBack(msg)
				}
			default:
			}
			return nil
		}(); err != nil {
			logs.LogWarn.Println(err)
		}
	}
}

type MsgTickQueue struct{}

func funcTick(ctx *actor.RootContext, pid *actor.PID, timeout time.Duration, quit <-chan int) {

	tick := time.NewTicker(timeout)
	defer tick.Stop()
	for {
		select {
		case <-quit:
			return
		case <-tick.C:
			ctx.Send(pid, &MsgTickQueue{})
		}
	}
}
