package peer

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func Create(req *http.Request, w http.ResponseWriter, role Role) *Context {
	return &Context{request: req, writer: w, address: req.RemoteAddr, role: role}
}

func GetAddress(ctx *Context) string {
	return ctx.address
}

func StartTransmission(ctx *Context) {

	switch ctx.role {
	case TxPeer:
		peerTx(ctx)
		break

	case RxPeer:
		peerRx(ctx)
		break
	}
}

func SetSharedData(ctx *Context, shared Shared) {
	ctx.shared = shared
}

func peerTx(ctx *Context) {

	// get raw connection
	netconn, httpconn, err := ctx.writer.(http.Hijacker).Hijack()
	if err != nil {
		log.Println(err)
	}

	// start file transmission
	filename := strings.Split(ctx.shared.Information.Key, "/")
	httpconn.Write([]byte("HTTP/1.1 200 OK\r\n"))
	httpconn.Write([]byte("Server: " + ctx.shared.Information.ServerSign + "\r\n"))
	httpconn.Write([]byte("Content-Type: application/octet-stream\r\n"))
	httpconn.Write([]byte("Content-Disposition: " + "attachment; filename=\"" + filename[len(filename)-1] + "\"\r\n"))
	if ctx.shared.Progress.Maxsize != -1 {
		httpconn.Write([]byte("Content-Length: " + fmt.Sprintf("%d", ctx.shared.Progress.Maxsize)))
	}
	httpconn.Write([]byte("\r\n\r\n"))
	httpconn.Flush()

	// write stream data
	for block := range ctx.shared.Stream {

		blockWrote := 0
		for {
			wrote, _ := httpconn.Write(block)
			blockWrote += wrote

			if blockWrote == len(block) {
				break
			}
		}

		//log.Println(fmt.Sprintf("wrote %d bytes to peer [%s]", len(block), conn.peer.tx))

		ctx.shared.Progress.Wrote += int64(blockWrote)
		if ctx.shared.Progress.Wrote >= ctx.shared.Progress.Maxsize || ctx.shared.Control.Abort {
			break
		}

		ctx.shared.Control.Done <- true
	}

	netconn.Close()
	ctx.shared.Control.Done <- true
	log.Println(fmt.Sprintf("connection closed. "+
		"%d of %d bytes transmitted", ctx.shared.Progress.Wrote, ctx.shared.Progress.Maxsize))
}

func peerRx(ctx *Context) {

	log.Println(fmt.Sprintf("peer [%s] reported "+
		"the stream size is %d bytes", ctx.address, ctx.shared.Progress.Maxsize))

	// get raw connection
	netconn, httpconn, err := ctx.writer.(http.Hijacker).Hijack()
	if err != nil {
		log.Println(err)
	}

	// send 100-continue header
	if expect := ctx.request.Header.Get("Expect"); expect == "100-continue" {
		httpconn.Write([]byte("HTTP/1.1 100 Continue\r\n\r\n"))
		log.Println(fmt.Sprintf("reply to peer [%s] with Http/1.1 100 Continue", ctx.address))
	}

	// read the buffer from PUT stream
	readEof := false
	for {

		read, err := httpconn.Read(ctx.shared.Buffer)

		// write data to another peer
		if read != 0 {
			ctx.shared.Stream <- ctx.shared.Buffer[0:read]
			<-ctx.shared.Control.Done

			// increase buffer size
			if read == len(ctx.shared.Buffer) {
				ctx.shared.Buffer = make([]byte, len(ctx.shared.Buffer)*2)
			}
		}

		// end of file
		ctx.shared.Progress.Read += int64(read)
		if err == io.EOF || (ctx.shared.Progress.Read >= ctx.shared.Progress.Maxsize && ctx.shared.Progress.Maxsize != -1) {
			readEof = true
			break
		} else if err != nil {
			log.Println(err)
		}
	}

	// transport finished
	if readEof {
		httpconn.Write([]byte("HTTP/1.1 201 Created\r\n"))
		httpconn.Write([]byte("Server: " + ctx.shared.Information.ServerSign + "\r\n"))
		httpconn.Write([]byte("Location: " + ctx.shared.Information.Url + "\r\n"))
		httpconn.Write([]byte("Content-Length: 0\r\n"))
		httpconn.Write([]byte("Connection: keep-alive\r\n"))
		httpconn.Write([]byte("\r\n\r\n"))
		httpconn.Flush()
		netconn.Close()
		close(ctx.shared.Stream)
		close(ctx.shared.Control.Done)
	}
}
