package server

import (
	"context"
	"crypto/rand"
	"log/slog"
	"math/big"
	"net"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
)

type (
	// Server represent server main struct.
	Server struct {
		repo repo

		conns sync.Map
	}

	repo interface {
		// CheckOrStore return true if represents, false if saved.
		CheckOrStore(i *big.Int) bool
	}

	response struct {
		RandomNumber *big.Int `json:"random_number"`
	}
)

// New return new server.
func New(
	repo repo,
) *Server {
	return &Server{
		repo: repo,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  0,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Start server
func (s *Server) Start(ctx context.Context, port uint64) {
	slog.Info(
		"service started",
		"port", port,
	)

	http.HandleFunc("/ws", s.serveWS)

	go func() {
		if err := http.ListenAndServe(":"+strconv.FormatUint(port, 10), nil); err != nil {
			slog.Error(
				"http serve",
				"error", err,
			)
		}
	}()

	<-ctx.Done()

	slog.Info("service stopped")
}

// serveWS is a HTTP Handler websocket connection.
func (s *Server) serveWS(w http.ResponseWriter, r *http.Request) {
	// Begin by upgrading the HTTP request.
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error(
			"upgrade connection",
			"error", err,
		)

		return
	}

	// get host from address.
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		slog.Error(
			"parse host",
			"error", err,
		)

		return
	}

	slog.Info(
		"new connection",
		"host", host,
	)

	// close  old connection if exists
	oldConn, ok := s.conns.Swap(host, conn)
	if ok {
		connToClose, ok := oldConn.(*websocket.Conn)
		if ok && connToClose != nil {
			if err := connToClose.Close(); err != nil {
				slog.Error(
					"connection close",
					"error", err,
				)
			}

			slog.Info(
				"connection renewed",
				"host", host,
			)
		}
	}

	// handle requests.
	go func() {
		defer func() {
			if ok := s.conns.CompareAndDelete(host, conn); ok {
				slog.Info(
					"connection deleted",
					"host", host,
				)
			}

			// Graceful Close the Connection once this function is done.
			conn.Close()
		}()

		for {
			// ReadMessage is used to read the next message in queue in the connection.
			_, _, err := conn.ReadMessage()
			if err != nil {
				// If Connection is closed, we will Recieve an error here.
				// We only want to log Strange errors, but not simple Disconnection.
				if websocket.IsUnexpectedCloseError(
					err,
					websocket.CloseNormalClosure,
					websocket.CloseGoingAway,
					websocket.CloseAbnormalClosure,
				) {
					slog.Warn(
						"reading message",
						"error", err,
					)
				}

				break // Break the loop to close conn & Cleanup
			}

			conn.WriteJSON(response{RandomNumber: s.getRndNumber()})
		}
	}()
}

func (s *Server) getRndNumber() *big.Int {
	var (
		i   *big.Int
		err error
	)

	for {
		i, err = rand.Int(rand.Reader, big.NewInt(0).MulRange(1, 1000))
		if err != nil {
			slog.Error(
				"get rnd number",
				"error", err,
			)
		}

		if !s.repo.CheckOrStore(i) {
			break
		}
	}

	return i
}
