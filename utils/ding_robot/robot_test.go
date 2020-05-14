package ding_robot

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var webHook = "https://oapi.dingtalk.com/robot/send?access_token=b4ff4c39e202803e650886c6a93003e5423796525d9ff1f777c13a2a03762da8"

func TestRobot_SendText(t *testing.T) {
	robot := NewRobot(webHook)
	content := "tac告警系统\n钉钉机器人测试\n"
	err := robot.SendText(content, []string{"18382255942"}, false)
	assert.NoError(t, err)
}
func TestRobot_SendMarkdown(t *testing.T) {
	robot := NewRobot(webHook)
	markdownText := "> #### tac\n" +
		"> ![screenshot](https://upload.wikimedia.org/wikipedia/commons/thumb/d/d9/Kim_Jong-un_IKS_2018.jpg/489px-Kim_Jong-un_IKS_2018.jpg)\n" +
		"> ###### 10点20分发布 [天气](https://www.seniverse.com/ )"
	err := robot.SendMarkdown(markdownText, nil, true)
	assert.NoError(t, err)
}
