package listen

type MsgListenError struct {
	ID int
}
type MsgToTest struct {
	Data []byte
}

type MsgLogRequest struct{}
type MsgLogResponse struct {
	Value []byte
}
type Subscribe struct{}
