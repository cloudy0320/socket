package main

import (
	"fmt"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	name string

	inputMessage string

	conn net.Conn
)

func main() {
	address, err := net.ResolveTCPAddr("tcp", "192.168.3.166:1200")
	checkError(err)
	conn, err = net.DialTCP("tcp", nil, address)
	checkError(err)

	readChan := make(chan []string)
	loginChan := make(chan string)
	registerChan := make(chan string)
	reconnectChan := make(chan bool)

	go heart()
	go receive(readChan, reconnectChan)

	go func() {
		for {
			select {
			case array := <-readChan:
				//fmt.Println(array)
				if array[0] == "message" {
					fmt.Println(array[1])
				} else if array[0] == "login" {
					loginChan <- array[1]
				} else if array[0] == "register" {
					registerChan <- array[1]
				} else if array[0] == "heart" {
					//fmt.Println("心跳检测")
					continue
				}
			case <-time.After(time.Second * 20):
				if reconnect(reconnectChan) {
					continue
				} else {
					fmt.Println("服务器链接失败")
					os.Exit(-1)
				}
			}
		}
	}()
A:
	for {
		fmt.Println("1：登录，2：注册")
		_, err := fmt.Scanln(&inputMessage)
		checkError(err)
		if !(inputMessage == "1" || inputMessage == "2") {
			fmt.Println("只能输入1和2")
			continue
		}
		if inputMessage == "1" {
			for {
				fmt.Println("请输入你要登录的姓名：")
				_, err := fmt.Scanln(&inputMessage)
				checkError(err)
				name = inputMessage
				_, err = conn.Write([]byte(request("login" + ";" + name + ";")))
				checkError(err)
				flag := <-loginChan
				if flag == "1" {
					fmt.Println("欢迎您：" + name)
					break A
				} else if flag == "2" {
					fmt.Println("这个账号已经被登录了")
					continue
				} else if flag == "3" {
					fmt.Println("根本没有这个账号")
					continue
				}
			}
		} else if inputMessage == "2" {
			for {
				fmt.Println("请输入你要注册的姓名：")
				_, err := fmt.Scanln(&inputMessage)
				checkError(err)
				name = inputMessage
				_, err = conn.Write([]byte(request("register" + ";" + name + ";")))
				checkError(err)
				flag := <-registerChan
				if flag == "1" {
					fmt.Println("恭喜你注册成功：" + name)
					break A
				} else if flag == "2" {
					fmt.Println("这个账号已经存在了")
					continue
				} else if flag == "3" {
					fmt.Println("服务器好像有点问题，稍后再试")
					continue
				}
			}
		}
	}
	for {
		fmt.Println("进入公共聊天室输入1，想要私聊输入2")
		_, err = fmt.Scanln(&inputMessage)
		checkError(err)
		if !(inputMessage == "1" || inputMessage == "2") {
			fmt.Println("只能输入1和2")
			continue
		}
		break
	}

	ch := make(chan int)
	if inputMessage == "1" {
		fmt.Println("欢迎进入公共聊天室")
		go many(ch)
		<-ch
	} else if inputMessage == "2" {
		go one(ch)
		<-ch
	}
}

func checkError(err error) {
	if err != nil {
		_, err := fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		reflect.ValueOf(err)
		checkError(err)
		os.Exit(1)
	}
}

func receive(readChan chan []string, ch chan bool) {
	for {
		l := make([]byte, 4)
		n, err := conn.Read(l)
		if err != nil || n == 0 {
			fmt.Println("连接好像有点问题，正在尝试重新连接")
			<-ch
			continue
		}
		length, err := strconv.Atoi(string(l))
		data := make([]byte, length)
		_, err = conn.Read(data)
		if err != nil || n == 0 {
			fmt.Println("连接好像有点问题，正在尝试重新连接")
			<-ch
			continue
		}
		array := strings.Split(string(data), ";")

		readChan <- array
	}
}

func one(ch chan int) {
Again:
	for {
		_, err := conn.Write([]byte(request("many" + ";" + "说" + ";" + "/list" + ";")))
		checkError(err)
		fmt.Println("请输入你要聊天的用户名")
		toName := ""
		_, err = fmt.Scanln(&toName)
		checkError(err)
		message := ""
		for {
			fmt.Scanln(&message)
			if message == "/quit" {
				continue Again
			} else if message == "/change" {
				go many(ch)
				return
			} else {
				_, err = conn.Write([]byte(request("one" + ";" + message + ";" + toName + ";" + name)))
				checkError(err)
			}
		}
	}
}

func many(ch chan int) {
	for {
		message := ""
		_, err := fmt.Scanln(&message)
		checkError(err)
		if message == "" {
			fmt.Println("不能输入为空")
			continue
		}
		if message == "/change" {
			go one(ch)
			return
		} else {
			_, err = conn.Write([]byte(request("many" + ";" + name + "说" + ";" + message + ";")))
			//_, err = conn.Write([]byte(request("many" + ";" + name + "说" + ";" + message + ";")))
			checkError(err)
		}
	}
}

func request(s string) string {
	msgLen := fmt.Sprintf("%04s", strconv.Itoa(len(s)))
	return msgLen + s
}

func heart() {
	for {
		conn.Write([]byte(request("heart")))
		time.Sleep(10 * time.Second)
	}
}

func reconnect(ch chan bool) bool {
	for i := 0; i < 7; i++ {
		if i == 6 {
			return false
		}
		address, err := net.ResolveTCPAddr("tcp", "192.168.3.166:1200")
		checkError(err)
		newConn, err := net.DialTCP("tcp", nil, address)
		if err == nil {
			conn = newConn
			if name != "" {
				_, err = conn.Write([]byte(request("reconnect" + ";" + name + ";")))
				checkError(err)
			}
			fmt.Println("服务器重连成功")
			ch <- true
			break
		} else {
			fmt.Println("连接好像有点问题，正在尝试重新连接")
		}
		time.Sleep(time.Second * 10)
	}
	return true
}
