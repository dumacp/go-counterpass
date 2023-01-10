package app

import (
	"fmt"
	"log"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/persistence"
	"github.com/dumacp/go-counterpass/internal/doors"
	"github.com/dumacp/go-counterpass/internal/gps"
	"github.com/dumacp/go-counterpass/internal/listen"
	"github.com/dumacp/go-counterpass/internal/pubsub"
	"github.com/dumacp/go-counterpass/messages"
	"github.com/dumacp/go-logs/pkg/logs"
)

const (
	clietnName        = "go-counting-actor"
	topicCounterEvent = "EVENTS/backcounter"
	topicEvents       = "EVENTS/counterevents"
	topicCounter      = "COUNTERSMAPDOOR"
)

type CountingActor struct {
	persistence.Mixin
	puertas            map[uint]uint
	openState          map[int32]uint
	disableDoorGpio    bool
	disableSend        bool
	counterType        int
	disablePersistence bool
	inputs             map[int32]int64
	outputs            map[int32]int64
	anomalies          map[int32]int64
	tampering          map[int32]int64
	rawInputs          map[int32]int64
	rawOutputs         map[int32]int64
	rawTampering       map[int32]int64
	rawAnomalies       map[int32]int64
	pidGps             *actor.PID
	pidQueue           *actor.PID
	firstEventInput0   bool
	firstEventInput1   bool
	firstEventOutput0  bool
	firstEventOutput1  bool
}

//NewCountingActor create CountingActor
func NewCountingActor(vendor string) *CountingActor {
	VendorCounter = vendor
	count := &CountingActor{}
	count.openState = make(map[int32]uint)
	count.puertas = make(map[uint]uint)

	return count
}

//SetZeroOpenStateDoor0 set the open state in gpio door
func (a *CountingActor) SetZeroOpenStateDoor0(state bool) {
	if state {
		a.openState[0] = 0
	} else {
		a.openState[0] = 1
	}
}

//SetZeroOpenStateDoor1 set the open state in gpio door
func (a *CountingActor) SetZeroOpenStateDoor1(state bool) {
	if state {
		a.openState[1] = 0
	} else {
		a.openState[1] = 1
	}
}

//DisableDoorGpioListen Disable DoorGpio
func (a *CountingActor) DisableDoorGpioListen(state bool) {
	if state {
		a.disableDoorGpio = true
	} else {
		a.disableDoorGpio = false
	}
}

//CounterType set counter type
func (a *CountingActor) CounterType(tp int) {
	a.counterType = tp
}

//SetGPStoConsole set gps to consolse
func (a *CountingActor) SetGPStoConsole(gpsConsole bool) {
}

//DisablePersistence disable persistence
func (a *CountingActor) DisablePersistence(disable bool) {
	a.disablePersistence = disable
}

func (a *CountingActor) Receive(ctx actor.Context) {
	fmt.Printf("actor \"%s\", message: %v, %T\n", ctx.Self().GetId(), ctx.Message(), ctx.Message())
	switch msg := ctx.Message().(type) {
	case *actor.Started:
		if a.disablePersistence {
			logs.LogInfo.Println("disable persistence")
		}
		a.inputs = make(map[int32]int64)
		a.outputs = make(map[int32]int64)
		a.rawInputs = make(map[int32]int64)
		a.rawOutputs = make(map[int32]int64)
		a.anomalies = make(map[int32]int64)
		a.tampering = make(map[int32]int64)
		a.rawAnomalies = make(map[int32]int64)
		a.rawTampering = make(map[int32]int64)

		a.firstEventInput0 = true
		a.firstEventInput1 = true
		a.firstEventOutput0 = true
		a.firstEventOutput1 = true

		propsQueue := actor.PropsFromProducer(NewQueueActor)
		pidQueue, err := ctx.SpawnNamed(propsQueue, "queueEventActor")
		if err != nil {
			time.Sleep(3 * time.Second)
			panic(err)
		}

		a.pidQueue = pidQueue

		logs.LogInfo.Printf("actor started \"%s\"", ctx.Self().Id)
	case *MsgRegisterGPS:
		if ctx.Sender() != nil {
			a.pidGps = ctx.Sender()
		}
	case *persistence.RequestSnapshot:
		logs.LogInfo.Printf("snapshot internal state: inputs -> '%v', outputs -> '%v', rawInputs -> %v, rawOutpts -> %v\n",
			a.inputs, a.outputs, a.rawInputs, a.rawOutputs)
		fmt.Printf(`snapshot internal state:
		inputs -> "%v", outputs -> "%v",
		rawInputs -> "%v", rawOutpts -> "%v",
		rawAnomalies -> "%v", rawTampering -> "%v"`,
			a.inputs, a.outputs, a.rawInputs, a.rawOutputs, a.rawAnomalies, a.rawTampering)
		snap := &messages.Snapshot{
			Inputs:       a.inputs,
			Outputs:      a.outputs,
			RawInputs:    a.rawInputs,
			RawOutputs:   a.rawOutputs,
			RawAnomalies: a.rawAnomalies,
			RawTampering: a.rawTampering,
			Anomalies:    a.anomalies,
			Tampering:    a.tampering,
		}
		if !a.disablePersistence {
			a.PersistSnapshot(snap)
		}
		data, err := registersMap(a.inputs, a.outputs, a.anomalies, a.tampering, a.counterType)
		if err != nil {
			logs.LogWarn.Println(err)
			break
		}
		log.Printf("data: %q", data)
		if !a.disableSend {
			pubsub.Publish(topicCounter, data)
		}

	case *MsgInitCounters:
		inputs := make(map[int32]int64)
		outputs := make(map[int32]int64)
		inputs[0] = msg.Inputs0
		inputs[1] = msg.Inputs1
		outputs[0] = msg.Outputs0
		outputs[1] = msg.Outputs1
		snap := &messages.Snapshot{
			Inputs:  inputs,
			Outputs: outputs,
		}
		if !a.disablePersistence {
			a.PersistSnapshot(snap)
		}
		log.Printf("init counters -> %v", snap)
	case *MsgSendRegisters:

		if verifySum(a.outputs) <= 0 && verifySum(a.inputs) <= 0 {
			break
		}
		data, err := registersMap(a.inputs, a.outputs, a.anomalies, a.tampering, a.counterType)
		if err != nil {
			logs.LogWarn.Println(err)
			break
		}
		log.Printf("data: %q", data)
		if !a.disableSend {
			pubsub.Publish(topicCounter, data)
		}
	case *messages.Snapshot:
		if msg.GetInputs() != nil {
			a.inputs = msg.GetInputs()
		}
		if msg.GetOutputs() != nil {
			a.outputs = msg.GetOutputs()
		}
		if msg.GetRawInputs() != nil {
			a.rawInputs = msg.GetRawInputs()
		}
		if msg.GetRawOutputs() != nil {
			a.rawOutputs = msg.GetRawOutputs()
		}
		if msg.GetAnomalies() != nil {
			a.anomalies = msg.GetRawAnomalies()
		}
		if msg.GetTampering() != nil {
			a.tampering = msg.GetRawTampering()
		}
		if msg.GetRawAnomalies() != nil {
			a.rawAnomalies = msg.GetRawAnomalies()
		}
		if msg.GetRawTampering() != nil {
			a.rawTampering = msg.GetRawTampering()
		}
		logs.LogInfo.Printf(`recover snapshot, internal state changed to:
	inputs -> "%v", outputs -> "%v"
	rawInputs: "%v", rawOutputs: "%v"
	rawAnomalies: "%v", rawTampering: "%v"`,
			a.inputs, a.outputs, a.rawInputs, a.rawOutputs,
			a.rawAnomalies, a.rawTampering)
		fmt.Printf(`recover snapshot, internal state changed to:
	inputs -> "%v", outputs -> "%v"
	rawInputs: "%v", rawOutputs: "%v"
	rawAnomalies: "%v", rawTampering: "%v"`,
			a.inputs, a.outputs, a.rawInputs, a.rawOutputs,
			a.rawAnomalies, a.rawTampering)
	case *persistence.ReplayComplete:
		logs.LogInfo.Printf("replay completed, internal state changed to:\n\tinputs -> '%v', outputs -> '%v'\n\trawInputs: %v, rawOutputs: %v",
			a.inputs, a.outputs, a.rawInputs, a.rawOutputs)
		snap := &messages.Snapshot{
			Inputs:       a.inputs,
			Outputs:      a.outputs,
			RawInputs:    a.rawInputs,
			RawOutputs:   a.rawOutputs,
			RawAnomalies: a.rawAnomalies,
			RawTampering: a.rawTampering,
			Anomalies:    a.anomalies,
			Tampering:    a.tampering,
		}
		if !a.disablePersistence {
			a.PersistSnapshot(snap)
		}
	case *messages.Event:
		if !a.disablePersistence && !a.Recovering() {
			a.PersistReceive(msg)
		}
		switch msg.GetType() {
		case messages.INPUT:
			if err := func() error {
				id := msg.Id
				if a.firstEventInput0 || a.firstEventInput1 {
					logs.LogInfo.Printf("new first input(%d). last value: %d, new value: %d",
						id, a.inputs[id], msg.GetValue())
					data, err := buildListenStarted(ctx, id, 0, a.pidGps)
					if err != nil {
						return err
					}
					if !a.disableSend {
						pubsub.Publish(topicEvents, data)
					}
					a.firstEventInput0 = (!a.firstEventInput0 && true) || (a.firstEventInput0 && false)
					a.firstEventInput1 = (!a.firstEventInput1 && true) || (a.firstEventInput1 && false)
				}
				defer func() {
					a.rawInputs[id] = msg.GetValue()
				}()
				if _, ok := a.rawInputs[id]; !ok {
					a.rawInputs[id] = 0
				}
				if _, ok := a.inputs[id]; !ok {
					a.inputs[id] = 0
				}

				diff := msg.GetValue() - a.rawInputs[id]

				if diff > 0 && a.rawInputs[id] <= 0 {
					if v, ok := a.puertas[uint(id)]; a.disableDoorGpio || !ok || v == a.openState[id] {
						a.inputs[id] += 1

						data, err := buildEventPass(ctx, int(id), msg.GetType(), 1, a.pidGps, a.puertas, msg.Raw)
						if err != nil {
							return err
						}
						if !a.disableSend && (a.disablePersistence || !a.Recovering()) {
							// pubsub.Publish(topicCounterEvent, data)
							ctx.Send(a.pidQueue, &MsgQueueEvent{
								Event:     data,
								Type:      msg.GetType(),
								Timestamp: time.Now(),
								ID:        id,
							})
						}
					}
				} else if diff > 0 && diff < 60 {
					if v, ok := a.puertas[uint(id)]; a.disableDoorGpio || !ok || v == a.openState[id] {
						a.inputs[id] += diff

						data, err := buildEventPass(ctx, int(id), msg.GetType(), diff, a.pidGps, a.puertas, msg.Raw)
						if err != nil {
							return err
						}
						if !a.disableSend && (a.disablePersistence || !a.Recovering()) {
							ctx.Send(a.pidQueue, &MsgQueueEvent{
								Event:     data,
								Type:      msg.GetType(),
								Timestamp: time.Now(),
								ID:        id,
							})
						}
						if diff > 5 {
							logs.LogError.Printf("diff is greater than 5 -> msg.GetValue(): %d, a.rawInputs[%d]:: %d", msg.GetValue(), id, a.rawInputs[id])
						}
					}
				} else if diff > -5 && diff < 0 {
					logs.LogWarn.Printf("diff is negative -> msg.GetValue(): %d, a.rawInputs[%d]: %d", msg.GetValue(), id, a.rawInputs[id])

					a.inputs[id] += 1
					data, err := buildEventPass(ctx, int(id), msg.GetType(), 1, a.pidGps, a.puertas, msg.Raw)
					if err != nil {
						return err
					}
					if !a.disableSend && (a.disablePersistence || !a.Recovering()) {
						// pubsub.Publish(topicCounterEvent, data)
						ctx.Send(a.pidQueue, &MsgQueueEvent{
							Event:     data,
							Type:      msg.GetType(),
							Timestamp: time.Now(),
							ID:        id,
						})
					}
				} else if diff != 0 {
					logs.LogError.Printf("diff is greater than 60 or less than -5 -> msg.GetValue(): %d, a.rawInputs[%d]: %d", msg.GetValue(), id, a.rawInputs[id])
				}
				return nil

			}(); err != nil {
				logs.LogWarn.Println(err)
			}

		case messages.OUTPUT:
			if err := func() error {
				id := msg.Id
				if a.firstEventInput0 || a.firstEventInput1 {
					logs.LogInfo.Printf("new first input(%d). last value: %d, new value: %d",
						id, a.inputs[id], msg.GetValue())
					data, err := buildListenStarted(ctx, id, 0, a.pidGps)
					if err != nil {
						return err
					}
					if !a.disableSend {
						pubsub.Publish(topicEvents, data)
					}
					a.firstEventOutput0 = (!a.firstEventOutput0 && true) || (a.firstEventOutput0 && false)
					a.firstEventOutput1 = (!a.firstEventOutput1 && true) || (a.firstEventOutput1 && false)
				}
				defer func() {
					a.rawOutputs[id] = msg.GetValue()
				}()
				if _, ok := a.rawOutputs[id]; !ok {
					a.rawOutputs[id] = 0
				}
				if _, ok := a.outputs[id]; !ok {
					a.outputs[id] = 0
				}
				diff := msg.GetValue() - a.rawOutputs[id]
				if diff > 0 && a.rawOutputs[id] <= 0 {
					if v, ok := a.puertas[uint(id)]; a.disableDoorGpio || !ok || v == a.openState[id] {
						a.outputs[id] += 1
						data, err := buildEventPass(ctx, int(id), msg.GetType(), 1, a.pidGps, a.puertas, msg.Raw)
						if err != nil {
							return err
						}
						if !a.disableSend && (a.disablePersistence || !a.Recovering()) {
							// pubsub.Publish(topicCounterEvent, data)
							ctx.Send(a.pidQueue, &MsgQueueEvent{
								Event:     data,
								Type:      msg.GetType(),
								Timestamp: time.Now(),
								ID:        id,
							})
						}
					}
				} else if diff > 0 && diff < 60 {
					if v, ok := a.puertas[uint(id)]; a.disableDoorGpio || !ok || v == a.openState[id] {
						a.outputs[id] += diff
						data, err := buildEventPass(ctx, int(id), msg.GetType(), diff, a.pidGps, a.puertas, msg.Raw)
						if err != nil {
							return err
						}

						if !a.disableSend && (a.disablePersistence || !a.Recovering()) {
							// pubsub.Publish(topicCounterEvent, data)
							ctx.Send(a.pidQueue, &MsgQueueEvent{
								Event:     data,
								Type:      msg.GetType(),
								Timestamp: time.Now(),
								ID:        id,
							})
						}
						if diff > 5 {
							logs.LogWarn.Printf("diff is greater than 5 -> msg.GetValue(): %d, a.rawOutputs[%d]: %d", msg.GetValue(), id, a.rawOutputs[id])
						}
					}
				} else if diff > -5 && diff < 0 {
					// a.outputs[id] += msg.GetValue()
					logs.LogWarn.Printf("diff is negative -> msg.GetValue(): %d, a.rawOutputs[%d]: %d", msg.GetValue(), id, a.rawOutputs[id])

					a.outputs[id] += 1
					data, err := buildEventPass(ctx, int(id), msg.GetType(), 1, a.pidGps, a.puertas, msg.Raw)
					if err != nil {
						return err
					}
					if !a.disableSend && (a.disablePersistence || !a.Recovering()) {
						// pubsub.Publish(topicCounterEvent, data)
						ctx.Send(a.pidQueue, &MsgQueueEvent{
							Event:     data,
							Type:      msg.GetType(),
							Timestamp: time.Now(),
							ID:        id,
						})
					}
				} else if diff != 0 {
					logs.LogError.Printf("diff is greater than 60 or less than -5 -> msg.GetValue(): %d, a.rawOutputs[%d]: %d", msg.GetValue(), id, a.rawOutputs[id])
				}
				return nil
			}(); err != nil {
				logs.LogWarn.Println(err)
			}
		case messages.TAMPERING:
			if err := func() error {
				id := msg.Id
				defer func() {
					a.rawTampering[id] = msg.GetValue()
				}()
				logs.LogWarn.Printf("TAMPERING(id=%d), value = %d", id, msg.GetValue())
				diff := msg.GetValue() - a.rawTampering[id]
				if diff > 0 && diff < 60 {
					a.tampering[id] += diff
					data, err := buildEventTampering(ctx, msg.Id, diff, a.pidGps, a.puertas, msg.Raw)
					if err != nil {
						return err
					}
					if !a.disableSend && (a.disablePersistence || !a.Recovering()) {
						pubsub.Publish(topicEvents, data)
					}
				} else if diff != 0 {
					a.tampering[id] += 1
					data, err := buildEventTampering(ctx, msg.Id, 1, a.pidGps, a.puertas, msg.Raw)
					if err != nil {
						return err
					}
					if !a.disableSend && (a.disablePersistence || !a.Recovering()) {
						pubsub.Publish(topicEvents, data)
					}
				}
				return nil
			}(); err != nil {
				logs.LogWarn.Println(err)
			}
		case messages.ANOMALY:
			if err := func() error {
				id := msg.Id
				defer func() {
					a.rawAnomalies[id] = msg.GetValue()
				}()
				logs.LogWarn.Printf("ANOMALY(id=%d), value = %d", id, msg.GetValue())
				diff := msg.GetValue() - a.rawAnomalies[id]
				if diff > 0 && diff <= 5 {
					a.anomalies[id] += diff
					data, err := buildEventAnomalies(ctx, msg.Id, diff, a.pidGps, a.puertas, msg.Raw)
					if err != nil {
						return err
					}
					if !a.disableSend && (a.disablePersistence || !a.Recovering()) {
						pubsub.Publish(topicEvents, data)
					}
				} else if diff != 0 {
					a.anomalies[id] += 1
					data, err := buildEventAnomalies(ctx, msg.Id, 1, a.pidGps, a.puertas, msg.Raw)
					if err != nil {
						return err
					}
					if !a.disableSend && (a.disablePersistence || !a.Recovering()) {
						pubsub.Publish(topicEvents, data)
					}
				}
				return nil
			}(); err != nil {
				logs.LogWarn.Println(err)
			}
		}
	case *listen.MsgListenError:
		logs.LogWarn.Printf("counter(id=%d) listen error", msg.ID)
		data, err := buildListenError(ctx, int32(msg.ID), 1, a.pidGps)
		if err != nil {
			logs.LogWarn.Println(err)
			break
		}
		if !a.disableSend {
			pubsub.Publish(topicEvents, data)
		}
	case *listen.MsgListenStarted:
		logs.LogInfo.Printf("counter(typeCounter=%d) listen started", msg.TypeCounter)
	case *doors.MsgDoor:
		a.puertas[msg.ID] = msg.Value
	case *gps.MsgGPS:
	case *MsgSendEvents:
		if !msg.Data {
			a.disableSend = true
		}
	case *actor.Terminated:
		logs.LogWarn.Printf("actor terminated: %s", msg.GetWho().GetAddress())
	case *actor.Stopped:
		logs.LogWarn.Printf("actor stopped, reason: %s", msg)
	}
}
