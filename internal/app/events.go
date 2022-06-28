package app

import (
	"encoding/json"
	"log"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-counterpass/internal/doors"
	"github.com/dumacp/go-counterpass/internal/gps"
	"github.com/dumacp/go-counterpass/messages"
	"github.com/dumacp/go-logs/pkg/logs"
	"github.com/dumacp/pubsub"
)

func buildEventPass(ctx actor.Context, id int32,
	ttype messages.Event_EventType, value int64, pidGps *actor.PID,
	puerta map[uint]uint, raw []byte) ([]byte, error) {

	contadores := []int64{0, 0}
	if ttype == messages.INPUT {
		contadores[0] = value
	} else if ttype == messages.OUTPUT {
		contadores[1] = value
	}
	frame := ""

	if pidGps != nil {
		res, err := ctx.RequestFuture(pidGps, &gps.MsgGetGps{}, 300*time.Millisecond).Result()
		if err == nil {
			switch msg := res.(type) {
			case *gps.MsgGPS:
				if msg.Data != nil {
					frame = string(msg.Data)
				}
			default:
				logs.LogWarn.Println("get gps nil")
			}
		} else {
			logs.LogWarn.Printf("get gps err: %s", err)
		}
	}

	doorState := uint(0)
	switch id {
	case 0:
		if vm, ok := puerta[doors.GpioPuerta1]; ok {
			doorState = vm
		}
	case 1:
		if vm, ok := puerta[doors.GpioPuerta2]; ok {
			doorState = vm
		}
	}

	message := &pubsub.Message{
		Timestamp: float64(time.Now().UnixNano()) / 1000000000,
		Type:      "COUNTERSDOOR",
	}

	val := struct {
		Coord    string  `json:"coord"`
		ID       int32   `json:"id"`
		State    uint    `json:"state"`
		Counters []int64 `json:"counters"`
		Type     string  `json:"type,omitempty"`
		Raw      []byte  `json:"raw,omitempty"`
	}{
		frame,
		int32(id),
		doorState,
		contadores[0:2],
		VendorCounter,
		raw,
	}
	message.Value = val

	msg, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}
	log.Printf("%s\n", msg)

	return msg, nil
}

func buildEventTampering(ctx actor.Context, id int32, value int64, pidGps *actor.PID, puerta map[uint]uint, raw []byte) ([]byte, error) {

	frame := ""
	if pidGps != nil {
		res, err := ctx.RequestFuture(pidGps, &gps.MsgGetGps{}, 300*time.Millisecond).Result()
		if err == nil {
			switch msg := res.(type) {
			case *gps.MsgGPS:
				if msg.Data != nil {
					frame = string(msg.Data)
				}
			default:
				logs.LogWarn.Println("get gps nil")
			}
		} else {
			logs.LogWarn.Printf("get gps err: %s", err)
		}
	}

	doorState := uint(0)
	switch id {
	case 0:
		if vm, ok := puerta[doors.GpioPuerta1]; ok {
			doorState = vm
		}
	case 1:
		if vm, ok := puerta[doors.GpioPuerta2]; ok {
			doorState = vm
		}
	}

	message := &pubsub.Message{
		Timestamp: float64(time.Now().UnixNano()) / 1000000000,
		Type:      "TAMPERING",
	}

	val := struct {
		Coord    string  `json:"coord"`
		ID       int32   `json:"id"`
		State    uint    `json:"state"`
		Counters []int64 `json:"counters"`
		Type     string  `json:"type,omitempty"`
		Raw      []byte  `json:"raw,omitempty"`
	}{
		frame,
		int32(id),
		doorState,
		[]int64{0, 0},
		VendorCounter,
		raw,
	}
	if id == 0 {
		val.Counters[0] = value
	} else if id == 1 {
		val.Counters[1] = value
	}

	message.Value = val

	msg, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}
	log.Printf("%s\n", msg)

	return msg, nil
}

func buildEventAnomalies(ctx actor.Context, id int32, value int64, pidGps *actor.PID, puerta map[uint]uint, raw []byte) ([]byte, error) {

	frame := ""
	if pidGps != nil {
		res, err := ctx.RequestFuture(pidGps, &gps.MsgGetGps{}, 300*time.Millisecond).Result()
		if err == nil {
			switch msg := res.(type) {
			case *gps.MsgGPS:
				if msg.Data != nil {
					frame = string(msg.Data)
				}
			default:
				logs.LogWarn.Println("get gps nil")
			}
		} else {
			logs.LogWarn.Printf("get gps err: %s", err)
		}
	}

	doorState := uint(0)
	switch id {
	case 0:
		if vm, ok := puerta[doors.GpioPuerta1]; ok {
			doorState = vm
		}
	case 1:
		if vm, ok := puerta[doors.GpioPuerta2]; ok {
			doorState = vm
		}
	}

	message := &pubsub.Message{
		Timestamp: float64(time.Now().UnixNano()) / 1000000000,
		Type:      "PAX_ANOMALIES",
	}

	val := struct {
		Coord    string  `json:"coord"`
		ID       int32   `json:"id"`
		State    uint    `json:"state"`
		Counters []int64 `json:"counters"`
		Type     string  `json:"type,omitempty"`
		Raw      []byte  `json:"raw,omitempty"`
	}{
		frame,
		int32(id),
		doorState,
		[]int64{0, 0},
		VendorCounter,
		raw,
	}
	if id == 0 {
		val.Counters[0] = value
	} else if id == 1 {
		val.Counters[1] = value
	}

	message.Value = val

	msg, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}
	log.Printf("%s\n", msg)

	return msg, nil
}

func buildListenError(ctx actor.Context, id int32, value int64, pidGps *actor.PID, puerta map[uint]uint, raw []byte) ([]byte, error) {
	// message := &pubsub.Message{
	// 	Timestamp: float64(time.Now().UnixNano()) / 1000000000,
	// 	Type:      "CounterDisconnect",
	// 	Value:     "{}",
	// }
	// data, err := json.Marshal(message)
	// if err != nil {
	// 	return nil, err
	// }
	// log.Printf("data: %q", data)
	// return data, nil

	frame := ""
	if pidGps != nil {
		res, err := ctx.RequestFuture(pidGps, &gps.MsgGetGps{}, 300*time.Millisecond).Result()
		if err == nil {
			switch msg := res.(type) {
			case *gps.MsgGPS:
				if msg.Data != nil {
					frame = string(msg.Data)
				}
			default:
				logs.LogWarn.Println("get gps nil")
			}
		} else {
			logs.LogWarn.Printf("get gps err: %s", err)
		}
	}

	doorState := uint(0)
	switch id {
	case 0:
		if vm, ok := puerta[doors.GpioPuerta1]; ok {
			doorState = vm
		}
	case 1:
		if vm, ok := puerta[doors.GpioPuerta2]; ok {
			doorState = vm
		}
	}

	message := &pubsub.Message{
		Timestamp: float64(time.Now().UnixNano()) / 1000000000,
		Type:      "CounterDisconnect",
	}

	val := struct {
		Coord    string  `json:"coord"`
		ID       int32   `json:"id"`
		State    uint    `json:"state"`
		Counters []int64 `json:"counters"`
		Type     string  `json:"type,omitempty"`
		Raw      []byte  `json:"raw,omitempty"`
	}{
		frame,
		int32(id),
		doorState,
		[]int64{0, 0},
		VendorCounter,
		raw,
	}
	if id == 0 {
		val.Counters[0] = value
	} else if id == 1 {
		val.Counters[1] = value
	}

	message.Value = val

	msg, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}
	log.Printf("%s\n", msg)

	return msg, nil
}
