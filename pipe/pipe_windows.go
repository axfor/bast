// +build windows

package pipe

import (
	"net"

	"github.com/microsoft/go-winio"
)

// Pipe creates a listener on a Windows named pipe path
// on windows e.g. \\.\pipe\mypipe.
// on unix /tmp/mypipe
// The pipe must not already exist.
func Listen(name string) (net.Listener, error) {
	return winio.ListenPipe(`\\.\pipe\`+name, nil)
}

//Dial winio.DialPipe by wrap
func Dial(name string) (net.Conn, error) {
	return winio.DialPipe(`\\.\pipe\`+name, nil)
}
