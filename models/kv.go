package models

import "github.com/jinzhu/gorm"

type Kv struct {
	gorm.Model
	Key string
	Val string
}

func SetKv(k, v string) error {
	kv := Kv{
		Key: k,
		Val: v,
	}
	return DB.Create(&kv).Error
}

func GetKv(k string) (string, error) {
	var kv Kv
	err := DB.Where("key = ?", k).First(&kv).Error
	if err != nil {
		return "", err
	}
	return kv.Val, nil
}
