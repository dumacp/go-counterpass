//go:build ingnovus
// +build ingnovus

package listen

import (
	"encoding/json"
	"flag"
	"fmt"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-counterpass/messages"
	"github.com/dumacp/go-ingnovus"
	"github.com/dumacp/go-logs/pkg/logs"
)

func Listen(dev interface{}, quit <-chan int, ctx actor.Context, typeCounter int, externalConsole bool) error {

	var devv ingnovus.Device

	if v, ok := dev.(ingnovus.Device); !ok {
		return fmt.Errorf("device is not levis device")
	} else {
		devv = v
	}

	inputsFront := uint32(0)
	inputsBack := uint32(0)
	outputsFront := uint32(0)
	outputsBack := uint32(0)

	self := ctx.Self()
	rootctx := ctx.ActorSystem().Root

	go func(ctx *actor.RootContext, self *actor.PID) {
		defer func() {
			id := typeCounter >> 1
			ctx.Send(self, &MsgListenError{ID: id})
		}()

		ch1 := devv.ListennEvents()

		for {
			select {
			case <-quit:
				logs.LogWarn.Println("device levis is closed")
				return
			case v := <-ch1:

				switch v.Type() {
				case ingnovus.CountingEvent:

					if data, err := json.Marshal(v); err == nil {
						fmt.Printf("counting event: %s\n", data)
					}

					inputsP1 := uint32(v.InputsP1())
					inputsP2 := uint32(v.InputsP2())
					outputsP1 := uint32(v.OutputsP1())
					outputsP2 := uint32(v.OutputsP2())

					// fmt.Printf("inputs (1): %d\n", inputs)
					if inputsP1 > 0 && inputsP1 != inputsFront {
						rootctx.Send(self, &messages.Event{
							Id:    int32(0),
							Value: int64(inputsP1),
							Type:  messages.INPUT,
						})
					}
					inputsFront = inputsP1
					if inputsP2 > 0 && inputsP2 != inputsBack {
						rootctx.Send(self, &messages.Event{
							Id:    int32(1),
							Value: int64(inputsP2),
							Type:  messages.INPUT,
						})
					}
					inputsBack = inputsP2

					if outputsP1 > 0 && outputsP1 != outputsFront {
						rootctx.Send(self, &messages.Event{
							Id:    int32(0),
							Value: int64(outputsP1),
							Type:  messages.OUTPUT,
						})
					}
					outputsFront = outputsP1
					if outputsP2 > 0 && outputsP2 != outputsBack {
						rootctx.Send(self, &messages.Event{
							Id:    int32(1),
							Value: int64(outputsP2),
							Type:  messages.OUTPUT,
						})
					}
					outputsBack = outputsP2
				case ingnovus.AlarmEvent:
					alarm, err := ingnovus.ParseAlarm(v)
					if err != nil {
						logs.LogWarn.Printf("event isn't alarm: %s\n", err)
						break
					}
					if data, err := json.Marshal(alarm); err == nil {
						fmt.Printf("alarm event: %s\n", data)
					}
					switch alarm.Code() {
					case ingnovus.SensorBloqueado1:
						tamperingP1 := uint32(alarm.AlarmTotal())
						if tamperingP1 > 0 {
							rootctx.Send(self, &messages.Event{
								Id:    int32(0),
								Value: int64(tamperingP1),
								Type:  messages.TAMPERING,
							})
						}
					case ingnovus.SensorBloqueado2:
						tamperingP2 := uint32(alarm.AlarmTotal())
						if tamperingP2 > 0 {
							rootctx.Send(self, &messages.Event{
								Id:    int32(1),
								Value: int64(tamperingP2),
								Type:  messages.TAMPERING,
							})
						}
					}
				}
			}
		}
	}(rootctx, self)

	return nil
}
