package models

import (
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

const (
	// passWordCost 密码加密难度
	passWordCost = 12
)

type User struct {
	gorm.Model
	Address        string // 用户地址
	PrivateCrypted string // 通过aes加密过后的私钥
	PasswordDigest string // 用户的加密之后的支付密码
}

func (User) TableName() string {
	return "user"
}

// SetPassword 设置或者修改user在数据库中保存的密码
func (user *User) SetPassword(rawPassword string) error {
	// 对原始密码进行加密
	bytes, err := bcrypt.GenerateFromPassword([]byte(rawPassword), passWordCost)
	if err != nil {
		return err
	}
	user.PasswordDigest = string(bytes)
	return nil
}

// CheckPassword 校验用户密码
func (user *User) CheckPassword(rawPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordDigest), []byte(rawPassword))
	return err == nil
}

// GetUserByAddress
func (u *User) GetUserByAddress(Address string) (*User, error) {
	var user User
	result := DB.Where("address = ?", Address).First(&user)
	return &user, result.Error
}

// AddUser 增加user并返回结果
func (u *User) AddUser(newUser *User) (*User, error) {
	err := DB.Create(newUser).Error
	return newUser, err
}

func (u *User) Update() error {
	return DB.Model(u).Updates(u).Error
}
