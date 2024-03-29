package listen

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-counterpass/messages"
	"github.com/dumacp/go-levis"
	"github.com/dumacp/go-logs/pkg/logs"
)

func ListenBeane(dev interface{}, quit <-chan int, ctx actor.Context, typeCounter int, externalConsole bool) error {

	timeoutsamples := time.Duration(timeout_samples) * time.Millisecond

	var devv levis.Device

	if v, ok := dev.(levis.Device); !ok {
		return fmt.Errorf("device is not levis device")
	} else {
		devv = v
	}

	timeout := devv.ReadTimeout()

	inputsFront := uint32(0)
	inputsBack := uint32(0)
	outputsFront := uint32(0)
	outputsBack := uint32(0)
	tamperingFront := uint32(0)
	tamperingBack := uint32(0)
	anomaliesFront := uint32(0)
	anomaliesBack := uint32(0)
	alarmCacheFront := byte(0x00)
	alarmCacheBack := byte(0x00)

	self := ctx.Self()
	rootctx := ctx.ActorSystem().Root

	go func(ctx *actor.RootContext, self *actor.PID) {
		var tamperingTimerBack = time.NewTimer(100 * time.Millisecond)
		if !tamperingTimerBack.Stop() {
			select {
			case <-tamperingTimerBack.C:
			case <-time.After(100 * time.Millisecond):
			}
		}
		var tamperingTimerFront = time.NewTimer(100 * time.Millisecond)
		if !tamperingTimerFront.Stop() {
			select {
			case <-tamperingTimerFront.C:
			case <-time.After(100 * time.Millisecond):
			}
		}
		firstFrameCh1 := true
		firstFrameCh2 := true
		const tamperingTimeout = 6 * time.Second
		var ch1 <-chan time.Time
		var ch2 <-chan time.Time
		if typeCounter == 0 || typeCounter == 1 {
			tick1 := time.NewTicker(timeoutsamples)
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
		countErr := 0
		for {
			select {
			case <-quit:
				logs.LogWarn.Println("device levis is closed")
				ctx.Send(self, &MsgListenError{ID: 0})
				ctx.Send(self, &MsgListenError{ID: 1})
				return
			case <-ch1:
				tn := time.Now()
				fmt.Printf("%s, request (slaveID=1)\n", time.Now().Format("02-01-2006 15:04:05.000"))
				id := 0
				devv.SetSlaveID(1)
				result, err := devv.ReadBytesRegister(0x0001, 22/2)
				if err != nil {
					fmt.Printf("slaveID=1 error: %s\n", err)
					if errors.Is(err, io.EOF) {
						if time.Since(tn) < timeout/10 {
							countErr++
						}
					} else {
						countErr++
					}
					log.Println(err)
					if countErr > max_error {
						ctx.Send(self, &MsgListenError{ID: id})
						return
					}
					break
				}
				countErr = 0
				fmt.Printf("%s: result readbytes (slaveId=1): [% X], len: %d\n",
					time.Now().Format("02-01-2006 15:04:05.000"), result, len(result))

				if len(result) < 0x0011-0x0001 {
					logs.LogWarn.Printf("the result (slaveId=1) is incomplete, [% X]", result)
					break
				}
				if firstFrameCh1 {
					logs.LogInfo.Printf("first readbytes (id=0): [% X], len: %d\n", result, len(result))
					firstFrameCh1 = false
				}

				inputs := binary.BigEndian.Uint32(result[0:4])
				outputs := binary.BigEndian.Uint32(result[4:8])
				anomalies := binary.BigEndian.Uint32(result[8:12])
				tampering := binary.BigEndian.Uint32(result[12:16])
				alarm := result[17]

				// fmt.Printf("inputs (1): %d\n", inputs)
				if inputs > 0 && inputs != uint32(inputsFront) {
					rootctx.Send(self, &messages.Event{
						Id:    int32(id),
						Value: int64(inputs),
						Type:  messages.INPUT,
						Raw:   result,
					})
				}
				inputsFront = inputs
				if outputs > 0 && outputs != uint32(outputsFront) {
					rootctx.Send(self, &messages.Event{
						Id:    int32(id),
						Value: int64(outputs),
						Type:  messages.OUTPUT,
						Raw:   result,
					})
				}
				outputsFront = outputs
				if anomalies > 0 && anomalies != uint32(anomaliesFront) {
					rootctx.Send(self, &messages.Event{
						Id:    int32(id),
						Value: int64(anomalies),
						Type:  messages.ANOMALY,
						Raw:   result,
					})
				}
				anomaliesFront = anomalies
				if alarm != 0x00 {
					select {
					case <-tamperingTimerFront.C:
						tamperingTimerFront.Reset(tamperingTimeout)
						rootctx.Send(self, &messages.Event{
							Id:    int32(id),
							Value: int64(tampering - 1),
							Type:  messages.TAMPERING,
							Raw:   result,
						})
					default:
						if tampering > 0 && tampering != uint32(tamperingFront) {
							rootctx.Send(self, &messages.Event{
								Id:    int32(id),
								Value: int64(tampering),
								Type:  messages.TAMPERING,
								Raw:   result,
							})
							if !tamperingTimerFront.Stop() {
								select {
								case <-tamperingTimerFront.C:
								case <-time.After(100 * time.Millisecond):
								}
							}
							tamperingTimerFront.Reset(tamperingTimeout)
						}
					}
				}
				tamperingFront = tampering

				if alarm != 0x00 && alarmCacheFront != alarm {
					logs.LogWarn.Printf("tampering UP(id=%d): %X, %X", id, alarm, tampering)
					fmt.Printf("%s, tampering UP(id=%d): %X\n", time.Now().Format("02-01-2006 15:04:05.000"), id, alarm)
				}
				if alarm == 0x00 && alarmCacheFront != 0x00 {
					logs.LogWarn.Printf("tampering DOWN(id=%d): %X", id, result[17])
					fmt.Printf("%s, tampering UP(id=%d): %X\n", time.Now().Format("02-01-2006 15:04:05.000"), id, alarm)
				}
				alarmCacheFront = alarm

			case <-ch2:
				tn := time.Now()
				fmt.Printf("%s, request (slaveId=2)\n", time.Now().Format("02-01-2006 15:04:05.000"))
				id := 1
				devv.SetSlaveID(2)
				result, err := devv.ReadBytesRegister(0x0001, (22)/2)
				if err != nil {
					fmt.Printf("slaveID=2 error: %s\n", err)
					if errors.Is(err, io.EOF) {
						if time.Since(tn) < timeout/10 {
							countErr++
						}
					} else {
						countErr++
					}
					log.Println(err)
					if countErr > max_error {
						ctx.Send(self, &MsgListenError{ID: id})
						return
					}
					break
				}
				fmt.Printf("%s: result readbytes (slaveId=2): [% X], len: %d\n",
					time.Now().Format("02-01-2006 15:04:05.000"), result, len(result))

				if len(result) < 0x0011-0x0001 {
					logs.LogWarn.Printf("the result (slaveId=2) is incomplete, [% X]", result)
					break
				}
				if firstFrameCh2 {
					logs.LogInfo.Printf("first readbytes (id=1): [% X], len: %d\n", result, len(result))
					firstFrameCh2 = false
				}

				inputs := binary.BigEndian.Uint32(result[0:4])
				outputs := binary.BigEndian.Uint32(result[4:8])
				anomalies := binary.LittleEndian.Uint32(result[8:12])
				tampering := binary.BigEndian.Uint32(result[12:16])
				alarm := result[17]

				// fmt.Printf("inputs (2): %d\n", inputs)
				if inputs > 0 && inputs != uint32(inputsBack) {
					rootctx.Send(self, &messages.Event{
						Id:    int32(id),
						Value: int64(inputs),
						Type:  messages.INPUT,
					})
				}
				inputsBack = inputs
				if outputs > 0 && outputs != uint32(outputsBack) {
					rootctx.Send(self, &messages.Event{
						Id:    int32(id),
						Value: int64(outputs),
						Type:  messages.OUTPUT,
						Raw:   result,
					})
				}
				outputsBack = outputs
				if anomalies > 0 && anomalies != uint32(anomaliesBack) {
					rootctx.Send(self, &messages.Event{
						Id:    int32(id),
						Value: int64(anomalies),
						Type:  messages.ANOMALY,
						Raw:   result,
					})
				}
				anomaliesBack = anomalies
				if alarm != 0x00 {
					select {
					case <-tamperingTimerBack.C:
						tamperingTimerBack.Reset(tamperingTimeout)
						rootctx.Send(self, &messages.Event{
							Id:    int32(id),
							Value: int64(tampering - 1),
							Type:  messages.TAMPERING,
							Raw:   result,
						})
					default:
						if tampering > 0 && tampering != uint32(tamperingBack) {
							rootctx.Send(self, &messages.Event{
								Id:    int32(id),
								Value: int64(tampering),
								Type:  messages.TAMPERING,
								Raw:   result,
							})
							if !tamperingTimerBack.Stop() {
								select {
								case <-tamperingTimerBack.C:
								case <-time.After(100 * time.Millisecond):
								}
							}
							tamperingTimerBack.Reset(tamperingTimeout)
						}
					}
				}
				tamperingBack = tampering

				if alarm != 0x00 && alarmCacheBack != alarm {
					logs.LogWarn.Printf("tampering UP(id=%d): %X, %X", id, alarm, tampering)
					fmt.Printf("%s, tampering UP(id=%d): %X\n", time.Now().Format("02-01-2006 15:04:05.000"), id, alarm)
				}
				if alarm == 0x00 && alarmCacheBack != 0x00 {
					logs.LogWarn.Printf("tampering DOWN(id=%d): %X", id, result[17])
					fmt.Printf("%s, tampering UP(id=%d): %X\n", time.Now().Format("02-01-2006 15:04:05.000"), id, alarm)
				}
				alarmCacheBack = alarm

			}
		}
	}(rootctx, self)

	return nil
}
