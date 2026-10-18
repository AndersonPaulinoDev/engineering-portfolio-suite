package realtime

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"io"
	"net"
	"net/http"
)

const websocketGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

func ServeWebSocket(w http.ResponseWriter, r *http.Request, hub *Hub) error {
	key := r.Header.Get("Sec-WebSocket-Key")
	if key == "" || r.Header.Get("Upgrade") != "websocket" {
		return errors.New("invalid websocket upgrade")
	}
	hj, ok := w.(http.Hijacker)
	if !ok {
		return errors.New("hijacking unsupported")
	}
	conn, rw, err := hj.Hijack()
	if err != nil {
		return err
	}
	sum := sha1.Sum([]byte(key + websocketGUID))
	accept := base64.StdEncoding.EncodeToString(sum[:])
	_, err = rw.WriteString("HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: " + accept + "\r\n\r\n")
	if err != nil {
		conn.Close()
		return err
	}
	if err = rw.Flush(); err != nil {
		conn.Close()
		return err
	}

	ch, unsub := hub.Subscribe()
	defer unsub()
	defer conn.Close()
	done := make(chan struct{})
	go func() { defer close(done); drainFrames(conn) }()
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return nil
			}
			if err := writeTextFrame(conn, msg); err != nil {
				return err
			}
		case <-done:
			return nil
		case <-r.Context().Done():
			return nil
		}
	}
}

func writeTextFrame(w io.Writer, payload []byte) error {
	header := []byte{0x81}
	n := len(payload)
	switch {
	case n < 126:
		header = append(header, byte(n))
	case n <= 65535:
		header = append(header, 126, byte(n>>8), byte(n))
	default:
		header = append(header, 127, 0, 0, 0, 0, byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
	}
	if _, err := w.Write(header); err != nil {
		return err
	}
	_, err := w.Write(payload)
	return err
}

func drainFrames(conn net.Conn) {
	r := bufio.NewReader(conn)
	for {
		b1, err := r.ReadByte()
		if err != nil {
			return
		}
		b2, err := r.ReadByte()
		if err != nil {
			return
		}
		opcode := b1 & 0x0f
		masked := b2&0x80 != 0
		n := int(b2 & 0x7f)
		if n == 126 {
			a, _ := r.ReadByte()
			b, _ := r.ReadByte()
			n = int(a)<<8 | int(b)
		} else if n == 127 {
			buf := make([]byte, 8)
			if _, err := io.ReadFull(r, buf); err != nil {
				return
			}
			n = 0
			for _, v := range buf {
				n = n<<8 | int(v)
			}
		}
		var mask [4]byte
		if masked {
			if _, err := io.ReadFull(r, mask[:]); err != nil {
				return
			}
		}
		payload := make([]byte, n)
		if _, err := io.ReadFull(r, payload); err != nil {
			return
		}
		if opcode == 0x8 {
			return
		}
		if opcode == 0x9 {
			_ = writeControlFrame(conn, 0xA, payload)
		}
	}
}

func writeControlFrame(w io.Writer, opcode byte, payload []byte) error {
	if len(payload) > 125 {
		payload = payload[:125]
	}
	if _, err := w.Write([]byte{0x80 | opcode, byte(len(payload))}); err != nil {
		return err
	}
	_, err := w.Write(payload)
	return err
}
