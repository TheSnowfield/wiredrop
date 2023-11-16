package peer

import "net/http"

type Role int

const (
	TxPeer Role = 0
	RxPeer Role = 1
)

type Context struct {
	role    Role
	address string
	request *http.Request
	writer  http.ResponseWriter
	shared  Shared
}

type Shared struct {
	Stream      chan []byte
	Buffer      []byte
	Information Information
	Progress    Progress
	Control     Control
}

type Information struct {
	Key        string
	Url        string
	ServerSign string
}

type Progress struct {
	Read    int64
	Wrote   int64
	Maxsize int64
}
type Control struct {
	Done  chan bool
	Abort bool
}
