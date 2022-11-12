package tcpc

import (
	"encoding/gob"
	"log"
	"net"
	"time"
)

type TCPC[T any] struct {
	listenAddr string
	remoteAddr string

	Sendchan     chan T
	Recvchan     chan T
	outboundConn net.Conn
	ln           net.Listener
}

func New[T any](listenAddr, remoteAddr string) (*TCPC[T], error) {
	tcpc := &TCPC[T]{
		listenAddr: listenAddr,
		remoteAddr: remoteAddr,
		Sendchan:   make(chan T, 10),
		Recvchan:   make(chan T, 10),
	}

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, err
	}
	tcpc.ln = ln

	go tcpc.loop()
	go tcpc.acceptLoop()
	go tcpc.dialRemoteAndRead()

	return tcpc, nil
}

func (t *TCPC[T]) loop() {
	for {
		msg := <-t.Sendchan
		log.Println("sending msg over the wire: ", msg)
		if err := gob.NewEncoder(t.outboundConn).Encode(&msg); err != nil {
			log.Println(err)
		}
	}
}

func (t *TCPC[T]) acceptLoop() {
	defer func() {
		t.ln.Close()
	}()

	for {
		conn, err := t.ln.Accept()
		if err != nil {
			log.Println("accept error:", err)
			return
		}

		log.Printf("sender connected %s", conn.RemoteAddr())

		go t.handleConn(conn)
	}
}

func (t *TCPC[T]) handleConn(conn net.Conn) {
	for {
		var msg T
		if err := gob.NewDecoder(conn).Decode(&msg); err != nil {
			log.Println(err)
			continue
		}
		t.Recvchan <- msg
	}
}

func (t *TCPC[T]) dialRemoteAndRead() {
	conn, err := net.Dial("tcp", t.remoteAddr)
	if err != nil {
		log.Printf("dial error (%s)", err)
		time.Sleep(time.Second * 3)
		t.dialRemoteAndRead()
	}

	t.outboundConn = conn
}
