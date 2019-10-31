package main

import (
	"fmt"
	"log"
	"net"
)

func Vf(level int, format string, v ...interface{}) {
	if level <= *verbosity {
		log.Printf(format, v...)
	}
}
func V(level int, v ...interface{}) {
	if level <= *verbosity {
		log.Print(v...)
	}
}
func Vln(level int, v ...interface{}) {
	if level <= *verbosity {
		log.Println(v...)
	}
}

func LogIncomeMessage(f Frame) {
	incomingType := fmt.Sprintf("[%s]", f.cmd)
	Vln(4, fmt.Sprintf("[%s]", incomingType), f)
}

func Respond(p net.Conn, typ string, bytes []byte) {
	Vln(4, typ, bytes)
	_, err := p.Write(bytes)
	if err != nil {
		fmt.Printf("Error writing %v to socket because %v", bytes, err)
	}
}
