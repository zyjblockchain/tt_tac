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

// 修改支付密码
type ModifyPassword struct {
	Address     string `json:"address" binding:"required"`
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func (m *ModifyPassword) ModifyPwd() error {
	// 查看地址是否在数据库中存在
	user, err := new(models.User).GetUserByAddress(m.Address)
	if err != nil {
		// 表示不存在或者查询失败则返回error
		log.Errorf("通过address 查询user error: %v", err)
		return err
	}
	// 存在则查看oldPassword对不对
	if !user.CheckPassword(m.OldPassword) {
		// 密码不正确
		return errors.New("旧的密码验证不通过")
	}
	// 修改密码
	if err := user.SetPassword(m.NewPassword); err != nil {
		log.Errorf("对原始密码加密失败：%v", err)
		return err
	}
	// update
	if err := user.Update(); err != nil {
		log.Errorf(" update user 表失败：%v", err)
		return err
	}
	return nil
}

type CheckPassword struct {
	Address  string `json:"address" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (c *CheckPassword) CheckPwd() (error, bool) {
	// 查看地址是否在数据库中存在
	user, err := new(models.User).GetUserByAddress(c.Address)
	if err != nil {
		// 表示不存在或者查询失败则返回error
		log.Errorf("通过address 查询user error: %v", err)
		return err, false
	}
	// 存在则查看password对不对
	if !user.CheckPassword(c.Password) {
		// 密码验证 不通过
		return nil, false
	} else {
		return nil, true
	}
}

type TokenTxsReceiveRecord struct {
	Address string `json:"address" binding:"required"`
	Page    int    `json:"page" binding:"required"`
	Limit   int    `json:"limit"`
}

type ReceiveTxRecord struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount string `json:"amount"`
	TimeAt int64  `json:"time_at"`
}

func (p *TokenTxsReceiveRecord) GetEthTokenTxRecord(tokenAddress string, decimal int) ([]ReceiveTxRecord, error) {

	address := p.Address
	if p.Limit == 0 {
		p.Limit = 5
	}
	// 拉取最近的Limit条 token记录
	txs, err := utils.GetAddressTokenTransfers(tokenAddress, address, p.Page, p.Limit)
	if err != nil {
		log.Errorf("获取收款记录失败：err: %v, address: %s", err, address)
		return nil, err
	}
	// 筛选出接收的record
	var receiveTxs = make([]ReceiveTxRecord, 0, 5)
	for _, tx := range txs {
		if strings.ToLower(tx.To) == strings.ToLower(address) {
			receiveTxs = append(receiveTxs, ReceiveTxRecord{
				From:   tx.From,
				To:     address,
				Amount: utils.UnitConversion(tx.Value.Int().String(), decimal, 6),
				TimeAt: tx.TimeStamp.Time().Unix(),
			})
		}
	}
	return receiveTxs, nil
}

type EthTxsRecord struct {
	Address string `json:"address" binding:"required"`
	Page    int    `json:"page" binding:"required"`
	Limit   int    `json:"limit"`
}

func (e *EthTxsRecord) GetEthTxsRecord() ([]ReceiveTxRecord, error) {
	address := e.Address
	if e.Limit == 0 {
		e.Limit = 5
	}
	// 拉取最近的Limit条pala token记录
	txs, err := utils.GetAddressEthTransfers(address, e.Page, e.Limit)
	if err != nil {
		log.Errorf("获取收款记录失败：err: %v, address: %s", err, address)
		return nil, err
	}
	// 筛选出接收的record
	var receiveTxs = make([]ReceiveTxRecord, 0, 5)
	for _, tx := range txs {
		if strings.ToLower(tx.To) == strings.ToLower(address) {
			receiveTxs = append(receiveTxs, ReceiveTxRecord{
				From:   tx.From,
				To:     address,
				Amount: utils.UnitConversion(tx.Value.Int().String(), 18, 6),
				TimeAt: tx.TimeStamp.Time().Unix(),
			})
		}
	}
	return receiveTxs, nil
}
