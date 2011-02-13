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
	"strings"
	)

var svr = flag.String("s", "127.0.0.1", "host")
var port = flag.Int("x", 21, "port")
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

func execQuit (conn net.Conn, req string ) (bool, os.Error, string) {
	var resp string
	err := sendCtrlCmd(conn, "QUIT");
	if err == nil {
		err, _, resp = recvCtrlResp(conn)
	}
	return false, err, resp
}

func execPass (conn net.Conn, cmd string) (bool, os.Error, string) {
	var resp string
	err := sendCtrlCmd(conn, cmd)
	if err == nil {
		err, _, resp = recvCtrlResp(conn)
	}
	return true, err, resp
}

func execUser (conn net.Conn, cmd string) (bool, os.Error, string) {
	var resp string
	err := sendCtrlCmd(conn, cmd)
	if err == nil {
		err, _, resp = recvCtrlResp(conn)
	}
	return true, err, resp
}

func execDefault (conn net.Conn, cmd string) (bool, os.Error, string) {
	return true, nil, "Invalid Command" + NEWLINE
}

// The main table
// Key is the command line command
// Value is the function (command pattern) to execute against the command
// Arguments of the function:
// conn -> Control Connection to ftp server
// cmd -> the command line provided
// Return values:
// bool -> true, unless the main loop needs to quit
// os.Error -> the error
// string -> the string to be returned and displayed to user
var cmdTable = map [string] func(net.Conn, string) (bool, os.Error, string) {
	"QUIT" : execQuit,
	"PASS" : execPass,
	"USER" : execUser,
	// Add more commands here
}

func execCmd(conn net.Conn, cmd string) (bool, string) {
	var resp string
	var cont bool = true

	tokens := strings.SplitAfter(cmd, " ", 2)
	key := strings.Trim(strings.ToUpper(tokens[0]), " \t\r\n")

	if f, ok := cmdTable[key]; ok {
		cont, _, resp = f(conn, cmd)
	} else {
		cont, _, resp = execDefault(conn, cmd)
	}
	return cont, resp
}

func main() {

	flag.Parse()
	address := *svr + ":" + strconv.Itoa(*port);

	fmt.Println("Connecting to " + address);
	conn, error := net.Dial("tcp", "", address);
	if error != nil { fmt.Printf("Error: %s\n", error ); os.Exit(1); }
	defer conn.Close();

	_, _, response := recvCtrlResp(conn);

	var userMsg string = "USER " + *login;
	fmt.Println(PROMPT + userMsg);
	_, error, response = execUser(conn, userMsg)
	if error != nil { fmt.Printf("Error : %s while sending %s\n", error, userMsg ); os.Exit(2); }
	fmt.Print(response);

	var passMsg string = "PASS " + *passwd;
	fmt.Println(PROMPT + passMsg);
	_, error, response = execPass(conn, passMsg)
	if error != nil { fmt.Printf("Error : %s while sending %d\n", error, passMsg ); os.Exit(2); }
	fmt.Print(response);


	cin := bufio.NewReader(os.Stdin)
	cont := true
	var resp, cmd string
	for cont {
		fmt.Print(PROMPT)
		cmd, error = cin.ReadString('\n')
		if error != nil {
			fmt.Printf("Error : %s \n", error );
			break
		}
		cont, resp = execCmd(conn, cmd);
		fmt.Print(resp)
	}
	conn.Close();
}

