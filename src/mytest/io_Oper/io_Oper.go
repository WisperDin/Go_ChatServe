// io_Oper
package io_Oper

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func RecvMsg(clnSck net.Conn, buf []byte) (string, int) {
	dataLen, err := clnSck.Read(buf)
	if err != nil {
		fmt.Println(err)
		//下线
		return "", -1
	}
	buf = buf[:dataLen]
	sap := clnSck.RemoteAddr().String()
	//keepMsg(msg, sap)
	return sap, dataLen
}

func GetMsg() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	msg, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println(err.Error())
	}
	return msg, err

}
