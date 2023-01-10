package listen

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-counterpass/messages"
	"github.com/dumacp/go-ingnovus/st3310H04"
	"github.com/dumacp/go-logs/pkg/logs"
)

func ListenSt3310_h04(dev interface{}, quit <-chan int, ctx actor.Context, typeCounter int, externalConsole bool) error {

	timeoutsamples := time.Duration(timeout_samples) * time.Millisecond

	var devv st3310H04.Device

	if v, ok := dev.(st3310H04.Device); !ok {
		return fmt.Errorf("device is not ingnovus st3310 device")
	} else {
		devv = v
	}

	timeout := devv.Conf().ReadTimeout

	inputsFront := uint(0)
	inputsBack := uint(0)
	outputsFront := uint(0)
	outputsBack := uint(0)
	inputsBack3 := uint(0)
	outputsBack3 := uint(0)

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
				logs.LogWarn.Println("device ingnovus st3310H04 is closed")
				ctx.Send(self, &MsgListenError{ID: -1})
				return
			case <-ch1:

				tn := time.Now()
				// fmt.Printf("%s, request\n", time.Now().Format("02-01-2006 15:04:05.000"))

				result, err := devv.ReadEvent()
				if err != nil {
					if errors.Is(err, io.EOF) {
						if time.Since(tn) < timeout/10 {
							fmt.Println(err)
							countErr++
						}
					} else {
						countErr++
						fmt.Println(err)
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

				switch result.Type() {
				case st3310H04.CountingEvent:
					if firstFrameCh1 {
						logs.LogInfo.Printf("first readbytes: [%X]\n", result.RawResponse())
						firstFrameCh1 = false
					}

					id := 0

					if result.InputsP1() > 0 && uint(result.InputsP1()) != inputsFront {
						rootctx.Send(self, &messages.Event{
							Id:    int32(id),
							Value: int64(result.InputsP1()),
							Type:  messages.INPUT,
						})
					}
					inputsFront = uint(result.InputsP1())
					if result.OutputsP1() > 0 && uint(result.OutputsP1()) != outputsFront {
						rootctx.Send(self, &messages.Event{
							Id:    int32(id),
							Value: int64(result.OutputsP1()),
							Type:  messages.OUTPUT,
						})
					}
					outputsFront = uint(result.OutputsP1())

					id = 1
					if result.InputsP2() > 0 && uint(result.InputsP2()) != inputsBack {
						rootctx.Send(self, &messages.Event{
							Id:    int32(id),
							Value: int64(result.InputsP2()),
							Type:  messages.INPUT,
						})
					}
					inputsBack = uint(result.InputsP2())
					if result.OutputsP2() > 0 && uint(result.OutputsP2()) != outputsBack {
						rootctx.Send(self, &messages.Event{
							Id:    int32(id),
							Value: int64(result.OutputsP2()),
							Type:  messages.OUTPUT,
						})
					}
					outputsBack = uint(result.OutputsP2())

					if result.InputsP3() > 0 && uint(result.InputsP3()) != inputsBack3 {
						rootctx.Send(self, &messages.Event{
							Id:    int32(id),
							Value: int64(result.InputsP2()),
							Type:  messages.INPUT,
						})
					}
					inputsBack3 = uint(result.InputsP3())
					if result.OutputsP3() > 0 && uint(result.OutputsP3()) != outputsBack3 {
						rootctx.Send(self, &messages.Event{
							Id:    int32(id),
							Value: int64(result.OutputsP3()),
							Type:  messages.OUTPUT,
						})
					}
					outputsBack3 = uint(result.OutputsP3())

				case st3310H04.AlarmEvent:
					alarm, err := st3310H04.ParseAlarm(result)
					if err != nil {
						logs.LogWarn.Printf("error parse alarm: %s", err)
						break
					}
					switch alarm.Code() {
					case st3310H04.SensorBloqueado1:
						tamperingP1 := uint32(alarm.AlarmTotal())
						if tamperingP1 > 0 {
							rootctx.Send(self, &messages.Event{
								Id:    int32(0),
								Value: int64(tamperingP1),
								Type:  messages.TAMPERING,
							})
						}
					case st3310H04.SensorBloqueado2:
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
