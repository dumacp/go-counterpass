package device

import (
	"context"
	"fmt"
	"time"

	"github.com/dumacp/go-logs/pkg/logs"
	"github.com/looplab/fsm"
)

const (
	sStart = "sStart"
	sOpen  = "sOpen"
	sClose = "sClose"
	sWait  = "sWait"
	sStop  = "sStop"
)

const (
	eStarted = "eStarted"
	eOpenned = "eOpenned"
	eClosed  = "eClosed"
	eError   = "eError"
	eStop    = "eStop"
)

func beforeEvent(event string) string {
	return fmt.Sprintf("before_%s", event)
}
func enterState(state string) string {
	return fmt.Sprintf("enter_%s", state)
}
func leaveState(state string) string {
	return fmt.Sprintf("leave_%s", state)
}

func (a *Actor) Fsm() {

	var disp Device
	callbacksfsm := fsm.Callbacks{
		"before_event": func(_ context.Context, e *fsm.Event) {
			if e.Err != nil {
				// log.Println(e.Err)
				e.Cancel(e.Err)
			}
		},
		"leave_state": func(_ context.Context, e *fsm.Event) {
			if e.Err != nil {
				// log.Println(e.Err)
				e.Cancel(e.Err)
			}
		},
		"enter_state": func(_ context.Context, e *fsm.Event) {
			logs.LogBuild.Printf("FSM APP DEVICE, state src: %s, state dst: %s", e.Src, e.Dst)
		},
		beforeEvent(eStarted): func(_ context.Context, e *fsm.Event) {
			var err error

			if disp != nil {
				disp.Close()
			}

			for _, t := range []int{0, 3, 3, 10, 10, 10, 10, 10} {
				if t > 0 {
					time.Sleep(time.Duration(t) * time.Second)
				}
				disp, err = NewDevice(a.portSerial, a.speedBaud)
				if err == nil {
					break
				}
				if disp != nil {
					disp.Close()
				}
			}
			if err != nil {
				e.Cancel(err)
				a.ctx.Send(a.ctx.Self(), &StartDevice{})
				return
			}
			a.ctx.Send(a.ctx.Self(), &MsgDevice{Device: disp})
		},
		enterState(sClose): func(_ context.Context, e *fsm.Event) {
			if disp == nil {
				return
			}
			disp.Close()
		},
	}

	f := fsm.NewFSM(
		sStart,
		fsm.Events{
			{
				Name: eStarted,
				Src:  []string{sStart, sClose, sStop},
				Dst:  sOpen,
			},
			{
				Name: eOpenned,
				Src:  []string{sOpen},
				Dst:  sWait,
			},
			{
				Name: eError,
				Src:  []string{sStart, sOpen, sWait},
				Dst:  sClose,
			},
			{
				Name: eStop,
				Src:  []string{sStart, sOpen, sWait},
				Dst:  sStop,
			},
		},
		callbacksfsm,
	)

	a.fmachinae = f
}
