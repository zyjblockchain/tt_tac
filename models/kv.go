package models

import (
	"github.com/jinzhu/gorm"
)

type Kv struct {
	gorm.Model
	TKey string
	TVal []byte `gorm:"type:text(0)"`
}

func (Kv) TableName() string {
	return "kv"
}

func SetKv(k string, v []byte) error {
	kv := Kv{
		TKey: k,
		TVal: v,
	}
	return DB.Create(&kv).Error
}

func GetKv(k string) ([]byte, error) {
	var vv Kv
	err := DB.Where(&Kv{TKey: k}).First(&vv).Error
	if err != nil {
		return nil, err
	}
	return vv.TVal, nil
}

func Update(key string, newVal []byte) error {
	return DB.Model(Kv{}).Where("t_key = ?", key).Update("t_val", newVal).Error
}
