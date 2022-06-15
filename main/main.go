package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"fmt"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/persistence"
	"github.com/dumacp/go-counterpass/internal/app"
	"github.com/dumacp/go-counterpass/internal/device"
	"github.com/dumacp/go-counterpass/internal/doors"
	"github.com/dumacp/go-counterpass/internal/gps"
	"github.com/dumacp/go-counterpass/internal/listen"
	pubs "github.com/dumacp/go-counterpass/internal/pubsub"
	"github.com/dumacp/go-logs/pkg/logs"
)

const (
	showVersion = "2.0.7"
)

var debug bool
var logStd bool
var socket string
var pathdb string
var baudRate int
var version bool

// var loglevel int
var isZeroOpenStateDoor0 bool
var isZeroOpenStateDoor1 bool
var disableDoorGpioListen bool
var typeCounterDoor int
var sendGpsToConsole bool
var simulate bool
var mqtt bool
var vendor bool
var disablePersistence bool
var initCounters string

func init() {
	flag.BoolVar(&debug, "debug", false, "debug enable")
	flag.BoolVar(&mqtt, "disablePublishEvents", false, "disable local to publish local events")
	flag.BoolVar(&disablePersistence, "disablePersistence", false, "disable persistence")
	flag.BoolVar(&logStd, "logStd", false, "log in stderr")
	flag.StringVar(&pathdb, "pathdb", "/SD/boltdbs/countingdb", "socket to listen events")
	flag.StringVar(&socket, "port", "/dev/ttyS2", "serial port")
	flag.StringVar(&initCounters, "initCounters", "", `init counters in database, example: "101 11 202 22"
	101 -> inputs front door,
	11  -> outputs front door,
	202 -> inputs back door,
	22  -> outputs back door`)
	flag.IntVar(&baudRate, "baud", 19200, "baudrate")
	flag.IntVar(&typeCounterDoor, "typeCounterDoor", 0, "0: two counters (front and back), 1: front counter, 2: back counter")
	// flag.IntVar(&loglevel, "loglevel", 0, "level log")
	flag.BoolVar(&version, "version", false, "show version")
	flag.BoolVar(&vendor, "vendor", false, "show supported vendor")
	flag.BoolVar(&isZeroOpenStateDoor0, "zeroOpenStateDoor0", false, "Is Zero the open state in front door?")
	flag.BoolVar(&isZeroOpenStateDoor0, "zeroOpenStateDoor1", false, "Is Zero the open state in back door?")
	flag.BoolVar(&disableDoorGpioListen, "disableDoorGpioListen", false, "disable DoorGpio Listen")
	flag.BoolVar(&sendGpsToConsole, "sendGpsToConsole", false, "Send GPS frame to Sonar console?")
	flag.BoolVar(&simulate, "simulate", false, "Simulate Test data")
}

func main() {

	flag.Parse()

	if typeCounterDoor < 0 || typeCounterDoor > 2 {
		log.Fatalln("wrong typeCounterDoor")
	}

	if version {
		fmt.Printf("version: %s\n", showVersion)
		os.Exit(2)
	}

	if vendor {
		fmt.Printf("vendor: %s\n", app.VendorCounter)
		os.Exit(2)
	}

	initLogs(debug, logStd)

	var provider *provider
	var err error
	if !disablePersistence {
		provider, err = newProvider(pathdb, 10)
		if err != nil {
			log.Fatalln(err)
		}
	}

	rootContext := actor.NewActorSystem().Root

	pubs.Init(rootContext)

	actorGps := gps.NewActor()
	propGps := actor.PropsFromProducer(func() actor.Actor { return actorGps })
	pidGps, err := rootContext.SpawnNamed(propGps, "gps-actor")
	if err != nil {
		logs.LogError.Fatalln(err)
	}

	var pidDoors *actor.PID
	if !disableDoorGpioListen {
		actorDoors := doors.New()
		propsDoors := actor.PropsFromProducer(func() actor.Actor { return actorDoors })
		pidDoors, err = rootContext.SpawnNamed(propsDoors, "doors")
		if err != nil {
			logs.LogError.Println(err)
		}
	}

	actorDevice := device.NewActor(socket, baudRate, 300*time.Millisecond)
	propDevice := actor.PropsFromProducer(func() actor.Actor { return actorDevice })
	pidDevice, err := rootContext.SpawnNamed(propDevice, "device")
	if err != nil {
		logs.LogError.Fatalln(err)
	}

	actorListen := listen.NewListen(typeCounterDoor)
	propListen := actor.PropsFromProducer(func() actor.Actor { return actorListen })
	pidListen, err := rootContext.SpawnNamed(propListen, "listenner")
	if err != nil {
		logs.LogError.Fatalln(err)
	}

	counting := app.NewCountingActor()

	counting.SetZeroOpenStateDoor0(isZeroOpenStateDoor0)
	counting.SetZeroOpenStateDoor1(isZeroOpenStateDoor1)
	counting.DisableDoorGpioListen(disableDoorGpioListen)
	counting.CounterType(typeCounterDoor)
	counting.SetGPStoConsole(sendGpsToConsole)

	propsCounting := actor.PropsFromProducer(func() actor.Actor { return counting })
	if !disablePersistence {
		propsCounting = propsCounting.WithReceiverMiddleware(persistence.Using(provider))
	} else {
		counting.DisablePersistence(true)
	}

	pidCounting, err := rootContext.SpawnNamed(propsCounting, "counting")
	if err != nil {
		logs.LogError.Fatalln(err)
	}

	rootContext.RequestWithCustomSender(
		pidDevice,
		&device.Subscribe{},
		pidListen)

	rootContext.RequestWithCustomSender(
		pidListen,
		&listen.Subscribe{},
		pidCounting)

	rootContext.RequestWithCustomSender(
		pidListen,
		&listen.Subscribe{},
		pidDevice)

	rootContext.RequestWithCustomSender(
		pidCounting,
		&app.MsgRegisterGPS{},
		pidGps)

	if !disableDoorGpioListen {
		rootContext.RequestWithCustomSender(
			pidDoors,
			&doors.Subscribe{},
			pidCounting)
	}

	if mqtt {
		rootContext.Send(pidCounting, &app.MsgSendEvents{Data: false})
	}

	if len(initCounters) > 0 {
		space := regexp.MustCompile(`\s+`)
		s := space.ReplaceAllString(initCounters, " ")
		spplit := strings.Split(s, " ")
		if len(spplit) != 4 {
			log.Fatalln("init error, len initCounter is wrong. Len allow is 4, example -> \"0 0 0 0\"")
		}
		data := make([]int64, 0)
		for _, sv := range spplit {
			v, err := strconv.Atoi(sv)
			if err != nil {
				log.Fatalln(err)
			}
			data = append(data, int64(v))
		}
		rootContext.Send(pidCounting, &app.MsgInitCounters{
			Inputs0:  data[0],
			Inputs1:  data[1],
			Outputs0: data[2],
			Outputs1: data[3]})
		// rootContext.Send(pidCounting, &messages.Event{Type: 0, Value: 160, Id: 0})
		time.Sleep(1 * time.Second)
		rootContext.PoisonFuture(pidCounting).Wait()
		log.Fatalln("database is initialize")
	}

	time.Sleep(3 * time.Second)

	rootContext.Send(pidCounting, &app.MsgSendRegisters{})

	logs.LogInfo.Printf("back door counter START --  version: %s\n", showVersion)

	go func() {
		t1 := time.NewTicker(6 * time.Second)
		defer t1.Stop()
		for range t1.C {
			rootContext.Send(pidCounting, &app.MsgSendRegisters{})
		}
	}()

	finish := make(chan os.Signal, 1)
	signal.Notify(finish, syscall.SIGINT)
	signal.Notify(finish, syscall.SIGTERM)
	<-finish
}
