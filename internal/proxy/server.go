package proxy

import (
	"errors"
	"io"
	"net"

	log "github.com/sirupsen/logrus"
)

const (
	BufferSize = 1024
)

//func Serve(from *net.TCPConn, to *net.TCPConn, proto string, fd string, td string) {
//	defer func() {
//		from.Close()
//		to.Close()
//
//		log.Debugf("%s closing proxy connection: %s -> %s", proto, fd, td)
//	}()
//
//	buf := make([]byte, BufferSize)
//	for {
//		bytesRead, err := ReadBytes(from, buf)
//		if err != nil {
//			if err == io.EOF {
//				log.Debugf("%s finished reading from %s", proto, fd)
//				return
//			}
//			log.Debugf("%s error reading from %s: %s", proto, fd, err)
//			return
//		}
//
//		if _, err := to.Write(bytesRead); err != nil {
//			log.Debugf("%s error Writing to %s", proto, td)
//			return
//		}
//	}
//}

func Serve(from *net.TCPConn, to *net.TCPConn, fd string, td string) {
	defer func() {
		from.Close()
		to.Close()

		log.Debugf("closing proxy connection: %s -> %s", fd, td)
	}()

	buf := make([]byte, BufferSize)
	for {
		bytesRead, err := readBytes(from, buf)
		if err != nil {
			if err == io.EOF {
				log.Debugf("finished reading from %s", fd)
				return
			}
			log.Debugf("error reading from %s: %s", fd, err)
			return
		}

		if _, err := to.Write(bytesRead); err != nil {
			log.Debugf("error Writing to %s", td)
			return
		}
	}
}

func readBytes(conn *net.TCPConn, dest []byte) ([]byte, error) {
	totalRead, err := conn.Read(dest)
	if err != nil {
		var opError *net.OpError
		switch {
		case errors.As(err, &opError) && opError.Timeout():
			return dest[:totalRead], errors.New("timed out")
		default:
			return dest[:totalRead], err
		}
	}

	return dest[:totalRead], err
}
