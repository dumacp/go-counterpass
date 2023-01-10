package listen

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-counterpass/messages"
	"github.com/dumacp/go-ingnovus/st300"
	"github.com/dumacp/go-logs/pkg/logs"
)

func ListenSt300(dev interface{}, quit <-chan int, ctx actor.Context, typeCounter int, externalConsole bool) error {

	timeoutsamples := time.Duration(timeout_samples) * time.Millisecond

	var devv st300.Device

	if v, ok := dev.(st300.Device); !ok {
		return fmt.Errorf("device is not ingnovus_st300 device")
	} else {
		devv = v
	}

	timeout := devv.Conf().ReadTimeout

	inputsFront := uint(0)
	// inputsBack := uint(0)
	outputsFront := uint(0)
	// outputsBack := uint(0)
	//tamperingFront := uint(0)
	//tamperingBack := uint(0)
	// anomaliesFront := uint(0)
	// anomaliesBack := uint(0)
	// alarmCacheFront := byte(0x00)
	// alarmCacheBack := byte(0x00)

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
				logs.LogWarn.Println("device ingnovus st300 is closed")
				ctx.Send(self, &MsgListenError{ID: -1})
				return
			case <-ch1:
				tn := time.Now()
				id := 0
				// fmt.Printf("%s, request\n", time.Now().Format("02-01-2006 15:04:05.000"))
				result, err := devv.Read()
				if err != nil {
					fmt.Println(err)
					if errors.Is(err, io.EOF) {
						if time.Since(tn) < timeout/10 {
							countErr++
						}
					} else {
						countErr++
					}
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

				if result == nil {
					break
				}
				if firstFrameCh1 {
					logs.LogInfo.Printf("first readbytes: [%s]\n", result.RawResponse())
					firstFrameCh1 = false
				}

				if result.Inputs() > 0 && uint(result.Inputs()) != inputsFront {
					rootctx.Send(self, &messages.Event{
						Id:    int32(id),
						Value: int64(result.Inputs()),
						Type:  messages.INPUT,
					})
				}
				inputsFront = uint(result.Inputs())
				if result.Outputs() > 0 && uint(result.Outputs()) != outputsFront {
					rootctx.Send(self, &messages.Event{
						Id:    int32(id),
						Value: int64(result.Outputs()),
						Type:  messages.OUTPUT,
					})
				}
				outputsFront = uint(result.Outputs())
				//if result.Locks1() > 0 && uint(result.Locks1()) != tamperingFront {
				//	rootctx.Send(self, &messages.Event{
				//		Id:    int32(id),
				//		Value: int64(result.Locks1()),
				//		Type:  messages.TAMPERING,
				//	})
				//}
				//tamperingFront = uint(result.Locks1())

			}
		}
	}(rootctx, self)

	return nil
}
