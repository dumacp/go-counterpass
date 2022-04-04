package app

//MsgSendRegisters messages to send registers to pubsub
type MsgSendRegisters struct{}

//MsgInitCounters message to init counter database
type MsgInitCounters struct {
	Inputs0  int64
	Outputs0 int64
	Inputs1  int64
	Outputs1 int64
}

type MsgRegisterGPS struct {
	ID   string
	Addr string
}

type MsgSendEvents struct {
	Data bool
}
