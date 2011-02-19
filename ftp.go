// Copyright 2011 by Ashani Ray. All Rights Reserved. 
// This material may be freely copied and distributed subject to inclusion 
// of this copyright notice

// Author: Ashani Ray
// Version: Initial
// Description: Go FTP Client


// TODO: Remove the multiple returns from function and have a single return
// whereever possible

package ftp

import ("fmt"
	"net"
	"os"
	"bufio"
	"strings"
	"regexp"
	"strconv"
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

func getIpPort(resp string) (ip string, port uint, err os.Error) {
	portRegex:= "([0-9]+,[0-9]+,[0-9]+,[0-9]+),([0-9]+,[0-9]+)"
	re, err := regexp.Compile(portRegex)
	if err != nil {
		return "", 0, err
	}
	match := re.FindStringSubmatch(resp)
	if len(match) != 3 {
		msg := "Cannot handle server response: " + resp
		return "", 0, os.NewError(msg)
	}

	ip = strings.Replace(match[1], ",", ".", -1)

	octets := strings.Split(match[2], ",", 2)
	firstOctet, _ := strconv.Atoui(octets[0])
	secondOctet, _ := strconv.Atoui(octets[1])
	port = firstOctet*256 + secondOctet

	return ip, port, nil
}

func storeDataToFile (ip string, port uint, fileName string, c chan string) {
	address := ip + ":" + strconv.Uitoa(port);

	conn, error := net.Dial("tcp", "", address);
	defer conn.Close();
	if error == nil {
		c <- "s"
	} else {
		c <- "e"
		return
	}
	<-c
	f, err := os.Open(fileName, os.O_RDONLY, 0666)
	defer f.Close()
	if err != nil {
		msg := fmt.Sprintf("Cannot open file '%s': %v", fileName, err)
		c <-msg
		return
	}

	// Buffer for reading from file
	bufLen := 1024
	buf := make([]byte, bufLen)

	// Read from the server and write the contents to a file
	for {
		count, err := f.Read(buf)
		if err == os.EOF {
			break
		}
		if err != nil {
			msg := fmt.Sprintf("%v", err)
			c <-msg
			break
		}
		_, err = conn.Write(buf[0:count])
		if err != nil {
			msg := fmt.Sprintf("Coudn't write to file, '%s'. Error: %v", fileName, err)
			c <-msg
			return
		}
	}
	c <-"C"
}

func recvDataToFile (ip string, port uint, fileName string, c chan string) {
	address := ip + ":" + strconv.Uitoa(port);
	conn, error := net.Dial("tcp", "", address);
	defer conn.Close();
	if error == nil {
		c <- "s"
	} else {
		c <- "e"
		return
	}

	// Read from socket and redirect to file
	f, err := os.Open(fileName, os.O_CREAT|os.O_WRONLY, 0666)
	defer f.Close()
	if err != nil {
		msg := fmt.Sprintf("Cannot open file '%s': %v", fileName, err)
		c <-msg
		return
	}

	// Buffer for downloading and writing to file
	bufLen := 1024
	buf := make([]byte, bufLen)

	// Read from the server and write the contents to a file
	for {
		bytesRead, err := conn.Read(buf)
		if bytesRead > 0 {
			_, err := f.Write(buf[0:bytesRead])
			if err != nil {
				msg := fmt.Sprintf("Coudn't write to file, '%s'. Error: %v", fileName, err)
				c <-msg
				return
			}
		}
		if err == os.EOF {
			break
		}
		if err != nil {
			msg := fmt.Sprintf("%v", err)
			c <-msg
		}
	}
	c <-"C"
}

func getPasvIpPort (conn *net.Conn) (ip string, port uint, resp string, err os.Error) {
	err = SendCtrlCmd(conn, "PASV")
	if err != nil {
		return
	}
	var code int
	err, code, resp = RecvCtrlResp(conn)
	if err != nil {
		return
	}
	if code != 227 {
		err = os.NewError("Code Returned from server for PASV is not 227")
		return
	}

	ip, port, err = getIpPort(resp)
	return
}

func ExecGet (conn *net.Conn, file string) (bool, os.Error, string) {
	err := SendCtrlCmd(conn, "PASV")
	if err != nil {
		return true, err, ""
	}
	err, code, resp := RecvCtrlResp(conn)
	if err != nil {
		return true, err, resp
	}
	if code != 227 {
		return true, err, resp
	}

	ip, port, err := getIpPort(resp)
	if err != nil {
		return true, err, resp
	}

	ch := make (chan string)
	go recvDataToFile(ip, port, file, ch)
	start := <-ch
	if start == "e" {
		msg := "Unable to connected to server is PASV port"
		return true, os.NewError(msg), ""
	}
	err = SendCtrlCmd(conn, "RETR " + file)
	err, _, resp = RecvCtrlResp(conn)
	if err != nil {
		return true, err, resp
	}
	recvMsg := <-ch
	if recvMsg != "C" {
		err = os.NewError(recvMsg)
	} else {
		var respT string
		err, _, respT = RecvCtrlResp(conn)
		resp += respT
	}

	return true, err, resp
}

func ExecDefault (conn *net.Conn, cmd string) (bool, os.Error, string) {
	resp := "Invalid Command. Valid Commands are:" + NEWLINE
	for k, _ := range cmdTable {
		resp = resp + k + " "
	}
	resp += NEWLINE
	return true, nil, resp
}

func ExecBinary (conn *net.Conn, cmd string) (bool, os.Error, string) {
	var resp string
	err := SendCtrlCmd(conn, "TYPE I" + cmd)
	if err == nil {
		err, _, resp = RecvCtrlResp(conn)
	}
	return true, err, resp
}

func ExecAscii (conn *net.Conn, cmd string) (bool, os.Error, string) {
	var resp string
	err := SendCtrlCmd(conn, "TYPE A" + cmd)
	if err == nil {
		err, _, resp = RecvCtrlResp(conn)
	}
	return true, err, resp
}

func ExecPut (conn *net.Conn, file string) (bool, os.Error, string) {
	err := SendCtrlCmd(conn, "PASV")
	if err != nil {
		return true, err, ""
	}
	err, code, resp := RecvCtrlResp(conn)
	if err != nil {
		return true, err, resp
	}
	if code != 227 {
		return true, err, resp
	}

	ip, port, err := getIpPort(resp)
	if err != nil {
		return true, err, resp
	}

	ch := make (chan string)
	go storeDataToFile(ip, port, file, ch)
	start := <-ch
	if start == "e" {
		msg := "Unable to connected to server is PASV port"
		return true, os.NewError(msg), ""
	}
	// Our socket connected to remote
	err = SendCtrlCmd(conn, "STOR " + file)
	err, code, resp = RecvCtrlResp(conn)
	if err != nil || code != 150 {
		return true, err, resp
	}
	ch <- "S"
	recvMsg := <-ch
	if recvMsg != "C" {
		err = os.NewError(recvMsg)
	} else {
		var respT string
		err, _, respT = RecvCtrlResp(conn)
		resp += respT
	}

	return true, err, resp
}


func ExecList (conn *net.Conn, file string) (bool, os.Error, string) {
	ip, port, resp, err := getPasvIpPort(conn)
	if err != nil {
		return true, err, resp
	}

	ch := make (chan string)
	msgCh := make (chan string)

	go getDirList(ip, port, ch, msgCh)
	status := <-ch
	if status == "e" {
		msg := "Unable to connected to server is PASV port"
		return true, os.NewError(msg), ""
	}

	err = SendCtrlCmd(conn, "LIST")
	if err != nil {
		return true, err, ""
	}
	var respThis string
	err, _, respThis = RecvCtrlResp(conn)
	resp += respThis
	if err != nil {
		return true, err, resp
	}

	status =<-ch
	for status == "c" {
		recvd := <-msgCh
		resp += recvd
		status =<-ch
	}

	err, _, respThis = RecvCtrlResp(conn)
	resp += respThis

	if status != "C" {
		return true, os.NewError(status), resp
	}

	return true, err, resp
}



func getDirList (ip string, port uint, c chan string, msg chan string) {
	address := ip + ":" + strconv.Uitoa(port);
	conn, error := net.Dial("tcp", "", address);
	defer conn.Close();
	if error == nil {
		c <- "s"
	} else {
		c <- "e"
		return
	}

	// Read from socket and redirect to file
	bufLen := 1024
	buf := make([]byte, bufLen)

	// Read from the server and write the contents to channel
	for {
		bytesRead, err := conn.Read(buf)
		if bytesRead > 0 {
			c <-"c"
			msg <- string(buf[0:bytesRead])
			if err != nil {
				msg := fmt.Sprintf("Error: %v", err)
				c <-msg
				return
			}
		}
		if err == os.EOF {
			break
		}
		if err != nil {
			msg := fmt.Sprintf("%v", err)
			c <-msg
		}
	}
	c <-"C"
	return
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
	"GET"  : ExecGet,
	"ASCII": ExecAscii,
	"BIN"  : ExecBinary,
	"PUT"  : ExecPut,
	"LIST" : ExecList,
	// Add more commands here
}

func ExecCmd(conn *net.Conn, line string) (bool, string) {
	var resp string
	var cont bool = true

	cmd := strings.Trim(line, " \t\r\n")

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

