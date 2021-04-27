package listen

import (
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"github.com/dumacp/go-counterpass/logs"
	"github.com/dumacp/go-counterpass/messages"
	"github.com/dumacp/turnstilene"
	"github.com/tarm/serial"
)

type IoDev struct {
	dev turnstilene.DeviceIO
	ID  int
}

type BeaDevice struct {
	front     IoDev
	back      IoDev
	chProcess chan []byte
}

func NewDev(port string, baudRate int) (Dev, error) {
	config := &serial.Config{
		Name:        port,
		Baud:        baudRate,
		ReadTimeout: 600 * time.Second,
	}

	p, err := serial.OpenPort(config)
	if err != nil {
		return nil, err
	}

	ioFront, err := turnstilene.NewDeviceIO(p)
	if err != nil {
		return nil, err
	}
	ioFront.SetAddress(0x81)

	ioBack, err := turnstilene.NewDeviceIO(p)
	if err != nil {
		return nil, err
	}
	ioBack.SetAddress(0x82)

	dev := &BeaDevice{
		front:     IoDev{ioFront, 0x81},
		back:      IoDev{ioBack, 0x82},
		chProcess: make(chan []byte, 0),
	}

	return dev, nil
}

func (dev *BeaDevice) SendData(data []byte) {}

type mem struct {
	inputA  uint32
	inputB  uint32
	failure uint32
	battery bool
	alarm   bool
}

func (d *BeaDevice) ProcessData(data []byte) {}

//Listen listen device
func (d *BeaDevice) Listen(quit chan int) chan *Event {

	first := true

	lenData := 14 //length data

	memArray := make([]byte, lenData)
	t1 := time.Tick(1 * time.Second)
	ch := make(chan *Event, 0)

	memFront := new(mem)
	memBack := new(mem)

	funcRead := func(io IoDev, memInputs *mem) error {
		resp, err := io.dev.ReadData(byte(0x10), lenData)
		//logs.LogBuild.Printf("sendframe resp: [% X]\n", resp)
		if err != nil {
			logs.LogError.Println(err)
			return err
		}

		eq := true
		for i, v := range resp {
			if v != memArray[i] {
				eq = false
				break
			}
		}
		for i, v := range resp {
			memArray[i] = v
		}
		if eq {
			return nil
		}
		if len(resp) < lenData {
			return fmt.Errorf("len data is wrong: [%X]", resp)
		}
		inputA := binary.LittleEndian.Uint32(resp[0:4])
		inputB := binary.LittleEndian.Uint32(resp[4:8])
		failure := binary.LittleEndian.Uint32(resp[8:12])

		alarm := false
		if resp[12] > 0x00 {
			alarm = true
		}
		battery := false
		if resp[13] > 0x00 {
			battery = true
		}
		if first {
			memInputs.inputA = inputA
			memInputs.inputB = inputB
			memInputs.failure = failure
			memInputs.battery = battery
			memInputs.alarm = alarm
			first = false
		}
		if inputA != memInputs.inputA {
			select {
			case ch <- &Event{
				Type:  messages.INPUT,
				ID:    io.ID,
				Value: int64(inputA),
			}:
			case <-time.After(3 * time.Second):
				log.Println("timeout send event")
			}
			memInputs.inputA = inputA
		}
		if inputB != memInputs.inputB {
			select {
			case ch <- &Event{
				Type:  messages.OUTPUT,
				ID:    io.ID,
				Value: int64(inputB),
			}:
			case <-time.After(3 * time.Second):
				log.Println("timeout send event")
			}
			memInputs.inputB = inputB
		}
		if failure != memInputs.failure {
			select {
			case ch <- &Event{
				Type:  messages.TAMPERING,
				ID:    io.ID,
				Value: int64(failure),
			}:
			case <-time.After(3 * time.Second):
				log.Println("timeout send event")
			}
			memInputs.failure = failure
		}
		if battery != memInputs.battery {
			logs.LogWarn.Printf("BATTERY alarm, ID: %X, value: %v", io.ID, battery)
			memInputs.battery = battery
		}
		if alarm != memInputs.alarm {
			logs.LogWarn.Printf("ALARM alarm, ID: %X, value: %v", io.ID, alarm)
			memInputs.alarm = alarm
		}
		return nil
	}

	go func() {
		defer close(ch)
		for {
			select {
			case <-t1:
				funcRead(d.front, memFront)
				funcRead(d.back, memBack)
			case <-quit:
				return
			}
		}
	}()
	return ch
}
