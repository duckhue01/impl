package server

import (
	"log"
	"net"
	"syscall"

	"github.com/duckhue01/impl/event-loop/event_loop"
)

type (
	Socket struct {
		fD int
	}

	Server struct {
		socket *Socket
	}
)

func NewServer(host string, port int) (*Server, error) {
	fD, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if err != nil {
		return nil, err
	}
	// bind socket to the address
	socket := &Socket{fD: fD}
	addr := syscall.SockaddrInet4{
		Port: port,
	}
	copy(addr.Addr[:], net.ParseIP(host))
	syscall.Bind(socket.fD, &addr)

	// syscall.Listen marks that the socket will be used for accepting new connections
	err = syscall.Listen(socket.fD, syscall.SOMAXCONN)

	if err != nil {
		return nil, err
	}

	return &Server{socket: socket}, nil
}

func (s *Server) Listen() {
	loop, err := event_loop.NewEventLoop(s.socket.fD)
	if err != nil {
		log.Fatal(err)
	}
	loop.Start()
}

func (s *Server) Close() error {
	return syscall.Close(s.socket.fD)
}

func (s *Socket) Fd() int {
	return s.fD
}
