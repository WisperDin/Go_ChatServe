// cstSrv
package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	//"./frame_Oper"
	"mytest/frame_Oper"
	//"./io_Oper"
	"mytest/io_Oper"
	//"./sql_Oper"
	"mytest/sql_Oper"

	_ "github.com/go-sql-driver/mysql"
)

var clnFeedBackChannel chan frame_Oper.AbstFrame
var clnOnLineChannel chan net.Conn
var clnOffLineChannel chan net.Conn

//var msgChannel chan string
var clnMsgChannel chan frame_Oper.AbstFrame
var serveflag bool

var userList map[string]int //id和用户名
func showOnLines(arg_conns map[string]net.Conn) {
	fmt.Println("Online Number: " + strconv.Itoa(len(arg_conns)))
}

var fbm map[string]string
var connList map[string]net.Conn

//处理反馈帧
func HandleFbF(fbFrame *frame_Oper.AbstFrame) {
	v, ok := fbm[fbFrame.Sap] //看下这个用户的ip是否在fbm中
	if ok && fbFrame.RslCode == 1 && fbFrame.MsgType == 2 {
		fmt.Println(v + " feedback status:" + fbFrame.RslMsg)
	}
}
func SendFBF(fbMsg string, clnS net.Conn) error {
	fbFrame, err0 := MakeFbFrame(fbMsg)
	//fmt.Println("Send feedBack :" + string(fbFrame.DataBuf)) //test
	if err0 != nil {
		fmt.Println(err0.Error())
		return err0
	}
	_, err1 := clnS.Write(fbFrame.DataBuf)
	if err1 != nil {
		fmt.Println(err1.Error())
		return err1
	}
	return nil
}
func userLogin(transMsg *frame_Oper.AbstFrame) net.Conn {
	//登录操作，将连接列表，反馈列表里面的ip换成userName
	v, ok := connList[transMsg.Sap] //看下这个用户的ip是否在连接列表中
	if ok {
		delete(connList, transMsg.Sap)
		connList[transMsg.UserName] = v //用名称去替换ip
		fbm[transMsg.Sap] = transMsg.UserName
		fmt.Println(transMsg.UserName + "-登陆成功")
		return v
		//return //注册功能完成
	}
	return nil
}

//处理注册帧
func HandleInitUser(transMsg *frame_Oper.AbstFrame) {
	var srccln net.Conn
	var errFb error
	_, ok := userList[transMsg.UserName] //看下这个用户的name是否在用户列表中
	//fmt.Println(userList)
	if !ok { //用户未注册
		if transMsg.UserPW == "6666" { //如果口令正确则注册，没有口令则登录
			id := -1
			if transMsg.UserID != 0 {
				id = transMsg.UserID
			}
			sql_Oper.InitUser(id, transMsg.UserName)
			userList[transMsg.UserName] = id
			fmt.Println(transMsg.UserName + " -注册成功")
			srccln = userLogin(transMsg) //登录成功后返回socket对象
			if srccln != nil {
				errFb = SendFBF("SUCCESS Init and login", srccln)
				if errFb != nil {
					return
				}
			}
			return
		} else {
			fmt.Println(transMsg.UserName + " -口令不正确")
			return
		}

	} else { //用户已经注册了
		fmt.Println(transMsg.UserName + "-已经注册了,将不再注册,直接登录")

		srccln = userLogin(transMsg)
		if srccln != nil {
			errFb = SendFBF("SUCCESS login", srccln)
			if errFb != nil {
				return
			}
		}
	}

}

//处理信息帧
func HandleMsgF(transMsg *frame_Oper.AbstFrame) {
	//正常聊天功能
	//检查是否存在这个发送源,并暂存这个发送源的socket（可能要用来发反馈信息）
	var srccln net.Conn
	v, ok := connList[transMsg.UserName] //检查connList里有没有这个用户名
	//没有且发送源不是服务器就是错误情况
	if !ok && transMsg.UserName != "serve" { //发送者如果是服务器的话也会不在connList
		fmt.Println(transMsg.UserName + "not in connList")
		//return
		return
	}
	srccln = v
	//检查发送信息
	if transMsg.DataBuf == nil || len(transMsg.DataBuf) <= 0 {
		fmt.Println("transMsg.DataBuf is nil/empty")
		return
	}
	//找到发送目标，可能有多个
	for _, u := range transMsg.DstSlice { //遍历结构体的数组.结构体内含有一个string类型的dst
		//fmt.Println(transMsg.Dst)
		dstUser, ok := connList[u.DstItem]
		if u.DstItem == "serve" { //发送给服务器的
			fmt.Println("Receive from " + transMsg.UserName + ":" + transMsg.Msg) //直接显示
			//需要服务器反馈
			if transMsg.Feedback == "s" {
				///发送反馈信息到客户端
				//var f FbFrame
				errFb := SendFBF("SUCCESS REC", srccln)
				if errFb != nil {
					continue
				}
			}
			continue
		}
		if !ok {
			fmt.Println("no dst")
			//fmt.Println(connList)
			if srccln != nil {
				errFb := SendFBF("Not Exist This user", srccln)
				if errFb != nil {
					return
				}
			}
			continue
		}
		//发送信息
		_, err := dstUser.Write(transMsg.DataBuf)
		if err != nil {
			fmt.Println(err.Error())
			clnOffLineChannel <- dstUser
			continue
		}
		if transMsg.Feedback == "c" && transMsg.UserName != "serve" { //需要客户端反馈，且发送源不是服务器
			///发送反馈信息到客户端
			errFb := SendFBF("SUCCESS REC", srccln)
			if errFb != nil {
				continue
			}
		}
		//fmt.Println("Send from " + transMsg.UserName + ":" + string(transMsg.DataBuf))
		continue
	}

}

//管理客户端
func clnMgr() {

	//通道初始化
	clnOnLineChannel = make(chan net.Conn)
	clnOffLineChannel = make(chan net.Conn)
	//msgChannel = make(chan string)
	clnMsgChannel = make(chan frame_Oper.AbstFrame)
	clnFeedBackChannel = make(chan frame_Oper.AbstFrame)
	fbm = make(map[string]string) //sap和userName的map
	connList = make(map[string]net.Conn)
	for {
		select {
		case fbFrame := <-clnFeedBackChannel: //反馈通道
			{
				HandleFbF(&fbFrame)
				break
			}
		case transMsg := <-clnMsgChannel: //消息通道，转发客户端数据帧,服务器向客户端发消息的数据帧,还要转发反馈帧（客户端注册的消息帧，客户端登录的消息帧，无消息正文）
			{
				if len(transMsg.UserName) > 0 {
					if len(transMsg.Msg) <= 0 { //消息为空则注册或登录用户
						HandleInitUser(&transMsg)
						continue
					}
					HandleMsgF(&transMsg)
					continue
				}
				break
			}

		case clnSck := <-clnOnLineChannel:
			{ //有用户上线，记录这个用户的IP与其socket对象
				clnSap := clnSck.RemoteAddr().String()
				fmt.Println(clnSap + " online")
				connList[clnSap] = clnSck
				fbm[clnSap] = "null"
				showOnLines(connList)
				break
			}

		case clnSck := <-clnOffLineChannel:
			{
				for k, v := range connList {
					if v == clnSck {
						fmt.Println(k + " offline")
						delete(connList, k)
						v.Close()
						showOnLines(connList)
						break
					}
				}
				break
			}
		} //end select
	} //end for
} //end clnMgr
/////////////////////////////////代办 sendertime，用户注册机制
func MakeFbFrame(msg string) (frame_Oper.AbstFrame, error) {
	fbFrame := frame_Oper.GetFrame(nil, msg, "", 2) //2表示反馈帧
	err := fbFrame.Marshal()
	if err != nil {
		fmt.Println("Marshal failed")
	}
	return fbFrame, err
}
func recv(clnSck net.Conn) {
	buf := make([]byte, 1024)
	for {
		sap, dataLen := io_Oper.RecvMsg(clnSck, buf)
		if dataLen == -1 {
			clnOffLineChannel <- clnSck
			return
		}
		//构建数据帧
		msgFrame, err := frame_Oper.UnMarshal(buf, dataLen, sap) //1表示消息帧
		//fmt.Println("receive :" + string(msgFrame.DataBuf))      //测试构建的帧

		if err != nil {
			fmt.Println("ERROR IN UNMARSHALL")
			return
		}
		//fmt.Println(string(msgFrame.DataBuf))
		if msgFrame.MsgType == 22 {
			clnMsgChannel <- msgFrame //正常消息通道
		} else {
			clnFeedBackChannel <- msgFrame //反馈信息通道
		}

	}
}

func SimpMsg(peermsg string) int {
	var dstPos, msgPos int
	//for {
	peermsg = strings.Replace(
		strings.Replace(
			strings.Replace(
				strings.Replace(
					strings.Replace(peermsg, "　", " ", -1),
					"，", ",", -1),
				"＠", "@", -1),
			", ", ",", -1),
		" ,", ",", -1)
	//@开头的是目的
	dstPos = strings.Index(peermsg, "@")

	if dstPos == 0 {
		//空格后面的是信息正文
		msgPos = strings.IndexAny(peermsg, " ")
		return msgPos
	}
	return -1
}

//不断获取键盘输入发送到客户端，其实最后使用了转发信息的通道，只是将发送源改成了服务器
func sendMsg() {
	for {
		msg, err := io_Oper.GetMsg()
		if err != nil {
			return
		}
		msgPos := SimpMsg(msg)
		if msgPos <= 0 {
			fmt.Println("empty msg")
			return
		}
		//@张三,李四  将这些发送对象分开
		Dst := strings.Split(msg[1:msgPos], ",")
		Msg := msg[msgPos:]
		Sap := "127.0.0.1"
		msgFrame := frame_Oper.GetFrame(Dst, Msg, Sap, 1) //1表示消息帧
		err1 := msgFrame.Marshal()
		//fmt.Println(string(msgFrame.DataBuf))
		//err1 := transMsg.Marshal()
		if err1 != nil {
			fmt.Println("Marshal failed")
			return
		}
		clnMsgChannel <- msgFrame
	}
}

func main() {

	sql_Oper.SetupDB()

	//socket

	srvSck, err := net.Listen("tcp", ":6666")

	if err != nil {
		fmt.Println(err)
		return
	}
	defer srvSck.Close()
	userList = make(map[string]int)
	sql_Oper.ImpUser(userList) //从数据库导入用户列表
	fmt.Println("已注册用户列表")
	fmt.Println(userList)
	go clnMgr()

	go sendMsg()
	for {
		clnSck, err := srvSck.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		clnOnLineChannel <- clnSck
		go recv(clnSck)
		go sendMsg()
	}

}
