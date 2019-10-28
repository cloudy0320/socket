package main

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type User struct {
	Id      int
	Name    string
	Message *Message `orm:"reverse(one)"`
}

type user struct {
	User
	conn net.Conn
}

type Message struct {
	Id    int
	From  string
	To    *User `orm:"rel(one)"`
	Text  string
	Stats int `orm:"default(1)"`
}

var users []user

var maxSize int = 10

func init() {
	orm.RegisterDriver("mysql", orm.DRMySQL)

	orm.RegisterDataBase("default", "mysql", "root:123456@/socket?charset=utf8")

	orm.RegisterModel(new(User), new(Message))

	orm.RunSyncdb("default", false, true)

	get()
}

func main() {
	service := ":1200"
	tcpAddr, err := net.ResolveTCPAddr("tcp", service)
	checkError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	defer del(conn)

	readChan := make(chan []string)
	writeLoginChan := make(chan []string)
	writeOneChan := make(chan []string)
	writeManyChan := make(chan []string)
	writeRegisterChan := make(chan []string)
	stopChan := make(chan bool)
	var myName string

	go readConn(conn, readChan, stopChan)
	go writeLoginConn(conn, writeLoginChan, stopChan, &myName)
	go writeRegisterConn(conn, writeRegisterChan, stopChan, &myName)
	go writeOneConn(conn, writeOneChan, stopChan, &myName)
	go writeManyConn(conn, writeManyChan, stopChan, &myName)

	for {
		select {
		case array := <-readChan:
			if array[0] == "login" {
				writeLoginChan <- array
			} else if array[0] == "register" {
				writeRegisterChan <- array
			} else if array[0] == "one" {
				writeOneChan <- array
			} else if array[0] == "many" {
				writeManyChan <- array
			} else if array[0] == "reconnect" {
				for i, v := range users {
					if array[1] == v.Name {
						users[i].conn = conn
						check(array[1], conn)
					}
				}
			} else if array[0] == "heart" {
				_, err := conn.Write([]byte(request("heart;")))
				checkError(err)
				continue
			}
		case stop := <-stopChan:
			if stop {
				return
			}
		case <-time.After(time.Second * 30):
			fmt.Println("连接中断")
			return
		}
	}
}

func readConn(conn net.Conn, readChan chan<- []string, stopChan chan<- bool) {
	for {
		l := make([]byte, 4)
		dataTemp := make([]byte, maxSize)
		data := make([]byte, 0)
		n, err := conn.Read(l)
		if err != nil || n == 0 {
			fmt.Println(err)
			stopChan <- true
		}
		fmt.Println("读取的长度：" + string(l))
		length, err := strconv.Atoi(string(l))
		if err != nil {
			fmt.Println(err)

		}
		fmt.Println("实际长度:" + string(length))
		_, err = conn.Read(dataTemp)
		if err != nil {
			fmt.Println(err)
			stopChan <- true
		}
		if n == 0 {
			stopChan <- true
		}
		fmt.Println("数据体：" + string(dataTemp))
		if length <= maxSize {
			data = dataTemp[:length]
		} else {
			data = append(data, dataTemp...)
			for i := 0; i < length/maxSize; i++ {
				if i+1 == length/maxSize {
					dataTemp := make([]byte, length%maxSize)
					_, err = conn.Read(dataTemp)
					if err != nil {
						fmt.Println(err)
						stopChan <- true
					}
					fmt.Println("最后一次循环：" + string(dataTemp))
					data = append(data, dataTemp...)
				} else {
					_, err = conn.Read(dataTemp)
					if err != nil {
						fmt.Println(err)
						stopChan <- true
					}
					data = append(data, dataTemp...)
				}
			}
		}
		array := make([]string, 128)
		array = strings.Split(string(data), ";")

		fmt.Println("Received:", array)

		readChan <- array
	}
}

func writeLoginConn(conn net.Conn, writeChan <-chan []string, stopChan chan<- bool, name *string) {
	for {
		array := <-writeChan
		*name = array[1]
		for i, v := range users {
			if array[1] == v.Name && v.conn == nil {
				users[i].conn = conn
				_, err := conn.Write([]byte(request("login;" + "1;")))
				checkError(err)
				check(*name, conn)
			} else if array[1] == v.Name && v.conn != nil {
				_, err := conn.Write([]byte(request("login;" + "2;")))
				checkError(err)
				check(*name, conn)
			}
		}
	}
}

func writeRegisterConn(conn net.Conn, writeChan <-chan []string, stopChan chan<- bool, name *string) {
	o := orm.NewOrm()
	for {
		array := <-writeChan
		*name = array[1]
		u := User{Name: array[1]}
		o.Begin()
		if created, _, err := o.ReadOrCreate(&u, "Name"); err == nil {
			if created {
				o.Commit()
				_, err := conn.Write([]byte(request("register;" + "1;")))
				checkError(err)
				//stopChan <- true
			} else {
				_, err := conn.Write([]byte(request("register;" + "0;")))
				checkError(err)
				continue
			}
		} else {
			o.Rollback()
			_, err := conn.Write([]byte(request("register;" + "-1;")))
			checkError(err)
		}
	}
}

func writeOneConn(conn net.Conn, writeChan <-chan []string, stopChan chan<- bool, name *string) {
	for {
		array := <-writeChan
		if len(array) == 4 {
			for _, v := range users {
				if array[2] == v.Name && v.conn != nil {
					_, err := v.conn.Write([]byte(request("message;" + *name + ";对你说;" + array[1] + ";")))
					checkError(err)
				} else if array[2] == v.Name && v.conn == nil {
					o := orm.NewOrm()
					o.Begin()
					m := Message{Text: array[1], To: &v.User, From: *name}
					_, err := o.Insert(&m)
					if err != nil {
						o.Rollback()
					} else {
						o.Commit()
					}
				}
			}
		} else {
			continue
		}
	}
}

func writeManyConn(conn net.Conn, writeChan <-chan []string, stopChan chan<- bool, name *string) {
	for {
		array := <-writeChan
		if array[2] == "/list" {
			var online []string
			for _, v := range users {
				if *name == v.Name {
					online = append(online, v.Name+"(你自己)")
				} else if v.conn != nil {
					online = append(online, v.Name+"(在线)")
				} else {
					online = append(online, v.Name+"(不在线)")
				}
			}
			list := strings.Join(online, "\n")
			_, err := conn.Write([]byte(request("message;" + list + ";")))
			checkError(err)
		} else if array[2] == "/change" {
			continue
		} else {
			for _, v := range users {
				if v.conn != nil {
					_, err := v.conn.Write([]byte(request("message;" + array[1] + ";" + array[2] + ";")))
					checkError(err)
				}
			}
		}
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func del(conn net.Conn) {
	fmt.Println(conn.RemoteAddr().String() + "已经断开")
	for i, v := range users {
		if v.conn == conn {
			users[i].conn = nil
		}
	}
}

func get() {
	var usersData []User
	userTemp := new(user)
	o := orm.NewOrm()
	_, err := o.QueryTable("user").All(&usersData)
	for _, v := range usersData {
		userTemp.User = v
		users = append(users, *userTemp)
	}
	if err != nil {
		fmt.Println("获取数据库失败")
		os.Exit(1)
	}
}

func check(name string, conn net.Conn) {
	var response []string
	o := orm.NewOrm()
	userTemp := User{Name: name}
	err := o.Read(&userTemp, "Name")
	if err == orm.ErrNoRows {
		fmt.Println("查询不到")
	}
	var messages []Message
	n, err := o.QueryTable("message").Filter("To", userTemp.Id).Filter("stats", 0).All(&messages)
	if err == nil && n != 0 {
		for _, v := range messages {
			response = append(response, v.From+"对你说"+v.Text)
			v.Stats = 1
			o.Update(&v)
		}
		response = append(response, "以上是离线信息")
		s := strings.Join(response, "\n")
		_, err = conn.Write([]byte(request("message;" + s + ";")))
		checkError(err)
	}
}

func request(s string) string {
	msgLen := fmt.Sprintf("%04s", strconv.Itoa(len(s)))
	return msgLen + s
}
