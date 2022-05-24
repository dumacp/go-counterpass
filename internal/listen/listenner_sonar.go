// +build sonar

package listen

import (
	"bytes"
	"container/list"
	"fmt"
	"strconv"
	"strings"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-counterpass/messages"
	"github.com/dumacp/go-logs/pkg/logs"
	"github.com/dumacp/sonar/contador"
)

type mem struct {
	raw     []int64
	inputs  []int64
	outputs []int64
	locks   []int64
}

func Listen(dev interface{}, quit <-chan int, ctx actor.Context, typeCounter int, externalConsole bool) error {

	var devv contador.Device

	if v, ok := dev.(contador.Device); !ok {
		return fmt.Errorf("device is not sonar device")
	} else {
		devv = v
	}

	first := true

	memm := new(mem)
	memm.inputs = make([]int64, 2)
	memm.outputs = make([]int64, 2)
	memm.locks = make([]int64, 2)
	memm.raw = make([]int64, 2)

	queue := new(list.List)

	ch := devv.ListenRawChannel(quit)

	self := ctx.Self()
	rootctx := ctx.ActorSystem().Root

	go func(ctx *actor.RootContext, self *actor.PID) {
		defer ctx.Send(Self, &MsgListenError{})
		for v := range ch {
			if bytes.Contains(v, []byte("RPT")) {
				if first {
					first = false
					logs.LogInfo.Printf("actual door register: %s\n", v)
				}
				queue.PushBack(v)
				if queue.Len() > 5 {
					e := queue.Front()
					queue.Remove(e)
				}
				processData(ctx, pid*actor.PID, v, memm, typeCounter)
				if externalConsole {
					devv.Send([]byte(v))
				}
			}
		}
	}(rootctx, self)

	return nil
}

func processData(ctx actor.Context, pid *actor.PID, v []byte, mem *mem, counterType int) {

	split := strings.Split(string(v), ";")
	if len(split) < 17 {
		return
	}
	data := make([]int64, 15)
	ok := true
	for i := range data {
		xint, err := strconv.ParseInt((split[i+2]), 10, 64)
		if err != nil {
			ok = false
			break
		}
		data[i] = xint
	}
	if !ok {
		return
	}

	logs.LogBuild.Printf("data: %v", data)
	logs.LogBuild.Printf("mem.raw: %v", mem.raw)

	if len(mem.raw) <= 0 || data[0] > mem.raw[0] {
		enters := data[0]
		if diff := enters - mem.inputs[0]; diff > 0 || len(mem.raw) <= 0 {
			sendEvent(ctx, pid, counterType, &messages.Event{Id: 0, Type: messages.INPUT, Value: enters})
		}
		mem.inputs[0] = enters
	}
	if len(mem.raw) <= 0 || data[1] > mem.raw[1] {
		enters := data[1]
		if diff := enters - mem.outputs[0]; diff > 0 || len(mem.raw) <= 0 {
			sendEvent(ctx, pid, counterType, &messages.Event{Id: 0, Type: messages.OUTPUT, Value: enters})
		}
		mem.outputs[0] = enters
	}
	if counterType == 0 {
		if len(mem.raw) <= 0 || data[2] > mem.raw[2] {
			enters := data[2]
			if diff := enters - mem.inputs[1]; diff > 0 || len(mem.raw) <= 0 {
				sendEvent(ctx, counterType, &messages.Event{Id: 1, Type: messages.INPUT, Value: enters})
			}
			mem.inputs[1] = enters
		}
		if len(mem.raw) <= 0 || data[3] > mem.raw[3] {
			enters := data[3]
			if diff := enters - mem.outputs[1]; diff > 0 || len(mem.raw) <= 0 {
				sendEvent(ctx, pid, counterType, &messages.Event{Id: 1, Type: messages.OUTPUT, Value: enters})
			}
			mem.outputs[1] = enters
		}
	}

	switch {
	case len(split) == 17:

		if len(mem.raw) >= 10 {
			id := int32(0)
			if data[10] > mem.raw[10] {
				enters := data[10]
				if diff := enters - mem.locks[0]; diff > 0 {
					sendEvent(ctx, counterType, &messages.Event{
						Id: id, Type: messages.TAMPERING, Value: enters})
				}
				mem.locks[0] = enters
			}
		}
		if len(mem.raw) >= 13 {
			id := int32(1)
			if data[13] > mem.raw[13] {
				enters := data[13]
				if diff := enters - mem.locks[1]; diff > 0 {
					sendEvent(ctx, pid, counterType, &messages.Event{
						Id: id, Type: messages.TAMPERING, Value: enters})
				}
				mem.locks[1] = enters
			}
		}
	case len(split) == 18:
		if len(mem.raw) >= 14 && data[14] > mem.raw[14] {
			id := int32(0)
			if data[13] > mem.raw[13] {
				id = 1
				enters := data[13]
				if diff := enters - mem.locks[0]; diff > 0 {
					sendEvent(ctx, pid, counterType, &messages.Event{
						Id: id, Type: messages.TAMPERING, Value: enters})
				}
				mem.locks[0] = enters
			} else {
				enters := data[14]
				if diff := enters - mem.locks[1]; diff > 0 {
					sendEvent(ctx, pid, counterType, &messages.Event{
						Id: id, Type: messages.TAMPERING, Value: enters})
				}
				mem.locks[1] = enters
			}
		}
	}
	mem.raw = data
}

func sendEvent(ctx actor.Context, pid *actor.PID, tp int, msg *messages.Event) {
	idPuerta := id
	switch {
	case tp == 2 && mgs.Id == 0:
		idPuerta = 1
	case tp == 2 && mgs.Id == 1:
		return
	}
	ctx.Send(pid, msg)
}
