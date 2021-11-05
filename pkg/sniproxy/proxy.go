package sniproxy

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/meyskens/sniproxy/pkg/endpoints"
)

// SNIProxy is an SNI aware non-decrypting SNI proxy module
type SNIProxy struct {
	endpointsDB *endpoints.EndpointDB
}

// NewSNIProxy gives an new SNIProxy instance
func NewSNIProxy(endpointsDB *endpoints.EndpointDB) *SNIProxy {
	return &SNIProxy{
		endpointsDB: endpointsDB,
	}
}

func (s *SNIProxy) HandleConnection(conn net.Conn) error {
	defer conn.Close()

	if err := conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
		return fmt.Errorf("error setting read timeout: %w", err)
	}

	clientHello, peekedBytes, err := s.peekClientHello(conn)
	if err != nil {
		return fmt.Errorf("error reading SNI: %w", err)
	}

	clientReader := io.MultiReader(peekedBytes, conn)

	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		return fmt.Errorf("error removing timeout: %w", err)
	}

	ep, err := s.endpointsDB.Get(clientHello.ServerName)
	if err != nil {
		return fmt.Errorf("error checking endpoint db: %w", err)
	}

	backendConn, err := net.Dial("tcp", fmt.Sprintf("%s:443", ep))
	if err != nil {
		return fmt.Errorf("error dialing backend: %w", err)
	}
	defer backendConn.Close()

	// we make a wait group to wait for the 2-way copy to finish
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		io.Copy(conn, backendConn)
		conn.(*net.TCPConn).Close()
		wg.Done()
	}()
	go func() {
		io.Copy(backendConn, clientReader)
		backendConn.Close()
		wg.Done()
	}()

	wg.Wait()

	return nil
}

func (s *SNIProxy) peekClientHello(reader io.Reader) (*tls.ClientHelloInfo, *bytes.Buffer, error) {
	peekedBytes := new(bytes.Buffer)

	var hello *tls.ClientHelloInfo

	err := tls.Server(writeMockingConn{reader: io.TeeReader(reader, peekedBytes)}, &tls.Config{
		GetConfigForClient: func(argHello *tls.ClientHelloInfo) (*tls.Config, error) {
			hello = new(tls.ClientHelloInfo)
			*hello = *argHello
			return nil, nil
		},
	}).Handshake()

	// the error here is expected as we will not complete the handshake we just need the hello
	if hello == nil {
		return nil, nil, err
	}

	return hello, peekedBytes, nil
}
