//go:build st300
// +build st300

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
	"github.com/dumacp/go-ingnovus/st300"
	"github.com/dumacp/go-logs/pkg/logs"
)

var timeout_samples int

func init() {
	flag.IntVar(&timeout_samples, "timeout", 3000, "timeout samples in millis")
}

const (
	max_error = 3
)

func Listen(dev interface{}, quit <-chan int, ctx actor.Context, typeCounter int, externalConsole bool) error {

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
		defer ctx.Send(self, &MsgListenError{})
		countErr := 0
		tick1 := time.NewTicker(timeoutsamples)
		defer tick1.Stop()
		ch1 := tick1.C

		for {
			select {
			case <-quit:
				logs.LogWarn.Println("device ingnovus st300 is closed")
				return
			case <-ch1:
				tn := time.Now()
				fmt.Printf("%s, request (1)\n", time.Now().Format("02-01-2006 15:04:05.000"))

				result, err := devv.Read()
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

				if result == nil {
					break
				}

				// fmt.Printf("inputs (1): %d\n", inputs)
				id := 0
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
