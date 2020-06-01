package ding_robot

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/zyjblockchain/sandy_log/log"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
)

type Robots interface {
	SendText(content string, atMobiles []string, isAtall bool) error
	SendMarkdown(content string, atMobiles []string, isAtall bool) error
}

var (
	ErrSendDingRobot = errors.New("send ding ding robot failed. ")
)

// 钉钉机器人推送
type Robot struct {
	WebHook string // 机器人的Hook地址
	lock    sync.Mutex
}

func NewRobot(webHook string) Robots {
	return &Robot{
		WebHook: webHook,
		lock:    sync.Mutex{},
	}
}

// SendText 发送普通类型的message
func (r *Robot) SendText(content string, atMobiles []string, isAtall bool) error { // content: 发送的文本内容。atMobiles:需要@的手机号列表。isAtall: 为true表示@所有人
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.send(&textMsg{
		MsgType: "text",
		Text: textParams{
			Content: content + ".",
		},
		At: AtParams{
			AtMobiles: atMobiles,
			IsAtAll:   isAtall,
		},
	})
}

func (r *Robot) SendMarkdown(content string, atMobiles []string, isAtall bool) error { // title: markdown的标题
	r.lock.Lock()
	defer r.lock.Unlock()
	// 读取content中的文本，第一句作为title
	sstr := strings.Split(content, "\n")
	formatStr := make([]string, 0)
	// 去掉空字符串
	for _, s := range sstr {
		if s != "" {
			formatStr = append(formatStr, s)
		}
	}
	if len(formatStr) == 0 {
		return errors.New("Can not send nil content to ding robot. ")
	}
	title := formatStr[0]
	newSstr := make([]string, 0)
	for _, s := range formatStr {
		if strings.HasPrefix(s, "AlarmReason") || strings.HasPrefix(s, "Detail") || strings.HasPrefix(s, "AlarmTime") || strings.HasPrefix(s, "NodeIP") {
			s = "#### " + s
		} else {
			s = "> " + s
		}
		newSstr = append(newSstr, s)
	}

	markdownText := strings.Join(newSstr, "\n\n> ")

	return r.send(&markdownMsg{
		MsgType: "markdown",
		Markdown: markdownParams{
			Title: title,
			Text:  markdownText,
		},
		At: AtParams{
			AtMobiles: atMobiles,
			IsAtAll:   isAtall,
		},
	})
}

// 发送消息到ding ding机器人的接口
func (r *Robot) send(msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Debugf("json marshal msg error. err: %v", err)
		return err
	}

	resp, err := http.Post(r.WebHook, "application/json; charset=utf-8", bytes.NewReader(data))
	if err != nil {
		log.Debug("http post error")
		return err
	}
	by, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Debugf("Response read error. err: %v", err)
		return err
	}
	defer resp.Body.Close()
	respMsg := &ResponseMsg{}
	err = json.Unmarshal(by, respMsg)
	if err != nil {
		log.Debugf("json unmarshal response msg error. err: %v. responseData: %s", err, string(by))
		return err
	}
	if respMsg.Errcode != 0 {
		log.Errorf("Send ding ding robot failed. errcode: %d, errMsg: %s", respMsg.Errcode, respMsg.Errmsg)
		return ErrSendDingRobot
	}
	return nil
}
