package listen

import (
	"testing"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-counterpass/internal/device"
)

func TestNewListen(t *testing.T) {

	type args struct {
		devPort     string
		devBaud     int
		devTimeout  time.Duration
		typeCounter int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test1",
			args: args{
				devPort:     "/dev/pts/2",
				devBaud:     19200,
				devTimeout:  30 * time.Second,
				typeCounter: 1,
			},
		},
	}
	for _, tt := range tests {

		rootctx := actor.NewActorSystem().Root

		t.Run(tt.name, func(t *testing.T) {
			actDev := device.NewActor(tt.args.devPort, tt.args.devBaud, tt.args.devTimeout)
			propsDev := actor.PropsFromFunc(actDev.Receive)
			pidDev, err := rootctx.SpawnNamed(propsDev, "devTest")
			if err != nil {
				t.Errorf("error = %s", err)
			}

			act := NewListen(tt.args.typeCounter)
			props := actor.PropsFromFunc(act.Receive)

			pid, err := rootctx.SpawnNamed(props, "listenTest")
			if err != nil {
				t.Errorf("error = %s", err)
			}

			rootctx.RequestWithCustomSender(pidDev, &device.Subscribe{}, pid)

		})
	}

	time.Sleep(5 * 60 * time.Second)
}
