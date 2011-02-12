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
	"bufio"
	"strconv"
	)

var address = flag.String("s", "127.0.0.1:21", "host:port")
var login = flag.String("u", "user", "your login id")
var passwd = flag.String("p", "guest", "your password")

var NEWLINE string = "\r\n"
var PROMPT string = "> "

func recvCtrlResp(conn net.Conn) (os.Error, int, string) {
	var code int = -1
	r := bufio.NewReader(bufio.NewReader(conn))
	resp, err := r.ReadString('\n')
	if err == nil {
		fmt.Sscanf(resp, "%d", &code);
	}
	return err, code, resp;
}

func sendCtrlCmd(conn net.Conn, req string) (os.Error) {
	w := bufio.NewWriter(bufio.NewWriter(conn))
	w.WriteString(req + NEWLINE)
	error := w.Flush()
	return error
}


func main() {
	flag.Parse()

	fmt.Println("Connecting to " + *address);
	conn, error := net.Dial("tcp", "", *address);
	if error != nil { fmt.Printf("Error: %s\n", error ); os.Exit(1); }
	defer conn.Close();

	_, code, response := recvCtrlResp(conn);
	fmt.Println(response + "[CODE=" + strconv.Itoa(code) +"]");

	var userMsg string = "USER " + *login;
	var passMsg string = "PASS " + *passwd;

	fmt.Println(PROMPT + userMsg);
	error = sendCtrlCmd(conn, userMsg)
	if error != nil { fmt.Printf("Error : %s while sending %s\n", error, userMsg ); os.Exit(2); }

	_, code, response = recvCtrlResp(conn);
	fmt.Print(response);
	fmt.Println("[CODE=" + strconv.Itoa(code) +"]");

	fmt.Println(PROMPT + passMsg);
	error = sendCtrlCmd(conn, passMsg)
	if error != nil { fmt.Printf("Error : %s while sending %d\n", error, passMsg ); os.Exit(2); }

	_, code, response = recvCtrlResp(conn);
	fmt.Print(response);
	fmt.Println("[CODE=" + strconv.Itoa(code) +"]");

	fmt.Println(PROMPT + "QUIT");
	error = sendCtrlCmd(conn, "QUIT")

	conn.Close();
}

