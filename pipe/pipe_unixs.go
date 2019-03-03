// +build !windows

package pipe

import (
	"net"
)

// Listen creates a listener on a Windows named pipe path
// on windows e.g. \\.\pipe\mypipe.
// on unix /tmp/mypipe
// The pipe must not already exist.
func Listen(name string) (net.Listener, error) {
	return net.Listen("unix", "/tmp/"+name)
}

//Dial net.Dial by wrap
func Dial(name string) (net.Conn, error) {
	return net.Dial("unix", "/tmp/"+name)
}
