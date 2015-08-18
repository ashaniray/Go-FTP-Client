package main

import (
	"flag"
	"net"
	"os"
	"fmt"
	"strconv"
	"bufio"
	)


var svr = flag.String("s", "127.0.0.1", "host")
var port = flag.Int("x", 21, "port")
var login = flag.String("u", "user", "your login id")
var passwd = flag.String("p", "guest", "your password")

var PROMPT string = "> "

func main() {

	flag.Parse()
	address := *svr + ":" + strconv.Itoa(*port);

	fmt.Println("Connecting to " + address);
	conn, error := net.Dial("tcp", address);
	if error != nil { fmt.Printf("Error: %s\n", error ); os.Exit(1); }
	defer conn.Close();

	_, _, response := RecvCtrlResp(&conn);

	_, error, response = ExecUser(&conn, *login)
	if error != nil {
		fmt.Printf("Error : %s while sending USER %s\n", error, *login );
		os.Exit(2);
	}
	//fmt.Print(response);

	_, error, response = ExecPass(&conn, *passwd)
	if error != nil {
		fmt.Printf("Error : %s while sending passwd\n", error);
		os.Exit(2);
	}
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
		cont, resp = ExecCmd(&conn, cmd);
		fmt.Print(resp)
	}
	conn.Close();
}

