package main

import (
	"fmt"
	"net"
	"os"
)

var name string

func main() {
	if s := login(); s != "0" {
		ch := make(chan int)
		flag := ""
		for {
			fmt.Println("进入公共聊天室输入1，想要私聊输入2")
			fmt.Scanln(&flag)
			if !(flag == "1" || flag == "2") {
				fmt.Println("只能输入1和2")
				continue
			}
			break
		}
		address, err := net.ResolveTCPAddr("tcp", ":1200")
		checkError(err)
		conn, err := net.DialTCP("tcp", nil, address)
		checkError(err)
		_, err = conn.Write([]byte(name + ";" + flag + ";" + s + ";"))
		checkError(err)
		go recive(conn)
		if flag == "1" {
			go many(conn, ch)
			<-ch
		} else {
			go one(conn, ch)
			<-ch
		}
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func login() string {
	address, err := net.ResolveTCPAddr("tcp", ":1201")
	checkError(err)
	conn, err := net.DialTCP("tcp", nil, address)
	checkError(err)
Again:
	recvData := make([]byte, 128)
	fmt.Println("1：登录，2：注册")
	flag := ""
	fmt.Scanln(&flag)
	if !(flag == "1" || flag == "2") {
		fmt.Println("只能输入1和2")
		goto Again
	}
	fmt.Print("输入你的姓名")
	fmt.Scanln(&name)
	_, err = conn.Write([]byte(name + ";" + flag + ";"))
	checkError(err)
	_, err = conn.Read(recvData)
	checkError(err)
	if string(recvData[0]) == "1" && flag == "1" {
		fmt.Println("成功，欢迎你 " + name)
		conn.Close()
		return "1"
	} else if string(recvData[0]) == "1" && flag == "2" {
		fmt.Println("成功，欢迎你 " + name)
		conn.Close()
		return "2"
	} else if string(recvData[0]) == "2" {
		fmt.Println("这个名称已经被注册了")
		goto Again
	} else if string(recvData[0]) == "3" {
		fmt.Println("没有这个名称，请先去注册")
		goto Again
	} else {
		fmt.Println("服务器好像有点问题,待会再试试")
		conn.Close()
		return "0"
	}
}

func recive(conn net.Conn) {
	for {
		recvData := make([]byte, 2048)
		n, err := conn.Read(recvData)
		if err != nil {
			fmt.Println(err)
			break
		}
		if n == 0 {
			break
		} else {
			recvStr := string(recvData[:n])
			fmt.Println(recvStr)
		}
	}
	os.Exit(1)
}

func one(conn net.Conn, ch chan int) {
chat:
	fmt.Println("请输入你要聊天的用户名")
	toName := ""
	fmt.Scanln(&toName)
	_, err := conn.Write([]byte(toName + ";"))
	checkError(err)
	message := ""
	for {
		fmt.Scanln(&message)
		if message == "/quit" {
			conn.Write([]byte(message))
			goto chat
		} else if message == "/change" {
			conn.Write([]byte(message))
			go many(conn, ch)
			return
		} else {
			conn.Write([]byte(message))
		}
	}
}

func many(conn net.Conn, ch chan int) {
	message := ""
	for {
		fmt.Scanln(&message)
		if message == "/change" {
			conn.Write([]byte(message))
			go one(conn, ch)
			return
		} else {
			_, err := conn.Write([]byte(name + "说" + message))
			checkError(err)
		}
	}
}
