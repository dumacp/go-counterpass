package listen

import (
	"context"
	"fmt"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-counterpass/internal/device"
	"github.com/dumacp/go-ingnovus"
	"github.com/dumacp/go-ingnovus/st300"
	"github.com/dumacp/go-ingnovus/st3310"
	"github.com/dumacp/go-ingnovus/st3310H04"
	"github.com/dumacp/go-levis"
	logirastreo "github.com/dumacp/go-logirastreo/v1"
	"github.com/dumacp/go-optocontrol"
	"github.com/dumacp/sonar/contador"
	"github.com/dumacp/sonar/ins50"
	"github.com/dumacp/sonar/protostd"
)

const (
	max_error = 5
)

var timeout_samples int

func Listen(dev interface{}, contxt context.Context, ctx actor.Context, typeCounter int, externalConsole bool) error {
	timeout_samples = device.TimeoutSamples
	switch dev.(type) {
	case *device.IrmaDevice:
		return ListenIrma(dev, contxt, ctx, typeCounter, externalConsole)
	case protostd.Device:
		return ListenSonarEvents(dev, contxt, ctx, typeCounter, externalConsole)
	case *contador.Device:
		return ListenSonar(dev, contxt, ctx, typeCounter, externalConsole)
	case levis.Device:
		return ListenBeane(dev, contxt, ctx, typeCounter, externalConsole)
	case st300.Device:
		return ListenSt300(dev, contxt, ctx, typeCounter, externalConsole)
	case st3310.Device:
		return ListenSt3310(dev, contxt, ctx, typeCounter, externalConsole)
	case st3310H04.Device:
		return ListenSt3310_h04(dev, contxt, ctx, typeCounter, externalConsole)
	case ingnovus.Device:
		return ListenIngnovus(dev, contxt, ctx, typeCounter, externalConsole)
	case logirastreo.Device:
		return ListenLogirastreo(dev, contxt, ctx, typeCounter, externalConsole)
	case ins50.Device:
		return ListenExtreme(dev, contxt, ctx, typeCounter, externalConsole)
	case optocontrol.Device:
		return ListenOptocontrol(dev, contxt, ctx, typeCounter, externalConsole)
	case optocontrol.Device_v301:
		return ListenOptocontrolV301(dev, contxt, ctx, typeCounter, externalConsole)
	}
	return fmt.Errorf("vendor device is invalid (%T)", dev)
}
