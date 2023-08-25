package device

import "fmt"

type Device interface {
	Close() error
}

func NewDevice(port string, speed int) (Device, error) {

	switch VendorCounter {
	case "beane":
		return NewDeviceBaene(port, speed)
	case "sonar":
		return NewDeviceSonar(port, speed)
	case "sonarstd":
		return NewDeviceSonarEvents(port, speed)
	case "extreme":
		return NewDeviceExtreme(port, speed)
	case "optocontrol-v301":
		return NewDeviceOptocontrol_v301(port, speed)
	case "optocontrol":
		return NewDeviceOptocontrol(port, speed)
	case "ingnovus-st300":
		return NewDeviceSt300(port, speed)
	case "ingnovus-st3310":
		return NewDeviceSt3310(port, speed)
	case "ingnovus-st3310H04":
		return NewDeviceSt3310_H04(port, speed)
	case "ingnovus":
		return NewDeviceIngnovus(port, speed)
	default:
		return nil, fmt.Errorf("vendor device is invalid (%s)", VendorCounter)
	}
}
