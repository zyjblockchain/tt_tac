package logics

import (
	"errors"
	"github.com/zyjblockchain/sandy_log/log"
	"github.com/zyjblockchain/tt_tac/models"
	"github.com/zyjblockchain/tt_tac/utils"
)

type CreateUser struct {
	Password string `json:"password" binding:"required,min=6"`
}

// CreateUser 创建用户
func (c *CreateUser) CreateUser() (string, error) {
	// 生成密钥对
	account, err := utils.GenerateEthAccount()
	if err != nil {
		log.Errorf("密钥对生成失败：%v", err)
		return "", err
	}
	// 私钥加密
	encryptPrivate, err := utils.EncryptPrivate(account.Private)
	if err != nil {
		log.Errorf("对私钥进行aes加密失败：%v", err)
		return "", err
	}
	user := &models.User{
		Address:        account.Address,
		PrivateCrypted: encryptPrivate,
	}
	// password 加密
	err = user.SetPassword(c.Password)
	if err != nil {
		log.Errorf("对原始密码加密失败：%v", err)
		return "", err
	}
	// 保存到数据库
	newUser, err := new(models.User).AddUser(user)
	if err != nil {
		log.Errorf("保存user数据库失败：%v", err)
		return "", err
	}
	// 返回address
	return newUser.Address, nil
}

// 通过私钥导入账户
type LeadUser struct {
	Private  string `json:"private" binding:"required,min=64,max=66"`
	Password string `json:"password" binding:"required,min=6"`
}

// LeadUser 通过私钥导入账户
func (l *LeadUser) LeadUser() (string, error) {
	// 通过私钥生成地址
	addr, err := utils.PrivateToAddress(l.Private)
	if err != nil {
		return "", err
	}
	// 判断地址是否在数据库已经存在
	_, err = new(models.User).GetUserByAddress(addr.String())
	if err == nil {
		// 证明数据库中已经存在该地址 todo
		return "", errors.New("数据库中已经存在该地址")
	}

	// 对私钥加密存储
	encryptPrivate, err := utils.EncryptPrivate(l.Private)
	if err != nil {
		log.Errorf("对私钥进行aes加密失败：%v", err)
		return "", err
	}

	user := &models.User{
		Address:        addr.String(),
		PrivateCrypted: encryptPrivate,
	}
	// user设置密码
	err = user.SetPassword(l.Password)
	if err != nil {
		log.Errorf("对原始密码加密失败：%v", err)
		return "", err
	}
	newUser, err := new(models.User).AddUser(user)
	if err != nil {
		log.Errorf("保存user数据库失败：%v", err)
		return "", err
	}
	// 返回address
	return newUser.Address, nil
}
