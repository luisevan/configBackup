package main

import (
	"bufio"
	"container/list"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

func main() {
	for {
	list := readNode("info.csv")
	dirName := "config" + time.Now().Format("2006-01-02_15-04-05")
	errMkdir := os.Mkdir(dirName, 0777)
	if errMkdir != nil {
		log.Println("mkdir err, err=", errMkdir)
	}
	seccessCount := sshToDo(list, dirName)
	fmt.Printf("共计%v台设备,本次完成%v台设备配置备份\n", list.Len(), seccessCount)
	time.Sleep(24 * time.Hour)

	}
}

// sshToDo ssh访问
func sshToDo(list *list.List, dirName string) (count int) {
	for node := list.Front(); node != nil; node = node.Next() {
		switch value := node.Value.(type) {
		case BatchNode:
			Ifb := SSHDo(value.Hostname, value.User, value.Password, value.IPPort, value.Cmd, dirName)
			if Ifb {
				count++
			} else {
				continue
			}
		}
	}
	return
}

// listNode listNode
func listNode(fileName string) *list.List {
	list := readNode(fileName)
	fmt.Printf("共计 %d 条数据\n", list.Len())
	i := 1
	for node := list.Front(); node != nil; node = node.Next() {
		switch value := node.Value.(type) {
		case BatchNode:
			fmt.Println(i, "  ", value.String())
		}
		i++
	}
	return list
}

// SSHDo SSH_do
func SSHDo(hostname, user, password, IPPort string, cmd, dirName string) bool {
	PassWd := []ssh.AuthMethod{ssh.Password(password)}
	// algor := []string{
	// 	"aes128-cbc", "aes256-cbc", "3des-cbc", "des-cbc",
	// }
	Conf := &ssh.ClientConfig{User: user, Auth: PassWd, HostKeyCallback: ssh.InsecureIgnoreHostKey()}
	Conf.Config.Ciphers = append(Conf.Config.Ciphers, "aes128-cbc", "aes256-ctr")
	// IPPort = strings.TrimRight(IPPort, "\n")
	Client, DialErr := ssh.Dial("tcp", IPPort, Conf)
	if DialErr != nil {
		log.Printf("%v ip+port:%v dial ssh conn error, err=%v\n", hostname, IPPort, DialErr)
		return false
	}
	defer Client.Close()
	command := cmd
	session, NewSessionErr := Client.NewSession()
	if NewSessionErr != nil {
		log.Println("NewSession error, err=", NewSessionErr)
		return false
	}
	defer session.Close()
	// hostname, _ := os.Hostname()
	filename := dirName + "/" + hostname + time.Now().Format("2006-01-02_15-04-05")
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		log.Println("open file error, err=", err)
		return false
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	// session.Stdout = os.Stdout
	// session.Stderr = os.Stderr

	ByteSlice, errOutput := session.CombinedOutput(command)
	if errOutput != nil {
		log.Printf("%v ip+port:%v Output error, err=%v\n", hostname, IPPort, errOutput)
		return false
	}
	// errun := session.Run(command)
	// if errun != nil {
	// 	log.Println("run cmd error, err=", errun)
	// 	return
	// }
	_, errWrite := writer.WriteString(string(ByteSlice))
	if errWrite != nil {
		log.Println("write config error, err=", errWrite)
		return false
	}
	writer.Flush()
	log.Printf("%v config backup successed", hostname)
	return true
}

// BatchNode BatchNode
type BatchNode struct {
	Hostname string
	User     string
	Password string
	IPPort   string
	Cmd      string
}

func (batchNode *BatchNode) String() string {
	return "ssh " + batchNode.Hostname + batchNode.User + "@" + batchNode.IPPort + "  with password: " + batchNode.Password + "  and run: " + batchNode.Cmd
}

func readNode(fileName string) *list.List {
	inputFile, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("在打开文件的时候出现错误\n文件存在吗?\n有权限吗?\n")
		return list.New()
	}
	defer inputFile.Close()
	batchNodeList := list.New()
	inputReader := bufio.NewReader(inputFile)
	for {
		inputString, err := inputReader.ReadString('\n')
		r := csv.NewReader(strings.NewReader(string(inputString)))
		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Println("Read csv error, err=", err)
				break
			}
			batchNode := BatchNode{record[0], record[1], record[2], record[3], record[4]}
			batchNodeList.PushBack(batchNode)
		}
		if err == io.EOF {
			break
		}
	}
	return batchNodeList
}
