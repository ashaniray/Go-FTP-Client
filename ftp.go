// Copyright 2011 by Ashani Ray. All Rights Reserved. 
// This material may be freely copied and distributed subject to inclusion 
// of this copyright notice

// Author: Ashani Ray
// Version: Initial
// Description: Go FTP Client


// TODO: Code is incomplete......

package main

import ("fmt"
	"flag"
	"net"
	"os"
	)

var address = flag.String("s", "127.0.0.1:21", "host:port")
var login = flag.String("l", "", "loginid")
var passwd = flag.String("pass", "", "password")

func main() {
	flag.Parse()
	var (
		msg = *login + "\n" + *passwd
		data = make([]byte, 1024)
	)

	fmt.Println("Connecting to " + *address);
	conn, error := net.Dial("tcp", "", *address);

	if error != nil { fmt.Printf("Error: %s\n", error ); os.Exit(1); }
	defer conn.Close();

	fmt.Println("Writing " + msg);
	in, error := conn.Write([]byte(msg));
	if error != nil { fmt.Printf("Error : %s, in: %d\n", error, in ); os.Exit(2); }

	fmt.Println("Reading..." );
	var response string
	n, error := conn.Read(data);
	switch error {
	case os.EOF:
		fmt.Printf("EOF: %s \n", error);
	case nil:
		response = response + string(data[0:n]);
	default:
		fmt.Printf("Error: %s \n", error);
	}
	fmt.Println("Data Recvd: " + response);
	conn.Close();
}

