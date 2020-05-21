package logics

import (
	"errors"
	"github.com/zyjblockchain/sandy_log/log"
	"github.com/zyjblockchain/tt_tac/models"
	"github.com/zyjblockchain/tt_tac/utils"
	"strings"
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
	// 判断地址是否在数据库已经存在,如果已经存在则重载支付密码
	u, err := new(models.User).GetUserByAddress(addr.String())
	if err == nil {
		// 证明数据库中已经存在该地址,则重载支付密码
		err = u.SetPassword(l.Password)
		if err != nil {
			log.Errorf("对原始密码加密失败：%v", err)
			return "", err
		}
		// update数据库
		if err := u.Update(); err != nil {
			log.Errorf(" update user 表失败：%v", err)
			return "", err
		}
		// 返回address
		return u.Address, nil
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

// 导出私钥
type Export struct {
	Address  string `json:"address" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

// ExportPrivate
func (e *Export) ExportPrivate() (string, error) {
	// 通过address查询出user
	u, err := new(models.User).GetUserByAddress(e.Address)
	if err != nil {
		log.Errorf("通过address从表中查询user失败， err: %v, address: %s", err, e.Address)
		return "", err
	}
	// 校验password
	if !u.CheckPassword(e.Password) {
		log.Errorf("password交易失败， 输入的密码不正确")
		return "", errors.New("输入的密码有误")
	}
	// 返回private
	private, err := utils.DecryptPrivate(u.PrivateCrypted)
	if err != nil {
		return "", err
	} else {
		return strings.ToUpper(private[2:]), nil // 去除0x并把转成大写形式：F234120DE07D7F5CE27EAA1D7B954F55BDC49E6C3B2B19FB78C5000A191CEE4F
	}
}
