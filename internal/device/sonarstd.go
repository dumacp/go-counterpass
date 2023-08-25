package device

import (
	"time"

	"github.com/dumacp/sonar/protostd"
)

func NewDeviceSonarEvents(port string, speed int) (Device, error) {
	dev := protostd.NewDevice(port, speed, 3*time.Second)

	if err := dev.Open(); err != nil {
		return nil, err
	}

	// fmt.Printf("device: %T\n", dev)

	return dev, nil
}
