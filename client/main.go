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

	maxSize int = 128
)

func main() {
	address, err := net.ResolveTCPAddr("tcp", "192.168.3.166:1200")
	checkError(err)
	conn, err := net.DialTCP("tcp", nil, address)
	checkError(err)

	readChan := make(chan []string)
	loginChan := make(chan string)
	registerChan := make(chan string)

	go heart(conn)
	go receive(conn, readChan)

	go func() {
		for {
			select {
			case array := <-readChan:
				fmt.Println(array)
				if array[0] == "message" {
					fmt.Println(array[1])
				} else if array[0] == "login" {
					loginChan <- array[1]
				} else if array[0] == "register" {
					registerChan <- array[1]
				} else if array[0] == "heart" {
					fmt.Println("心跳检测")
					continue
				}
			case <-time.After(time.Second * 20):
				for i := 0; i < 6; i++ {
					fmt.Println("连接好像有点问题，正在尝试重连")
					address, err := net.ResolveTCPAddr("tcp", "192.168.3.166:1200")
					checkError(err)
					conn, err = net.DialTCP("tcp", nil, address)
					fmt.Println(err)
					if err == nil {
						if name != "" {
							_, err = conn.Write([]byte(request("reconnect" + ";" + name + ";")))
							checkError(err)
						}
						break
					}
					if i == 5 {
						fmt.Println("连接中断")
						os.Exit(-1)
					}
					time.Sleep(time.Second * 10)
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
	//for {
	//	recvData = make([]byte, 128)
	//	fmt.Print("输入你的姓名:")
	//	_, err = fmt.Scanln(&name)
	//	checkError(err)
	//	_, err = conn.Write([]byte(request(kind + ";" + name + ";")))
	//	checkError(err)
	//	n, err := conn.Read(recvData)
	//	if err != nil {
	//		fmt.Println("好像哪里出错了")
	//		os.Exit(0)
	//	}
	//	if n == 0 {
	//		fmt.Println("服务器和你断开了链接")
	//		os.Exit(0)
	//	}
	//	if flag == "1" {
	//		if string(recvData[0]) == "1" {
	//			fmt.Println("您已成功登陆 " + name)
	//			break
	//		} else if string(recvData[0]) == "0" {
	//			fmt.Println("没有这个名称，请先去注册")
	//			continue
	//		} else if string(recvData[0]) == "2" {
	//			fmt.Println("这个号已经被登录了")
	//			continue
	//		}
	//	} else if flag == "2" {
	//		if string(recvData[0]) == "1" {
	//			fmt.Println("您已成功注册" + name)
	//			break
	//		} else if string(recvData[0]) == "0" {
	//			fmt.Println("这个名称已经被注册了")
	//			continue
	//		} else {
	//			for {
	//				recvData = make([]byte, 128)
	//				fmt.Print("输入你的姓名:")
	//				_, err = fmt.Scanln(&name)
	//				checkError(err)
	//				_, err = conn.Write([]byte(request(kind + ";" + name + ";")))
	//				checkError(err)
	//				for {
	//					recvData = make([]byte, 128)
	//					fmt.Print("输入你的姓名:")
	//					_, err = fmt.Scanln(&name)
	//					checkError(err)
	//					_, err = conn.Write([]byte(request(kind + ";" + name + ";")))
	//					checkError(err)
	//					n, err := conn.Read(recvData)
	//					if err != nil {
	//						fmt.Println("好像哪里出错了")
	//						os.Exit(0)
	//					}
	//					if n == 0 {
	//						fmt.Println("服务器和你断开了链接")
	//						os.Exit(0)
	//					}
	//					if flag == "1" {
	//						if string(recvData[0]) == "1" {
	//							fmt.Println("您已成功登陆 " + name)
	//							break
	//						} else if string(recvData[0]) == "0" {
	//							fmt.Println("没有这个名称，请先去注册")
	//							continue
	//						} else if string(recvData[0]) == "2" {
	//							fmt.Println("这个号已经被登录了")
	//							continue
	//						}
	//					} else if flag == "2" {
	//						if string(recvData[0]) == "1" {
	//							fmt.Println("您已成功注册" + name)
	//							break
	//						} else if string(recvData[0]) == "0" {
	//							fmt.Println("这个名称已经被注册了")
	//							continue
	//						} else {
	//							fmt.Println("服务器好像有点问题")
	//							continue
	//						}
	//					}
	//				}
	//				n, err := conn.Read(recvData)
	//				if err != nil {
	//					fmt.Println("好像哪里出错了")
	//					os.Exit(0)
	//				}
	//				if n == 0 {
	//					fmt.Println("服务器和你断开了链接")
	//					os.Exit(0)
	//				}
	//				if flag == "1" {
	//					if string(recvData[0]) == "1" {
	//						fmt.Println("您已成功登陆 " + name)
	//						break
	//					} else if string(recvData[0]) == "0" {
	//						fmt.Println("没有这个名称，请先去注册")
	//						continue
	//					} else if string(recvData[0]) == "2" {
	//						fmt.Println("这个号已经被登录了")
	//						continue
	//					}
	//				} else if flag == "2" {
	//					if string(recvData[0]) == "1" {
	//						fmt.Println("您已成功注册" + name)
	//						break
	//					} else if string(recvData[0]) == "0" {
	//						fmt.Println("这个名称已经被注册了")
	//						continue
	//					} else {
	//						fmt.Println("服务器好像有点问题")
	//						continue
	//					}
	//				}
	//			}
	//			fmt.Println("服务器好像有点问题")
	//			continue
	//		}
	//	}
	//}
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
		go many(conn, ch)
		<-ch
	} else if inputMessage == "2" {
		go one(conn, ch)
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

func receive(conn net.Conn, readChan chan []string) {
	for {
		l := make([]byte, 4)
		dataTemp := make([]byte, maxSize)
		data := make([]byte, 0)
		n, err := conn.Read(l)
		if err != nil || n == 0 {
			fmt.Println(err)
			os.Exit(1)
		}
		length, err := strconv.Atoi(string(l))
		if err != nil {
			fmt.Println(err)
		}
		_, err = conn.Read(dataTemp)
		if err != nil || n == 0 {
			fmt.Println(err)
			os.Exit(1)
		}
		if length <= maxSize {
			data = dataTemp[:length]
		} else {
			data = append(data, dataTemp...)
			for i := 0; i < length/maxSize; i++ {
				if i+1 == length/maxSize {
					dataTemp := make([]byte, length%maxSize)
					n, err = conn.Read(dataTemp)
					if err != nil || n == 0 {
						fmt.Println(err)
						os.Exit(1)
					}
					data = append(data, dataTemp...)
				} else {
					n, err = conn.Read(dataTemp)
					if err != nil || n == 0 {
						fmt.Println(err)
						os.Exit(1)
					}
					data = append(data, dataTemp...)
				}
			}
		}
		array := make([]string, 128)
		array = strings.Split(string(data), ";")

		readChan <- array
	}
}

func one(conn net.Conn, ch chan int) {
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
				go many(conn, ch)
				return
			} else {
				_, err = conn.Write([]byte(request("one" + ";" + message + ";" + toName + ";")))
				checkError(err)
			}
		}
	}
}

func many(conn net.Conn, ch chan int) {
	for {
		message := ""
		_, err := fmt.Scanln(&message)
		checkError(err)
		if message == "" {
			fmt.Println("不能输入为空")
			continue
		}
		if message == "/change" {
			go one(conn, ch)
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

func heart(conn net.Conn) {
	for {
		_, err := conn.Write([]byte(request("heart")))
		if err != nil {
			fmt.Println("连接好像出了点问题，正在尝试重连")
		}
		time.Sleep(10 * time.Second)
	}
}
