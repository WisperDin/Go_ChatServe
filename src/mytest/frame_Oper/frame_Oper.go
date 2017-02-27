// frame_Oper
package frame_Oper

import (
	"encoding/json"
	//"errors"
	//"fmt"
	"time"
)

//坑爹之一，，在结构体内变量名首位小写表示私有，大写表示公有
type AbstFrame struct {
	Sign        int   `json:"protoSign"`
	Length      int64 `json:"msgLength"`
	MsgType     int   `json:"msgType"`
	SenderTimer int64 `json:"senderTimer"`

	flag    int    `json:"-"`
	DataBuf []byte `json:"-"`
	//需要反馈到的ip
	Sap string `json:"-"`

	DataFrame `json:",omitempty"`
	FbFrame   `json:",omitempty"`
}
type Dst struct {
	DstItem string `json:"userName,omitempty"`
}
type User struct {
	UserName  string `json:"userName,omitempty"`
	UserPW    string `json:"userPWD,omitempty"`
	UserID    int    `json:"userID,omitempty"`
	UserToken string `json:"userToken,omitempty"`
}
type DstSlice []Dst     //表示Dst的数组
type DataFrame struct { //22为私聊模式
	Feedback string `json:"feedbackType,omitempty"`
	User     `json:"sender,omitempty"`
	DstSlice `json:"userList,omitempty"`
	Msg      string `json:"payLoad,omitempty"`
}

type ActionStatus struct {
	AMsgType int    `json:"actionMsgType,omitempty"`
	RslMsg   string `json:"actionRslMsg,omitempty"`
	RslCode  int    `json:"actionRslCode,omitempty"` ////////////////////////定义1为正常反馈码
}
type FbFrame struct {
	ActionStatus `json:"actionStatus,omitempty"`
}

//构造帧
func GetFrame(dst []string, msg string, Sap string, flag int) AbstFrame {
	var curframe AbstFrame
	switch flag {
	case 1: //构建消息数据帧
		{
			curframe = AbstFrame{DataFrame: DataFrame{User: User{UserName: "serve"}}}
			curframe.DstSlice = make([]Dst, len(dst))
			for i, v := range dst {
				curframe.DstSlice[i].DstItem = v
			}

			curframe.Msg = msg
			curframe.Feedback = "c" //需要客户端反馈
			curframe.MsgType = 22
			curframe.Sap = Sap
			break
		}
	case 2: //构建反馈数据帧
		{
			curframe = AbstFrame{FbFrame: FbFrame{ActionStatus: ActionStatus{22, msg, 1}}}
			curframe.MsgType = 2
			break
		}
	}
	curframe.SenderTimer = time.Now().Unix()
	return curframe

}

//将收到的帧解包
func UnMarshal(buf []byte, dataLen int, Sap string) (AbstFrame, error) {
	var frame AbstFrame
	if buf == nil || dataLen == 0 || Sap == "" {
		var err error
		return frame, err
	}
	err := json.Unmarshal(buf[:dataLen], &frame)
	if err == nil {
		frame.DataBuf = make([]byte, dataLen)
		copy(frame.DataBuf, buf[:dataLen])
		frame.Sap = Sap
		return frame, err
	}
	return frame, err
}

//构造信息帧或者反馈帧的字节数据，放到dataBuf里面待发送(JSON编码)
func (frame *AbstFrame) Marshal() error {
	frame.Sign = 142857
	frame.Length = 1073742063
	/*
		switch frame.flag {
		case 1:
			{ //信息帧
				if len(frame.Msg) <= 0 {
					return errors.New("param Src/Dst/Msg is empty!")
				}
			}
		case 2:
			{

			}
		}*/
	var err error
	frame.DataBuf, err = json.Marshal(frame) //先解码一次获取除长度外的其他量
	if err != nil {
		return err
	}
	frame.Length = int64(len(frame.DataBuf) | 0x40000000) //fmt.Sprintf("%#08X", len(frame.DataBuf))
	//frame.Length = fmt.Sprintf("%#08X", len(frame.DataBuf)) //得到其他量后再获得databuf的长度
	frame.DataBuf, err = json.Marshal(frame)
	return nil
}
