//go:build st300
// +build st300

package device

import (
	"time"

	"github.com/dumacp/go-ingnovus/st300"
)

func NewDevice(port string, speed int) (Device, error) {
	dev := st300.NewDevice(port, speed, 600*time.Millisecond)

	if err := dev.Open(); err != nil {
		return nil, err
	}

	return dev, nil
}
