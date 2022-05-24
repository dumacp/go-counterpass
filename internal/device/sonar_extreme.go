// +build extreme

package device

import (
	"github.com/dumacp/sonar/ins50"
)

func NewDevice(port string, speed int) (interface{}, error) {
	dev, err := ins50.NewDevice(port, speed)
	if err != nil {
		return nil, err
	}

	return dev, nil
}
