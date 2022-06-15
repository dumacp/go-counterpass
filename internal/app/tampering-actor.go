package app

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-counterpass/messages"
	"github.com/dumacp/go-logs/pkg/logs"
)

type tamperingActor struct {
	quit chan int
	evts chan *messages.Event
}

func (a *tamperingActor) Receive(ctx actor.Context) {

	switch msg := ctx.Message().(type) {
	case *actor.Started:
		if a.quit != nil {
			select {
			case _, ok := <-a.quit:
				if ok {
					close(a.quit)
				}
			default:
				close(a.quit)
			}
		}
		a.quit = make(chan int)
		a.evts = tamperingFlipFlop(a.quit)
		logs.LogInfo.Printf("actor started \"%s\"", ctx.Self().Id)

	case *messages.Event:
		if msg.GetType() != messages.TAMPERING {
			break
		}
		select {
		case a.evts <- msg:
		case <-time.After(100 * time.Millisecond):
		}
	}

}

func tamperingFlipFlop(quit <-chan int) chan *messages.Event {
	ch := make(chan *messages.Event)

	type timer struct {
		Timer *time.Timer
		Begin time.Time
		ID    int32
	}

	go func() {
		defer fmt.Println(" ===== END ===== ")
		defer close(ch)

		mem := make(map[int32]*messages.Event)
		ticks := make(map[int]*timer)
		defer func() {
			for _, tick := range ticks {
				if !tick.Timer.Stop() {
					select {
					case <-tick.Timer.C:
					case <-time.After(100 * time.Millisecond):
					}
				}
			}
		}()

		chans := make([]reflect.SelectCase, 0)
		mapChans := make(map[int]*reflect.SelectCase)

		chans = append(chans, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(quit)})
		mapChans[0] = &chans[0]
		chans = append(chans, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)})
		mapChans[1] = &chans[1]
		for {

			n, value, ok := reflect.Select(chans)
			fmt.Printf("n: %d, value: %v, ok: %v\n", n, value, ok)
			if ok {
				if value.IsValid() {
					switch evt := value.Interface().(type) {
					case *messages.Event:
						// fmt.Printf("n: %d, value: %v, ok: %v\n", n, value, ok)
						if evt.GetType() != messages.TAMPERING {
							break
						}
						if _, ok := mem[evt.GetId()]; !ok {
							mem[evt.GetId()] = new(messages.Event)
							*(mem[evt.GetId()]) = *evt
							tick := time.NewTimer(10 * time.Second)

							chans = append(chans, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(tick.C)})
							mapChans[len(chans)-1] = &chans[len(chans)-1]
							ticks[len(chans)-1] = &timer{
								Timer: tick,
								ID:    evt.GetId(),
								Begin: time.Now(),
							}
							break
						}
						if mem[evt.GetId()].GetValue() < evt.GetValue() {
							for _, v := range ticks {
								if v.ID == evt.GetId() {
									if !v.Timer.Stop() {
										select {
										case <-v.Timer.C:
										case <-time.After(100 * time.Millisecond):
										}
									}
									v.Timer.Reset(10 * time.Second)
								}
								break
							}
							mem[evt.GetId()] = evt
						}
					case time.Time:
						if _, ok := mapChans[n]; ok {
							if v, ok := ticks[n]; ok {
								t1 := v.Begin
								delete(mem, v.ID)
								delete(ticks, n)
								log.Printf("send tampering: from -> %s, %d\n", t1, v.ID)
							}
						}
						delete(mapChans, n)
						chans = make([]reflect.SelectCase, 0)
						for _, v := range mapChans {
							chans = append(chans, *v)
						}
					}
				}
				if n == 0 {
					return
				}
			} else if n == 0 {
				return
			}
		}
	}()

	return ch
}
