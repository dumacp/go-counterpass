package main

import (
	"encoding/json"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"fmt"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/persistence"
	"github.com/asynkron/protoactor-go/remote"
	"github.com/dumacp/go-counterpass/internal/app"
	"github.com/dumacp/go-counterpass/internal/device"
	"github.com/dumacp/go-counterpass/internal/doors"
	"github.com/dumacp/go-counterpass/internal/gps"
	"github.com/dumacp/go-counterpass/internal/listen"
	"github.com/dumacp/go-counterpass/internal/parameters"
	pubs "github.com/dumacp/go-counterpass/internal/pubsub"
	"github.com/dumacp/go-counterpass/internal/utils"
	"github.com/dumacp/go-logs/pkg/logs"
	"github.com/dumacp/go-params/api/params"
)

const (
	showVersion = "2.4.0"
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
var enableDoorVerify bool
var disableDoorGpioListen bool
var typeCounterDoor int = -1
var sendGpsToConsole bool
var simulate bool
var mqtt bool
var disablePersistence bool
var initCounters string
var id string

func init() {
	flag.StringVar(&id, "id", "", "device ID")
	flag.BoolVar(&debug, "debug", false, "debug enable")
	flag.BoolVar(&mqtt, "disablePublishEvents", false, "disable local to publish local events")
	flag.BoolVar(&disablePersistence, "disablePersistence", false, "disable persistence")
	flag.BoolVar(&logStd, "logStd", false, "log in stderr")
	flag.StringVar(&pathdb, "pathdb", "/SD/boltdbs/countingdb", "socket to listen events")
	flag.StringVar(&socket, "port", "", "serial port (default \"/dev/ttyS2\")")
	flag.StringVar(&initCounters, "initCounters", "", `init counters in database, example: "101 11 202 22"
	101 -> inputs front door,
	11  -> outputs front door,
	202 -> inputs back door,
	22  -> outputs back door`)
	flag.IntVar(&baudRate, "baud", 0, fmt.Sprintf("baudrate (default %d)", baud_defaut))
	flag.IntVar(&typeCounterDoor, "typeCounterDoor", 0, "0: two counters (front and back), 1: front counter, 2: back counter")
	// flag.IntVar(&loglevel, "loglevel", 0, "level log")
	flag.BoolVar(&version, "version", false, "show version")
	// flag.BoolVar(&vendor, "vendor", false, "show supported vendor")
	flag.BoolVar(&isZeroOpenStateDoor0, "zeroOpenStateDoor0", false, "Is Zero the open state in front door?")
	flag.BoolVar(&isZeroOpenStateDoor0, "zeroOpenStateDoor1", false, "Is Zero the open state in back door?")
	flag.BoolVar(&enableDoorVerify, "enableDoorVerify", false, "enable the verification function for the door in the counter system")
	flag.BoolVar(&disableDoorGpioListen, "disableDoorGpioListen", false, "DEPRECATED")
	flag.BoolVar(&sendGpsToConsole, "sendGpsToConsole", false, "Send GPS frame to Sonar console?")
	flag.BoolVar(&simulate, "simulate", false, "Simulate Test data")
}

func main() {

	flag.Parse()
	// if vendor {
	// 	fmt.Printf("vendor: %s\n", app.VendorCounter)
	// 	os.Exit(2)
	// }

	getENV()

	if version {
		fmt.Printf("version: %s\n", showVersion)
		os.Exit(2)
	}

	if typeCounterDoor < 0 || typeCounterDoor > 2 {
		log.Fatalln("wrong typeCounterDoor")
	}

	initLogs(debug, logStd)

	if len(id) <= 0 {
		id = utils.Hostname()
	} else {
		utils.SetHostname(id)
	}

	// verify and create path database
	if _, err := os.Stat(utils.DATABASE_PATH); os.IsNotExist(err) {
		if err := os.MkdirAll(utils.DATABASE_PATH, 0755); err != nil {
			log.Fatalf("error creating database directory: %s", err)
		}
	}

	var err error
	var provider *provider
	if !disablePersistence {
		provider, err = newProvider(pathdb, 10)
		if err != nil {
			log.Fatalln(err)
		}
	}

	sys := actor.NewActorSystem()
	rootContext := sys.Root

	pubs.Init(rootContext)

	var pidCounting *actor.PID
	var pidDevice *actor.PID
	var pidGps *actor.PID
	var pidListen *actor.PID
	var pidDoors *actor.PID
	var errorCount int
	// var errorCountPidCounting int
	// var errorCountPidListen int
	var maxSleepDuration = 5 * time.Minute

	type initActorCounting struct{}
	type initActorListen struct{}
	type tickMsg struct{}

	props := actor.PropsFromFunc(func(ctx actor.Context) {

		switch msg := ctx.Message().(type) {
		case *actor.Started:

			if err := func() error {

				actorGps := gps.NewActor()
				propGps := actor.PropsFromProducer(func() actor.Actor { return actorGps })
				pidGps, err = ctx.SpawnNamed(propGps, "gps-actor")
				if err != nil {
					logs.LogError.Printf("gps actor error: %s", err)
					return err
				}

				if _, err := ctx.SpawnNamed(actor.PropsFromFunc(parameters.NewActor(id,
					fmt.Sprintf("%s/%s", utils.DATABASE_PATH, utils.DATABASE_PARAMS_NAME), 60*time.Minute).Receive), "params-actor"); err != nil {
					log.Fatalf("params actor error: %s", err)
				}

				// var pidDoors *actor.PID
				if !disableDoorGpioListen {
					actorDoors := doors.NewActor()
					propsDoors := actor.PropsFromProducer(func() actor.Actor { return actorDoors })
					pidDoors, err = ctx.SpawnNamed(propsDoors, "doors")
					if err != nil {
						logs.LogError.Printf("doors actor error: %s", err)
						return err
					}
				}

				// Reset error count on successful initialization
				errorCount = 0
				ctx.Send(ctx.Self(), &initActorListen{})
				ctx.Send(ctx.Self(), &initActorCounting{})
				return nil
			}(); err != nil {
				// Increment error count
				errorCount++

				// Calculate sleep duration with exponential backoff
				// Start with 3 seconds, then 15 seconds, 30 seconds, up to 5 minutes
				var sleepDuration time.Duration
				switch {
				case errorCount == 1:
					sleepDuration = 3 * time.Second
				case errorCount == 2:
					sleepDuration = 15 * time.Second
				case errorCount == 3:
					sleepDuration = 30 * time.Second
				default:
					sleepDuration = maxSleepDuration
				}

				logs.LogError.Printf("error in counterpass started actor (attempt %d): %s, sleeping for %v", errorCount, err, sleepDuration)
				time.Sleep(sleepDuration)

				// After max attempts, panic
				if errorCount >= 10 {
					logs.LogError.Panicf("error in counterpass started actor after %d attempts: %s", errorCount, err)
				}
			}
			logs.LogInfo.Printf("started \"%s\", %v", ctx.Self().GetId(), ctx.Self())
		case *initActorListen:
			if err := func() error {

				if pidListen != nil {
					ctx.PoisonFuture(pidListen).Wait()
				}
				if pidDevice != nil {
					ctx.PoisonFuture(pidDevice).Wait()
				}

				actorDevice := device.NewActor(socket, baudRate, 300*time.Millisecond)
				propDevice := actor.PropsFromProducer(func() actor.Actor { return actorDevice })
				pidDevice, err = ctx.SpawnNamed(propDevice, "device")
				if err != nil {
					logs.LogError.Printf("device actor error: %s", err)
					return err
				}

				actorListen := listen.NewListen(typeCounterDoor)
				propListen := actor.PropsFromProducer(func() actor.Actor { return actorListen })
				pidListen, err = ctx.SpawnNamed(propListen, "listenner")
				if err != nil {
					logs.LogError.Printf("listen actor error: %s", err)
					return err
				}
				return nil
			}(); err != nil {
				// Increment error count
				logs.LogError.Printf("error in counterpass started listen actor: %s", err)
			}
		case *initActorCounting:
			if err := func() error {

				if pidDevice == nil || pidListen == nil {
					logs.LogError.Println("device or listen actor not initialized")
					return fmt.Errorf("device or listen actor not initialized")
				}

				if pidCounting != nil {
					ctx.PoisonFuture(pidCounting).Wait()
				}

				counting := app.NewCountingActor(device.VendorCounter)

				counting.SetZeroOpenStateDoor0(isZeroOpenStateDoor0)
				counting.SetZeroOpenStateDoor1(isZeroOpenStateDoor1)
				counting.EnableDoorVerify(enableDoorVerify)
				counting.CounterType(typeCounterDoor)
				counting.SetGPStoConsole(sendGpsToConsole)

				propsCounting := actor.PropsFromProducer(func() actor.Actor { return counting })
				if !disablePersistence {
					propsCounting = propsCounting.Configure(actor.WithReceiverMiddleware(persistence.Using(provider)))
				} else {
					counting.DisablePersistence(true)
				}

				pidCounting, err = ctx.SpawnNamed(propsCounting, "counting")
				if err != nil {
					logs.LogError.Printf("counting actor error: %s", err)
					return err
				}

				ctx.RequestWithCustomSender(
					pidDevice,
					&device.Subscribe{},
					pidListen)

				ctx.RequestWithCustomSender(
					pidListen,
					&listen.Subscribe{},
					pidCounting)

				ctx.RequestWithCustomSender(
					pidListen,
					&listen.Subscribe{},
					pidDevice)

				ctx.RequestWithCustomSender(
					pidCounting,
					&app.MsgRegisterGPS{},
					pidGps)

				ctx.RequestWithCustomSender(
					pidDoors,
					&doors.Subscribe{},
					pidCounting)

				if mqtt {
					ctx.Send(pidCounting, &app.MsgSendEvents{Data: false})
				}
				return nil
			}(); err != nil {
				logs.LogError.Printf("error in counterpass started counting actor: %s", err)
			}

		case *actor.Stopping:
			logs.LogWarn.Printf("\"%s\" - Stopped actor, reason -> %v", ctx.Self(), msg)
		case *actor.Restarting:
			logs.LogWarn.Printf("\"%s\" - Restarting actor, reason -> %v", ctx.Self(), msg)
		case *actor.Terminated:
			logs.LogWarn.Printf("\"%s\" - Terminated actor, reason -> %v", ctx.Self(), msg)
		case *app.MsgInitCounters:
			if pidCounting != nil {
				ctx.Send(pidCounting, msg)
			}
		case *app.MsgSendRegisters:
			if pidCounting != nil {
				ctx.Send(pidCounting, msg)
			}
		case *tickMsg:
			if pidListen == nil {
				ctx.Send(ctx.Self(), &initActorListen{})
			}
			if pidCounting == nil {
				ctx.Send(ctx.Self(), &initActorCounting{})
			}
		case *params.Parameters:
			if len(msg.CounterpassConf) > 0 {
				logs.LogInfo.Printf("Counterpass config: %s", msg.CounterpassConf)
				conf := new(parameters.Configuration)

				if err := json.Unmarshal([]byte(msg.CounterpassConf), conf); err != nil {
					logs.LogError.Printf("error marshal counterpass config: %s", err)
				} else {
					changeDevice := false
					if len(conf.Port) > 0 && socket != conf.Port {
						socket = conf.Port
						changeDevice = true
						logs.LogInfo.Printf("socket changed to: %s", socket)
					}
					if conf.Baud > 0 && baudRate != conf.Baud {
						baudRate = conf.Baud
						changeDevice = true
						logs.LogInfo.Printf("baudrate changed to: %d", baudRate)
					}
					if changeDevice {
						ctx.Send(ctx.Self(), &initActorListen{})
					}
					changeDevice = false
					if device.VendorCounter != conf.Vendor {
						if err := device.VerifyVendor(conf.Vendor); err != nil {
							logs.LogError.Printf("error verify vendor: %s", err)
						} else {
							changeDevice = true
							logs.LogInfo.Printf("vendor changed to: %s", conf.Vendor)
						}
					}
					if typeCounterDoor != conf.Type {
						if conf.Type < 0 || conf.Type > 2 {
							logs.LogError.Printf("wrong typeCounterDoor: %d", conf.Type)
						} else {
							typeCounterDoor = conf.Type
							changeDevice = true
							logs.LogInfo.Printf("typeCounterDoor changed to: %d", conf.Type)
						}
					}
					if isZeroOpenStateDoor0 != conf.IsZeroOpenStateDoor0 {
						isZeroOpenStateDoor0 = conf.IsZeroOpenStateDoor0
						changeDevice = true
						logs.LogInfo.Printf("isZeroOpenStateDoor0 changed to: %t", conf.IsZeroOpenStateDoor0)
					}
					if isZeroOpenStateDoor1 != conf.IsZeroOpenStateDoor1 {
						isZeroOpenStateDoor1 = conf.IsZeroOpenStateDoor1
						changeDevice = true
						logs.LogInfo.Printf("isZeroOpenStateDoor1 changed to: %t", conf.IsZeroOpenStateDoor1)
					}
					if enableDoorVerify != conf.VerifyOpenDoor {
						enableDoorVerify = conf.VerifyOpenDoor
						changeDevice = true
						logs.LogInfo.Printf("enableDoorVerify changed to: %t", conf.VerifyOpenDoor)
					}
					if changeDevice {
						ctx.Send(ctx.Self(), &initActorCounting{})
					}
				}
			} else {
				logs.LogWarn.Println("Counterpass config is empty")
			}
		}
	})

	portlocal := 8099
	for {

		socket := fmt.Sprintf("127.0.0.1:%d", portlocal)
		testConn, err := net.DialTimeout("tcp", socket, 1*time.Second)
		if err != nil {
			break
		}
		testConn.Close()
		logs.LogWarn.Printf("socket busy -> \"%s\"", socket)
		time.Sleep(1 * time.Second)
		portlocal++
	}

	// kind := remote.NewKind("driverconsole", props)
	rconfig := remote.Configure("127.0.0.1", portlocal)
	//	remote.NewKind("driverconsole", props))
	r := remote.NewRemote(sys, rconfig)

	var pidMain *actor.PID

	pidMain, err = sys.Root.SpawnNamed(props, "counterpass")
	if err != nil {
		log.Fatalln(err)
	}
	r.Start()

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
		rootContext.Send(pidMain, &app.MsgInitCounters{
			Inputs0:  data[0],
			Inputs1:  data[1],
			Outputs0: data[2],
			Outputs1: data[3]})
		// rootContext.Send(pidCounting, &messages.Event{Type: 0, Value: 160, Id: 0})
		time.Sleep(1 * time.Second)
		rootContext.PoisonFuture(pidMain).Wait()
		log.Fatalln("database is initialize")
	}

	time.Sleep(3 * time.Second)

	rootContext.Send(pidMain, &app.MsgSendRegisters{})

	logs.LogInfo.Printf("back door counter START --  version: %s\n", showVersion)

	go func() {
		t0 := time.NewTimer(5 * time.Second)
		defer t0.Stop()
		t1 := time.NewTicker(5 * time.Second)
		defer t1.Stop()
		t2 := time.NewTicker(30 * time.Second)
		defer t2.Stop()
		for {
			select {
			case <-t0.C:
				rootContext.Send(pidMain, &tickMsg{})
			case <-t1.C:
				rootContext.Send(pidMain, &app.MsgSendRegisters{})
			case <-t2.C:
				rootContext.Send(pidGps, &tickMsg{})
			}
		}
	}()

	finish := make(chan os.Signal, 1)
	signal.Notify(finish, syscall.SIGINT)
	signal.Notify(finish, syscall.SIGTERM)
	<-finish
}
