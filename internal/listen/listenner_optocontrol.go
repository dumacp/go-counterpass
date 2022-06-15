//go:build optocontrol
// +build optocontrol

package listen

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-counterpass/messages"
	"github.com/dumacp/go-logs/pkg/logs"
	"github.com/dumacp/go-optocontrol"
)

var timeout_samples int

func init() {
	flag.IntVar(&timeout_samples, "timeout", 900, "timeout samples in millis")
}

const (
	max_error = 3
)

func Listen(dev interface{}, quit <-chan int, ctx actor.Context, typeCounter int, externalConsole bool) error {
	timeout := time.Duration(timeout_samples) * time.Millisecond

	var devv optocontrol.Device

	if v, ok := dev.(optocontrol.Device); !ok {
		return fmt.Errorf("device is not optocontrol device")
	} else {
		devv = v
	}

	inputsFront := uint(0)
	inputsBack := uint(0)
	outputsFront := uint(0)
	outputsBack := uint(0)
	tamperingFront := uint(0)
	tamperingBack := uint(0)
	// anomaliesFront := uint(0)
	// anomaliesBack := uint(0)
	// alarmCacheFront := byte(0x00)
	// alarmCacheBack := byte(0x00)

	self := ctx.Self()
	rootctx := ctx.ActorSystem().Root

	go func(ctx *actor.RootContext, self *actor.PID) {
		defer ctx.Send(self, &MsgListenError{})
		countErr := 0
		var ch1 <-chan time.Time
		var ch2 <-chan time.Time
		if typeCounter == 0 || typeCounter == 1 {
			tick1 := time.NewTicker(timeout)
			defer tick1.Stop()
			ch1 = tick1.C
		} else {
			ch1 = make(<-chan time.Time)
		}
		if typeCounter == 0 {
			time.Sleep(time.Duration(timeout_samples/5) * time.Millisecond)
		}
		if typeCounter == 0 || typeCounter == 2 {
			tick2 := time.NewTicker(timeout)
			defer tick2.Stop()
			ch2 = tick2.C
		} else {
			ch2 = make(<-chan time.Time)
		}
		for {
			select {
			case <-quit:
				logs.LogWarn.Println("device optocontrol is closed")
				return
			case <-ch1:
				tn := time.Now()
				fmt.Printf("%s, request (1)\n", time.Now().Format("02-01-2006 15:04:05.000"))
				id := 0
				result, err := devv.ReadData(optocontrol.DOOR_1)
				if err != nil {
					logs.LogWarn.Println(err)
					if errors.Is(err, io.EOF) {
						if time.Since(tn) < timeout/10 {
							countErr++
						}
					} else {
						countErr++
					}
					log.Println(err)
					if countErr > max_error {
						return
					}
					break
				}
				countErr = 0
				fmt.Printf("%s: result readbytes (1): %+v\n",
					time.Now().Format("02-01-2006 15:04:05.000"), result)

				inputs := result.AdultUp
				outputs := result.AdultDown
				tampering := result.Locks

				// fmt.Printf("inputs (1): %d\n", inputs)
				if inputs > 0 && inputs != inputsFront {
					rootctx.Send(self, &messages.Event{
						Id:    int32(id),
						Value: int64(inputs),
						Type:  messages.INPUT,
					})
				}
				inputsFront = inputs
				if outputs > 0 && outputs != outputsFront {
					rootctx.Send(self, &messages.Event{
						Id:    int32(id),
						Value: int64(outputs),
						Type:  messages.OUTPUT,
					})
				}
				outputsFront = outputs

				if tampering > 0 && tampering != tamperingFront {
					rootctx.Send(self, &messages.Event{
						Id:    int32(id),
						Value: int64(tampering),
						Type:  messages.TAMPERING,
					})
				}
				tamperingFront = tampering

			case <-ch2:
				tn := time.Now()
				fmt.Printf("%s, request (2)\n", time.Now().Format("02-01-2006 15:04:05.000"))
				id := 1
				result, err := devv.ReadData(optocontrol.DOOR_2)
				if err != nil {
					logs.LogWarn.Println(err)
					if errors.Is(err, io.EOF) {
						if time.Since(tn) < timeout/10 {
							countErr++
						}
					} else {
						countErr++
					}
					log.Println(err)
					if countErr > max_error {
						return
					}
					break
				}
				fmt.Printf("%s: result readbytes (2): %+v\n",
					time.Now().Format("02-01-2006 15:04:05.000"), result)

				inputs := result.AdultUp
				outputs := result.AdultDown
				tampering := result.Locks

				// fmt.Printf("inputs (2): %d\n", inputs)
				if inputs > 0 && inputs != inputsBack {
					rootctx.Send(self, &messages.Event{
						Id:    int32(id),
						Value: int64(inputs),
						Type:  messages.INPUT,
					})
				}
				inputsBack = inputs
				if outputs > 0 && outputs != outputsBack {
					rootctx.Send(self, &messages.Event{
						Id:    int32(id),
						Value: int64(outputs),
						Type:  messages.OUTPUT,
					})
				}
				outputsBack = outputs

				if tampering > 0 && tampering != tamperingBack {
					rootctx.Send(self, &messages.Event{
						Id:    int32(id),
						Value: int64(tampering),
						Type:  messages.TAMPERING,
					})
				}
				tamperingBack = tampering
			}
		}
	}(rootctx, self)

	return nil
}
