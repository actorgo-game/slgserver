package model

import "time"

// FacilityCostTime 由 mgr 注入：根据 (Type, NextLevel) 查询升级耗时（秒）。
var FacilityCostTime func(t int8, nextLevel int8) int

type Facility struct {
	Name         string `bson:"name" json:"name"`
	PrivateLevel int8   `bson:"level" json:"level"`
	Type         int8   `bson:"type" json:"type"`
	UpTime       int64  `bson:"up_time" json:"upTime"` // 升级开始时间戳，0 表示已完成
}

// GetLevel 升级是被动触发产生的：每次读取 Level 时检查是否已升级完成。
func (this *Facility) GetLevel() int8 {
	if this.UpTime > 0 && FacilityCostTime != nil {
		cost := FacilityCostTime(this.Type, this.PrivateLevel+1)
		if time.Now().Unix() >= this.UpTime+int64(cost) {
			this.PrivateLevel++
			this.UpTime = 0
		}
	}
	return this.PrivateLevel
}

func (this *Facility) CanLV() bool {
	this.GetLevel()
	return this.UpTime == 0
}

type CityFacility struct {
	Id         int        `bson:"id" json:"id"`
	ServerId   int        `bson:"server_id" json:"serverId"`
	RId        int        `bson:"rid" json:"rid"`
	CityId     int        `bson:"cityId" json:"cityId"`
	Facilities []Facility `bson:"facilities" json:"facilities"`
}

func (CityFacility) CollectionName() string {
	return "city_facilities"
}

// Facility 返回当前城市设施列表的副本，与原 slgserver Facility() 同语义。
func (this *CityFacility) Facility() []Facility {
	return this.Facilities
}

func (this *CityFacility) SyncExecute() {
	syncIfHooked(this)
}
