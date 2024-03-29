package listen

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-counterpass/messages"
	logirastreo "github.com/dumacp/go-logirastreo/v1"
	"github.com/dumacp/go-logs/pkg/logs"
)

func ListenLogirastreo(dev interface{}, quit <-chan int, ctx actor.Context, typeCounter int, externalConsole bool) error {

	timeoutsamples := time.Duration(timeout_samples) * time.Millisecond

	var devv logirastreo.Device

	if v, ok := dev.(logirastreo.Device); !ok {
		return fmt.Errorf("device is not logirastreo device")
	} else {
		devv = v
	}

	timeout := devv.Conf().ReadTimeout

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
		tick1 := time.NewTicker(timeoutsamples)
		defer tick1.Stop()
		ch1 := tick1.C

		for {
			select {
			case <-quit:
				logs.LogWarn.Println("device logirastreo is closed")
				ctx.Send(self, &MsgListenError{ID: 0})
				ctx.Send(self, &MsgListenError{ID: 1})
				return
			case <-ch1:
				tn := time.Now()
				fmt.Printf("%s, request\n", time.Now().Format("02-01-2006 15:04:05.000"))

				result, err := devv.Read()
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
						ctx.Send(self, &MsgListenError{ID: 0})
						ctx.Send(self, &MsgListenError{ID: 1})
						return
					}
					break
				}
				countErr = 0
				fmt.Printf("%s: result readbytes: %+v\n",
					time.Now().Format("02-01-2006 15:04:05.000"), result)

				if firstFrameCh1 {
					logs.LogInfo.Printf("first readbytes: [%s]\n", result.RawResponse())
					firstFrameCh1 = false
				}

				// fmt.Printf("inputs (1): %d\n", inputs)
				id := 0
				if result.Inputs1() > 0 && uint(result.Inputs1()) != inputsFront {
					rootctx.Send(self, &messages.Event{
						Id:    int32(id),
						Value: int64(result.Inputs1()),
						Type:  messages.INPUT,
					})
				}
				inputsFront = uint(result.Inputs1())
				if result.Outputs1() > 0 && uint(result.Outputs1()) != outputsFront {
					rootctx.Send(self, &messages.Event{
						Id:    int32(id),
						Value: int64(result.Outputs1()),
						Type:  messages.OUTPUT,
					})
				}
				outputsFront = uint(result.Outputs1())
				if result.Locks1() > 0 && uint(result.Locks1()) != tamperingFront {
					rootctx.Send(self, &messages.Event{
						Id:    int32(id),
						Value: int64(result.Locks1()),
						Type:  messages.TAMPERING,
					})
				}
				tamperingFront = uint(result.Locks1())

				id = 1
				if result.Inputs2() > 0 && uint(result.Inputs2()) != inputsBack {
					rootctx.Send(self, &messages.Event{
						Id:    int32(id),
						Value: int64(result.Inputs2()),
						Type:  messages.INPUT,
					})
				}
				inputsBack = uint(result.Inputs2())
				if result.Outputs2() > 0 && uint(result.Outputs2()) != outputsBack {
					rootctx.Send(self, &messages.Event{
						Id:    int32(id),
						Value: int64(result.Outputs2()),
						Type:  messages.OUTPUT,
					})
				}
				outputsBack = uint(result.Outputs2())

				if result.Locks2() > 0 && uint(result.Locks1()) != tamperingBack {
					rootctx.Send(self, &messages.Event{
						Id:    int32(id),
						Value: int64(result.Locks2()),
						Type:  messages.TAMPERING,
					})
				}
				tamperingBack = uint(result.Locks2())

			}
		}
	}(rootctx, self)

	return nil
}
