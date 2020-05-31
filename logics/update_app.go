package logics

import (
	"encoding/json"
	"github.com/jinzhu/gorm"
	"github.com/zyjblockchain/sandy_log/log"
	"github.com/zyjblockchain/tt_tac/models"
	"strconv"
	"strings"
)

const versionKey = "appversionkey"

var AppVersionInfo = &SetVersionInfo{
	Version: "0.0.1",
	WgtUrl:  "",
	PkgUrl:  "",
}

// 初始化读取数据库存储的最新版本信息到内存中
func InitAppVersionInfo() error {
	val, err := models.GetKv(versionKey)
	if err == gorm.ErrRecordNotFound {
		return nil
	} else if err != nil {
		return err
	} else {
		ss := &SetVersionInfo{}
		err = json.Unmarshal(val, ss)
		if err != nil {
			return err
		}
		// 更新内存中维护的版本信息
		AppVersionInfo = ss
		return nil
	}
}

type RespUpdate struct {
	Update  bool   `json:"update"`
	Version string `json:"version"` // 版本
	WgtUrl  string `json:"wgt_url"` // 小版本更新
	PkgUrl  string `json:"pkg_url"` // 大版本整包更新
}

// APP 版本更新后台接口
func CheckUpdate(version string) RespUpdate {
	if AppVersionInfo.Version == version {
		return RespUpdate{
			Update:  false,
			Version: "",
			WgtUrl:  "",
			PkgUrl:  "",
		}
	}

	latestVersionArr := strings.Split(AppVersionInfo.Version, ".")
	clientVersionArr := strings.Split(version, ".")
	// todo 应该返回error
	if len(clientVersionArr) != 3 {
		return RespUpdate{
			Update:  false,
			Version: "",
			WgtUrl:  "",
			PkgUrl:  "",
		}
	}

	lArr0, _ := strconv.Atoi(latestVersionArr[0])
	cArr0, _ := strconv.Atoi(clientVersionArr[0])
	if lArr0 > cArr0 {
		// 需要大版本更新
		return RespUpdate{
			Update:  true,
			Version: AppVersionInfo.Version,
			WgtUrl:  "",
			PkgUrl:  AppVersionInfo.PkgUrl,
		}
	}
	if lArr0 == cArr0 {
		// 判断是否需要小版本更新
		if latestVersionArr[1] != clientVersionArr[1] || latestVersionArr[2] != clientVersionArr[2] {
			// 小版本更新
			return RespUpdate{
				Update:  true,
				Version: AppVersionInfo.Version,
				WgtUrl:  AppVersionInfo.WgtUrl,
				PkgUrl:  "",
			}
		}
	}
	// 不需要更新
	return RespUpdate{
		Update:  false,
		Version: "",
		WgtUrl:  "",
		PkgUrl:  "",
	}
}

type SetVersionInfo struct {
	Version string `json:"version" binding:"required"`
	WgtUrl  string `json:"wgt_url"`
	PkgUrl  string `json:"pkg_url"`
}

// SetAppVersionInfo
func (s *SetVersionInfo) SetAppVersionInfo() error {
	val, err := models.GetKv(versionKey)
	if err == gorm.ErrRecordNotFound {
		// 第一次设置更新版本则直接设置进去
		newVal, err := json.Marshal(s)
		if err != nil {
			log.Errorf("marshal SetVersionInfo error: %v", err)
			return err
		}
		err = models.SetKv(versionKey, newVal)
		if err != nil {
			return err
		}
		// 更新内存中的版本info
		AppVersionInfo = s
		return nil
	} else if err != nil {
		return err
	} else {
		// 更新
		oldInfo := SetVersionInfo{}
		err := json.Unmarshal(val, &oldInfo)
		if err != nil {
			log.Errorf("json unmarshal error: %v", err)
			return err
		}
		if s.WgtUrl == "" {
			s.WgtUrl = oldInfo.WgtUrl
		}
		if s.PkgUrl == "" {
			s.PkgUrl = oldInfo.PkgUrl
		}
		// 更新
		newVal, err := json.Marshal(s)
		if err != nil {
			log.Errorf("marshal SetVersionInfo error: %v", err)
			return err
		}
		err = models.Update(versionKey, newVal)
		if err != nil {
			return err
		}
		// 更新内存中的版本info
		AppVersionInfo = s
		return nil
	}
}
