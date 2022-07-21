package app

import (
	"testing"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-counterpass/messages"
)

func TestNewCountingActor(t *testing.T) {

	rootctx := actor.NewActorSystem().Root

	type args struct {
		data []*messages.Event
	}

	tests := []struct {
		name      string
		args      args
		wantError bool
	}{
		// TODO: Add test cases.
		{
			name: "test1",
			args: args{
				data: []*messages.Event{
					{
						Id:    0,
						Value: 1,
						Type:  messages.OUTPUT,
					},
					{
						Id:    0,
						Value: 2,
						Type:  messages.INPUT,
					},
					{
						Id:    0,
						Value: 3,
						Type:  messages.OUTPUT,
					},
					{
						Id:    0,
						Value: 4,
						Type:  messages.INPUT,
					},
					{
						Id:    0,
						Value: 5,
						Type:  messages.OUTPUT,
					},
					{
						Id:    0,
						Value: 6,
						Type:  messages.INPUT,
					},
					{
						Id:    0,
						Value: 7,
						Type:  messages.OUTPUT,
					},
					{
						Id:    0,
						Value: 8,
						Type:  messages.INPUT,
					},
				},
			},
			wantError: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCountingActor()
			got.DisablePersistence(true)

			props := actor.PropsFromFunc(got.Receive)
			pid, err := rootctx.SpawnNamed(props, tt.name)
			if err != nil && !tt.wantError {
				t.Errorf("NewCountingActor(), error = %s", err)
			}
			time.Sleep(300 * time.Millisecond)
			for i, evt := range tt.args.data {
				rootctx.Send(pid, evt)

				if i == 3 {
					time.Sleep(1 * time.Second)
				} else {
					time.Sleep(3 * time.Second)
				}
			}
		})
	}
	time.Sleep(5 * time.Second)
}
