package main

import (
	"context"
	"fmt"
	"io"
	"iris/connmanager"
	"log/slog"
	"net"
	"os"
	"time"
)

var (
	ConnectionCountManager connmanager.ConnCountManager
	ConnectionCountLimit   int           = 1
	AcquireDeadline        time.Duration = time.Second * 30
)

type RelayedConnection struct {
	clientConn  net.Conn
	serverConn  net.Conn
	connManager connmanager.ConnCountManager
}

func NewRelayedConnection(clientConn net.Conn, connManager connmanager.ConnCountManager) *RelayedConnection {
	return &RelayedConnection{
		clientConn:  clientConn,
		serverConn:  nil,
		connManager: connManager,
	}
}

func (r *RelayedConnection) AcquireWithDeadline(ctx context.Context, deadlineDuration time.Time) bool {
	dlCtx, _ := context.WithDeadline(ctx, deadlineDuration)
	return (r.connManager).Acquire(dlCtx)
}

func (r *RelayedConnection) DialServer() error {
	var err error
	r.serverConn, err = net.Dial("tcp", "localhost:3001")
	if err != nil {
		slog.Error("error dialing server", err)
	}
	return err
}

func (r *RelayedConnection) Close() {
	if r.serverConn != nil {
		r.serverConn.Close()
	}
	if r.clientConn != nil {
		r.clientConn.Close()
	}
}

func (r *RelayedConnection) ServerToClient() {
	_, err := io.Copy(r.clientConn, r.serverConn)
	if err != nil {
		slog.Error("error occured while copying from server to client")
		r.Close()
	}
}

func (r *RelayedConnection) ClientToServer() {
	_, err := io.Copy(r.serverConn, r.clientConn)
	if err != nil {
		slog.Error("error occured while copying from client to server")
		r.Close()
	}
}

func (r *RelayedConnection) Relay(ctx context.Context) {
	defer r.Close()
	if r.AcquireWithDeadline(ctx, time.Now().Add(AcquireDeadline)) {
		if r.DialServer() != nil {
			return
		}
		go r.ServerToClient()
		r.ClientToServer()
	}
}

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:3000")
	if err != nil {
		slog.Error("failed to start tcp listener", err)
		os.Exit(1)
	}
	slog.Info(fmt.Sprintf("listening on %s ...", listener.Addr().String()))
	ConnectionCountManager = connmanager.NewSimpleConnCountManager(ConnectionCountLimit)
	AcceptAndServeConnections(listener)
}

func AcceptAndServeConnections(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Error("failed to accept tcp connection", err)
			continue
		}
		slog.Info(fmt.Sprintf("accepted connection from %s", conn.RemoteAddr().String()))
		go ServeConnection(conn)
	}
}

func ServeConnection(clientConn net.Conn) {
	rc := NewRelayedConnection(clientConn, ConnectionCountManager)
	rc.Relay(context.Background())
}

/*

	Connection:
	Client ------> Iris --------> Server

	Data:
	Client <-----> Iris <-------> Server

	HTTP request:
	Connection life = Request + Response

*/
