package listen

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/dumacp/go-counterpass/messages"
	"github.com/dumacp/go-logs/pkg/logs"
	"github.com/dumacp/sonar/protostd"
)

func ListenSonarEvents(dev interface{}, contxt context.Context, ctx actor.Context, typeCounter int, externalConsole bool) error {
	timeoutsamples := time.Duration(timeout_samples) * time.Millisecond

	var devv protostd.Device

	if v, ok := dev.(protostd.Device); !ok {
		return fmt.Errorf("device is not sonar_std device")
	} else {
		devv = v
	}

	timeout := devv.Conf().ReadTimeout

	maxtimeWithoutResponse := 3 * time.Minute
	lastResponse := time.Now().Add(-30 * time.Second)

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
		quit, cancel := context.WithCancel(contxt)
		defer cancel()
		firstFrameCh1 := true
		countErr := 0
		tick1 := time.NewTicker(timeoutsamples)
		defer tick1.Stop()
		ch1 := tick1.C

		for {
			select {
			case <-quit.Done():
				logs.LogWarn.Println("device sonar_std is closed")
				return
			case <-ch1:
				id := 0
				tn := time.Now()
				fmt.Printf("%s, request\n", time.Now().Format("02-01-2006 15:04:05.000"))

				result, err := devv.ReadData()
				if err != nil {
					// logs.LogWarn.Println(err)
					if errors.Is(err, io.EOF) {
						if time.Since(tn) < timeout/10 {
							countErr++
						} else if time.Since(lastResponse) > maxtimeWithoutResponse {
							ctx.Send(self, &MsgListenError{ID: 0})
							ctx.Send(self, &MsgListenError{ID: 1})
							return
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
				lastResponse = time.Now()
				if result == nil {
					break
				}

				fmt.Printf("%s: result readbytes: %+v\n",
					time.Now().Format("02-01-2006 15:04:05.000"), result)
				if firstFrameCh1 {
					logs.LogInfo.Printf("first readbytes: [%s]\n", result.RawResponse())
					firstFrameCh1 = false
				}
				if result.Inputs1() > 0 && uint(result.Inputs1()) != inputsFront {
					rootctx.Send(self, &messages.Event{
						Id:    int32(id),
						Value: int64(result.Inputs1()),
						Type:  messages.Event_INPUT,
					})
				}
				inputsFront = uint(result.Inputs1())
				if result.Outputs1() > 0 && uint(result.Outputs1()) != outputsFront {
					rootctx.Send(self, &messages.Event{
						Id:    int32(id),
						Value: int64(result.Outputs1()),
						Type:  messages.Event_OUTPUT,
					})
				}
				outputsFront = uint(result.Outputs1())
				if result.Locks1() > 0 && uint(result.Locks1()) != tamperingFront {
					rootctx.Send(self, &messages.Event{
						Id:    int32(id),
						Value: int64(result.Locks1()),
						Type:  messages.Event_TAMPERING,
					})
				}
				tamperingFront = uint(result.Locks1())

				id = 1
				if result.Inputs2() > 0 && uint(result.Inputs2()) != inputsBack {
					rootctx.Send(self, &messages.Event{
						Id:    int32(id),
						Value: int64(result.Inputs2()),
						Type:  messages.Event_INPUT,
					})
				}
				inputsBack = uint(result.Inputs2())
				if result.Outputs2() > 0 && uint(result.Outputs2()) != outputsBack {
					rootctx.Send(self, &messages.Event{
						Id:    int32(id),
						Value: int64(result.Outputs2()),
						Type:  messages.Event_OUTPUT,
					})
				}
				outputsBack = uint(result.Outputs2())

				if result.Locks2() > 0 && uint(result.Locks2()) != tamperingBack {
					rootctx.Send(self, &messages.Event{
						Id:    int32(id),
						Value: int64(result.Locks2()),
						Type:  messages.Event_TAMPERING,
					})
				}
				tamperingBack = uint(result.Locks2())

			}
		}
	}(rootctx, self)

	return nil
}
