// +build sonar

package device

import (
	"time"

	"github.com/dumacp/sonar/contador"
)

func NewDevice(port string, speed int) (interface{}, error) {
	dev, err := contador.NewDevice(port, speed, 3*time.Second)
	if err != nil {
		return nil, err
	}

	return dev, nil
}
