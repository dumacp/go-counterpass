package listen

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/dumacp/go-counterpass/internal/device"
	"github.com/dumacp/go-counterpass/messages"
	"github.com/dumacp/go-logs/pkg/logs"
)

func ListenIrma(dev interface{}, quit context.Context, ctx actor.Context, typeCounter int, externalConsole bool) error {

	fmt.Println("listenIrma")
	var devv *device.IrmaDevice

	if v, ok := dev.(*device.IrmaDevice); !ok {
		return fmt.Errorf("device is not irma device, %T", dev)
	} else {
		devv = v
	}

	fmt.Printf("irma dvice: %v\n", devv)

	self := ctx.Self()
	rootctx := ctx.ActorSystem().Root

	go func(ctx *actor.RootContext, self *actor.PID) {
		defer func() {
			ctx.Send(self, &MsgListenError{ID: 1})
			// ctx.Send(self, &MsgListenError{ID: 1})
		}()
		tick := time.NewTicker(5 * time.Second)
		defer tick.Stop()
		for {
			select {
			case <-quit.Done():
				return
			case <-tick.C:
				if err := processIrmaData(ctx, self, 1, devv.Url); err != nil {
					fmt.Printf("error command irma: %s\n", err)
				}

			}

		}
	}(rootctx, self)

	return nil
}

func processIrmaData(ctx *actor.RootContext, pid *actor.PID, id int32, url string) error {

	dataIrma := struct {
		Fa         int `json:"fa"`
		CategoryId int `json:"category"`
		In         int `json:"in"`
		Out        int `json:"out"`
	}{}

	command := "/usr/sbin/omvz7/countIrma"

	out, err := exec.Command(command, "--reset", "--url", url).Output()
	if err != nil {
		return err
	}

	if err := json.Unmarshal(out, &dataIrma); err != nil {
		return err
	}

	logs.LogBuild.Printf("irma data: %s", out)

	if dataIrma.In > 0 {
		ctx.Send(pid, &messages.Event{Id: id, Type: messages.Event_INPUT, DiffValue: int32(dataIrma.In)})
	}
	if dataIrma.Out > 0 {
		ctx.Send(pid, &messages.Event{Id: id, Type: messages.Event_OUTPUT, DiffValue: int32(dataIrma.Out)})
	}
	return nil
}
