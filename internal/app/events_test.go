package app

import (
	"reflect"
	"testing"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func Test_buildEventAnomalies(t *testing.T) {
	type args struct {
		ctx    actor.Context
		id     int32
		value  int64
		pidGps *actor.PID
		puerta map[uint]uint
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test1",
			args: args{
				ctx:    nil,
				id:     0,
				value:  1,
				pidGps: nil,
				puerta: map[uint]uint{0: 1, 1: 2},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildEventAnomalies(tt.args.ctx, tt.args.id, tt.args.value, tt.args.pidGps, tt.args.puerta, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildEventAnomalies() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildEventAnomalies() = %s, want %s", got, tt.want)
			}
		})
	}
}
