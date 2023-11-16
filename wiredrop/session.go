package wiredrop

import (
	"wiredrop/wiredrop/peer"
)

type Session struct {
	Timeout float64
	Peers   struct {
		Tx     *peer.Context
		Rx     *peer.Context
		Shared peer.Shared
	}
	ClientNum int
}
