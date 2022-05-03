package event_loop

import (
	"errors"
	"syscall"
)

type (
	EventLoop struct {
		kQueeFD  int
		socketFD int
	}
)

func NewEventLoop(socketFD int) (*EventLoop, error) {
	kQueeFD, err := syscall.Kqueue()
	if err != nil {
		return nil, err
	}

	loop := &EventLoop{
		kQueeFD:  kQueeFD,
		socketFD: socketFD,
	}

	socketEvent := syscall.Kevent_t{
		Ident:  uint64(socketFD),
		Filter: syscall.EVFILT_READ,
		Flags:  syscall.EV_ADD | syscall.EV_ENABLE,
	}

	r, err := syscall.Kevent(loop.kQueeFD, []syscall.Kevent_t{socketEvent}, nil, nil)

	if err != nil {
		return nil, err
	}

	if r == -1 {
		return nil, errors.New("failed to register socket with kqueue")
	}

	return loop, nil

}

func (e *EventLoop) Start() {
	for {
		// create an empty slice, this receives all the events that are ready to be processed
		events := make([]syscall.Kevent_t, 1)
		// Here we are using syscall.Kevent to poll for events
		// it populates the events slice and returns the number of events populated
		numEvents, err := syscall.Kevent(e.kQueeFD, nil, events, nil)

		if err != nil {
			continue
		}

		// loop over the events and process them
		for i := 0; i < numEvents; i++ {
			event := events[i]
			eventFd := int(event.Ident)

			// we reached eof so close connection
			if event.Flags&syscall.EV_EOF != 0 {
				syscall.Close(eventFd)
			} else if eventFd == e.socketFD {
				// we received an event on the socket,
				// which means a new connection arrived, so we accept the new connection and add it to the kqueue
				socketFD, _, err := syscall.Accept(eventFd)
				if err != nil {
					continue
				}

				// create an event to register with the kqueue
				sockEvent := syscall.Kevent_t{
					Ident:  uint64(socketFD),
					Filter: syscall.EVFILT_READ,
					Flags:  syscall.EV_ADD | syscall.EV_ENABLE,
				}

				r, err := syscall.Kevent(e.kQueeFD, []syscall.Kevent_t{sockEvent}, nil, nil)

				if err != nil || r == -1 {
					continue
				}
			} else if event.Filter&syscall.EVFILT_READ != 0 {
				buf := make([]byte, 1024)

				n, err := syscall.Read(eventFd, buf)
				if err != nil {
					continue
				}
				syscall.Write(eventFd, buf[:n])
			}
		}
	}
}
