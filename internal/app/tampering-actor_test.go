package app

import (
	"testing"
	"time"

	"github.com/dumacp/go-counterpass/messages"
)

func Test_tamperingFlipFlop(t *testing.T) {
	type args struct {
		quit <-chan int
		data []*messages.Event
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{
			name: "test1",
			args: args{
				quit: make(chan int),
				data: []*messages.Event{
					{
						Id:    0,
						Value: 1,
						Type:  messages.TAMPERING,
					},
					{
						Id:    0,
						Value: 2,
						Type:  messages.TAMPERING,
					},
					{
						Id:    0,
						Value: 3,
						Type:  messages.TAMPERING,
					},
					{
						Id:    1,
						Value: 4,
						Type:  messages.TAMPERING,
					},
					{
						Id:    0,
						Value: 4,
						Type:  messages.TAMPERING,
					},
					{
						Id:    0,
						Value: 5,
						Type:  messages.TAMPERING,
					},
					{
						Id:    0,
						Value: 6,
						Type:  messages.TAMPERING,
					},
					{
						Id:    0,
						Value: 7,
						Type:  messages.TAMPERING,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := tamperingFlipFlop(tt.args.quit)
			for _, v := range tt.args.data {
				ch <- v
				time.Sleep(1 * time.Second)
			}
		})
	}
	time.Sleep(12 * time.Second)
}
