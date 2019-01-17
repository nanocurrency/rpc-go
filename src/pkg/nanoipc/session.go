package nanoipc

import (
	"context"
	"encoding/binary"
	"io"
	"log"
	"net"
	"net/url"
	"sync"
	"time"
)

// A Session represents a persistent connection to the Nano node.
type Session struct {
	mutex      sync.Mutex
	connection net.Conn
	// True if the session has been connected to the node
	Connected bool
	// Read and Write timeout. Default is 30 seconds.
	TimeoutReadWrite int
	// Connection timeout. Default is 10 seconds.
	TimeoutConnection int
}

// Connect to a node. You can set Session#TimeoutConnection before this call, otherwise a default
// of 15 seconds is used.
// connectionString is an URI of the form tcp://host:port or local:///path/to/domainsocketfile
func (s *Session) Connect(connectionString string) *Error {
	var connError *Error
	uri, err := url.Parse(connectionString)
	if err != nil {
		connError = &Error{1, "Invalid connection string", "Connection"}
	} else {
		scheme := uri.Scheme
		host := uri.Host
		if scheme == "local" {
			scheme = "unix"
			host = uri.Path
		} else if scheme != "tcp" {
			connError = &Error{1, "Invalid schema: Use tcp or local.", "Connection"}
		}

		if s.TimeoutConnection == 0 {
			s.TimeoutConnection = 15
		}
		if s.TimeoutReadWrite == 0 {
			s.TimeoutReadWrite = 30
		}
		dialContext := (&net.Dialer{
			KeepAlive: 30 * time.Second,
			Timeout:   time.Duration(s.TimeoutConnection) * time.Second,
		}).DialContext

		con, err := dialContext(context.Background(), scheme, host)
		if err != nil {
			connError = &Error{1, err.Error(), "Connection"}
			s.Connected = false
			log.Println(err.Error())
		} else {
			s.connection = con
			s.Connected = true
		}
	}

	return connError
}

// Close the underlying connection to the node
func (s *Session) Close() *Error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var err *Error
	if s.Connected {
		s.Connected = false
		closeErr := s.connection.Close()
		if closeErr != nil {
			err = &Error{1, closeErr.Error(), "Connection"}
		}
	}
	return err
}

// Updates the write deadline using Session#TimeoutReadWrite
func (s *Session) updateWriteDeadline() {
	s.connection.SetWriteDeadline(time.Now().Add(time.Duration(s.TimeoutReadWrite) * time.Second))
}

// Updates the read deadline using Session#TimeoutReadWrite
func (s *Session) updateReadDeadline() {
	s.connection.SetReadDeadline(time.Now().Add(time.Duration(s.TimeoutReadWrite) * time.Second))
}

// Request send JSON request to the node via IPC. The session must be connected.
// This method is threadsafe. Use a larger pool size to increase concurrency
// when multiple threads are using the same Session.
// Returns the result as a byte array, or an error.
func (s *Session) Request(request string) ([]byte, *Error) {

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// RPC legacy IPC type
	const PROTOCOL_PREAMBLE_LEAD = 'N'
	const PROTOCOL_ENCODING = 1
	const PROTOCOL_RESERVED = 0

	var bufResponse []byte
	var errReply *Error
	if !s.Connected {
		errReply = &Error{1, "Not connected", "Network"}
	} else {
		sc := &CallChain{}

		var preamble [4]byte
		var bufLen [4]byte
		var err error
		sc.Do(func() {
			preamble = [4]byte{
				PROTOCOL_PREAMBLE_LEAD,
				PROTOCOL_ENCODING,
				PROTOCOL_RESERVED,
				PROTOCOL_RESERVED}
			s.updateWriteDeadline()
			if _, err = s.connection.Write(preamble[:]); err != nil {
				sc.Err = &Error{1, err.Error(), "Network"}
			}
		}).Do(func() {
			binary.BigEndian.PutUint32(bufLen[:], uint32(len(request)))
			s.updateWriteDeadline()
			if _, err = s.connection.Write(bufLen[:]); err != nil {
				sc.Err = &Error{1, err.Error(), "Network"}
			}
		}).Do(func() {
			s.updateWriteDeadline()
			if _, err = s.connection.Write([]byte(request)); err != nil {
				sc.Err = &Error{1, err.Error(), "Network"}
			}
		}).Do(func() {
			// Response is big endian size followed by json response payload
			s.updateReadDeadline()
			if _, err = io.ReadFull(s.connection, bufLen[:]); err != nil {
				sc.Err = &Error{1, err.Error(), "Network"}
			}
		}).Do(func() {
			bufResponse = make([]byte, binary.BigEndian.Uint32(bufLen[:]))
			s.updateReadDeadline()
			if _, err = io.ReadFull(s.connection, bufResponse); err != nil {
				sc.Err = &Error{1, err.Error(), "Network"}
			}
		}).Failure(func() {
			errReply = sc.Err
			log.Println(err.Error())
		})
	}
	return bufResponse, errReply
}
