package listen

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-counterpass/messages"
	"github.com/dumacp/go-logs/pkg/logs"
	"github.com/dumacp/go-optocontrol"
)

func ListenOptocontrolV301(dev interface{}, quit <-chan int, ctx actor.Context, typeCounter int, externalConsole bool) error {
	timeout := time.Duration(timeout_samples) * time.Millisecond

	var devv optocontrol.Device_v301

	if v, ok := dev.(optocontrol.Device_v301); !ok {
		return fmt.Errorf("device is not optocontrol-v301 device")
	} else {
		devv = v
	}

	inputsFront := uint(0)
	inputsBack := uint(0)
	outputsFront := uint(0)
	outputsBack := uint(0)
	tamperingFront := uint(0)
	tamperingBack := uint(0)

	self := ctx.Self()
	rootctx := ctx.ActorSystem().Root

	go func(ctx *actor.RootContext, self *actor.PID) {
		firstFrameCh1 := true
		countErr := 0
		var ch1 <-chan time.Time
		tick1 := time.NewTicker(timeout)
		defer tick1.Stop()
		ch1 = tick1.C

		for {
			select {
			case <-quit:
				logs.LogWarn.Println("device optocontrol is closed")
				if typeCounter == 0 || typeCounter == 1 {
					ctx.Send(self, &MsgListenError{ID: 0})
				}
				if typeCounter == 0 || typeCounter == 2 {
					ctx.Send(self, &MsgListenError{ID: 1})
				}
				return
			case <-ch1:
				tn := time.Now()
				fmt.Printf("%s, request (1)\n", time.Now().Format("02-01-2006 15:04:05.000"))
				id := 0
				result, err := devv.ReadData()
				if err != nil {
					// logs.LogWarn.Println(err)
					if errors.Is(err, io.EOF) {
						if time.Since(tn) < timeout/10 {
							countErr++
						}
					} else {
						countErr++
					}
					fmt.Println(err)
					if countErr > max_error {
						ctx.Send(self, &MsgListenError{ID: id})
						return
					}
					break
				}
				countErr = 0
				fmt.Printf("%s: result readbytes: %+v\n",
					time.Now().Format("02-01-2006 15:04:05.000"), result)

				if result == nil {
					break
				}

				if firstFrameCh1 {
					logs.LogInfo.Printf("first readbytes (id=%d): [%X]\n", id, result.RawResponse())
					firstFrameCh1 = false
				}

				if typeCounter == 0 || typeCounter == 1 {

					// fmt.Printf("inputs (1): %d\n", inputs)
					if result.AdultUpFront > 0 && result.AdultUpFront != inputsFront {
						rootctx.Send(self, &messages.Event{
							Id:    int32(id),
							Value: int64(result.AdultUpFront),
							Type:  messages.INPUT,
						})
					}
					inputsFront = result.AdultUpFront

					if result.AdultDownFront > 0 && result.AdultDownFront != outputsFront {
						rootctx.Send(self, &messages.Event{
							Id:    int32(id),
							Value: int64(result.AdultDownFront),
							Type:  messages.OUTPUT,
						})
					}
					outputsFront = result.AdultDownFront

					if result.LocksFront > 0 && result.LocksFront != tamperingFront {
						rootctx.Send(self, &messages.Event{
							Id:    int32(id),
							Value: int64(result.LocksFront),
							Type:  messages.TAMPERING,
						})
					}
					tamperingFront = result.LocksFront
				}

				if typeCounter == 0 || typeCounter == 2 {

					id = 1
					// fmt.Printf("inputs (1): %d\n", inputs)
					if result.AdultUpBack > 0 && result.AdultUpBack != inputsBack {
						rootctx.Send(self, &messages.Event{
							Id:    int32(id),
							Value: int64(result.AdultUpBack),
							Type:  messages.INPUT,
						})
					}
					inputsBack = result.AdultUpBack

					if result.AdultDownBack > 0 && result.AdultDownBack != outputsBack {
						rootctx.Send(self, &messages.Event{
							Id:    int32(id),
							Value: int64(result.AdultDownBack),
							Type:  messages.OUTPUT,
						})
					}
					outputsBack = result.AdultDownBack

					if result.LocksBack > 0 && result.LocksBack != tamperingBack {
						rootctx.Send(self, &messages.Event{
							Id:    int32(id),
							Value: int64(result.LocksBack),
							Type:  messages.TAMPERING,
						})
					}
					tamperingBack = result.LocksBack
				}
			}
		}
	}(rootctx, self)

	return nil
}
