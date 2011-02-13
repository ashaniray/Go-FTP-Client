// Copyright 2011 by Ashani Ray. All Rights Reserved. 
// This material may be freely copied and distributed subject to inclusion 
// of this copyright notice

// Author: Ashani Ray
// Version: Initial
// Description: Go FTP Client


// TODO: Code is incomplete......

package ftp

import ("fmt"
	"net"
	"os"
	"bufio"
	"strings"
	)

var NEWLINE string = "\r\n"

func RecvCtrlResp(conn *net.Conn) (os.Error, int, string) {
	var code int = -1
	r := bufio.NewReader(bufio.NewReader(*conn))
	resp, err := r.ReadString('\n')
	if err == nil {
		fmt.Sscanf(resp, "%d", &code);
	}
	return err, code, resp;
}

func SendCtrlCmd(conn *net.Conn, req string) (os.Error) {
	w := bufio.NewWriter(bufio.NewWriter(*conn))
	w.WriteString(req + NEWLINE)
	error := w.Flush()
	return error
}

func ExecQuit (conn *net.Conn, req string ) (bool, os.Error, string) {
	var resp string
	err := SendCtrlCmd(conn, "QUIT");
	if err == nil {
		err, _, resp = RecvCtrlResp(conn)
	}
	return false, err, resp
}

func ExecPass (conn *net.Conn, cmd string) (bool, os.Error, string) {
	var resp string
	err := SendCtrlCmd(conn, "PASS " + cmd)
	if err == nil {
		err, _, resp = RecvCtrlResp(conn)
	}
	return true, err, resp
}

func ExecUser (conn *net.Conn, cmd string) (bool, os.Error, string) {
	var resp string
	err := SendCtrlCmd(conn, "USER " + cmd)
	if err == nil {
		err, _, resp = RecvCtrlResp(conn)
	}
	return true, err, resp
}

func ExecDefault (conn *net.Conn, cmd string) (bool, os.Error, string) {
	return true, nil, "Invalid Command" + NEWLINE
}

// The main table
// Key is the command line command
// Value is the function (command pattern) to Execute against the command
// Arguments of the function:
// conn -> Control Connection to ftp server
// cmd -> the command line args provided
// Return values:
// bool -> true, unless the connection is snapped by QUIT
// os.Error -> the error
// string -> the string to be returned and displayed to user
var cmdTable = map [string] func(*net.Conn, string) (bool, os.Error, string) {
	"QUIT" : ExecQuit,
	"PASS" : ExecPass,
	"USER" : ExecUser,
	// Add more commands here
}

func ExecCmd(conn *net.Conn, cmd string) (bool, string) {
	var resp string
	var cont bool = true

	tokens := strings.SplitAfter(cmd, " ", 2)
	key := strings.Trim(strings.ToUpper(tokens[0]), " \t\r\n")
	args := ""
	if (len(tokens) > 1) {
		args = tokens[1]
	}
	if f, ok := cmdTable[key]; ok {
		cont, _, resp = f(conn, args)
	} else {
		cont, _, resp = ExecDefault(conn, cmd)
	}
	return cont, resp
}

