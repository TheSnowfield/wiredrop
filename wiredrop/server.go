package wiredrop

import (
	"fmt"
	"log"
	"net/http"
	"time"
	"wiredrop/wiredrop/peer"
)

const (
	defaultServerSign            = "wiredrop/1.0"
	defaultPeerTimeout           = 60 * time.Second
	defaultPeerInitialBufferSize = uint64(4096)
)

var reqTable = make(map[string]*Session)
var peerTimeout = defaultPeerTimeout
var peerInitBufSize = defaultPeerInitialBufferSize
var serverSign = defaultServerSign

// Start creates a http server
func Start(listen string) error {
	http.HandleFunc("/", internalHandler)
	http.HandleFunc("/favicon.ico", faviconHandler)
	return http.ListenAndServe(listen, nil)
}

// StartTLS creates a https server,
// The `cert` and `key` is the path to SSL certificate and key location.
func StartTLS(listen string, cert string, key string) error {
	log.Println("server started at :443")
	http.HandleFunc("/", internalHandler)
	http.HandleFunc("/favicon.ico", faviconHandler)
	return http.ListenAndServeTLS(listen, cert, key, nil)
}

// SetPeerTimeout set the waiting timeout of the peer. If peer connection timed out,
// server return HTTP 408 Request Timeout to peer.
func SetPeerTimeout(duration time.Duration) {
	peerTimeout = duration
}

func SetPeerInitialBufferSize(size uint64) {
	peerInitBufSize = size
}

func faviconHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(404)
}

func internalHandler(w http.ResponseWriter, req *http.Request) {

	// grab the transport key and http method
	var key = req.URL.Path[1:]
	var method = req.Method
	var conn, exist = reqTable[key]

	// only allow GET and PUT method
	if method != http.MethodGet && method != http.MethodPut {
		httpErrorBody(w, http.StatusMethodNotAllowed)
		return
	}

	// the empty key given
	if key == "" {
		httpErrorBody(w, http.StatusBadRequest)
		return
	}

	// if it does not exist, create a new task
	if !exist {
		conn = new(Session)
		conn.Peers.Shared = peer.Shared{
			Buffer: make([]byte, peerInitBufSize),
			Stream: make(chan []byte),
			Control: peer.Control{
				Done:  make(chan bool),
				Abort: false,
			},
			Progress: peer.Progress{
				Read:    0,
				Wrote:   0,
				Maxsize: 0,
			},
			Information: peer.Information{
				ServerSign: serverSign,
				Key:        "",
				Url:        "",
			},
		}
		reqTable[key] = conn
	}

	var ctx *peer.Context

	switch method {
	case http.MethodGet:
		ctx = peer.Create(req, w, peer.TxPeer)
		conn.Peers.Tx = ctx
		break
	case http.MethodPut:
		ctx = peer.Create(req, w, peer.RxPeer)
		conn.Peers.Rx = ctx
		conn.Peers.Shared.Progress.Maxsize = req.ContentLength
		conn.Peers.Shared.Information.Url = req.Host + req.RequestURI
		break
	}

	// wait until all peers come
	for !waitPeerConnection(conn) {
		log.Print(fmt.Sprintf("[%s] waiting for peer connection", req.RemoteAddr))

		// sleep
		time.Sleep(1 * time.Second)

		// timeout counter
		conn.Timeout += 1
		if conn.Timeout >= peerTimeout.Seconds() {
			delete(reqTable, key)
			log.Print("peer timed out")
			httpErrorBody(w, http.StatusRequestTimeout)
			return
		}
	}

	conn.Timeout = 0

	// client connections limit
	if conn.ClientNum < 2 {
		conn.ClientNum += 1

		log.Print(fmt.Sprintf("peer connection established "+
			"%s -> %s", peer.GetAddress(conn.Peers.Tx), peer.GetAddress(conn.Peers.Rx)))

		// start stream transport
		peer.SetSharedData(ctx, conn.Peers.Shared)
		peer.StartTransmission(ctx)
		defer delete(reqTable, key)
	} else {

		// refuse the new connection when started
		httpErrorBody(w, http.StatusTooManyRequests)
		return
	}
}

func waitPeerConnection(sess *Session) bool {
	return sess.Peers.Tx != nil && sess.Peers.Rx != nil
}

func httpErrorBody(w http.ResponseWriter, code int) {

	// write code first
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Server", serverSign)
	w.WriteHeader(code)

	// write html body
	_, err := w.Write([]byte("<html>\n" +
		"<head><title>Error</title></head>\n" +
		"<body>\n" +
		"<center><h1>" + fmt.Sprintf("%d %s", code, http.StatusText(code)) + "</h1></center>\n" +
		"<hr><center>" + serverSign + "</center>\n" +
		"</body>\n" +
		"</html>"))

	if err != nil {
		log.Println(err)
	}
}
