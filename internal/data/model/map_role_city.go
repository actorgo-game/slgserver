package model

import (
	"sync"
	"time"

	"github.com/llr104/slgserver/internal/protocol"
	"github.com/llr104/slgserver/internal/util"
)

type MapRoleCity struct {
	mutex      sync.Mutex `bson:"-" json:"-"`
	CityId     int        `bson:"cityId" json:"cityId"`
	ServerId   int        `bson:"server_id" json:"serverId"`
	RId        int        `bson:"rid" json:"rid"`
	Name       string     `bson:"name" json:"name"`
	X          int        `bson:"x" json:"x"`
	Y          int        `bson:"y" json:"y"`
	IsMain     int8       `bson:"is_main" json:"isMain"`
	CurDurable int        `bson:"cur_durable" json:"curDurable"`
	MaxDurable int        `bson:"max_durable" json:"maxDurable"`
	CreatedAt  time.Time  `bson:"created_at" json:"createdAt"`
	OccupyTime time.Time  `bson:"occupy_time" json:"occupyTime"`
}

func (MapRoleCity) CollectionName() string {
	return "cities"
}

// 战斗免战时间常量由 static_conf.Basic.Build.WarFree 决定，model 不直接引用
// static_conf 以避免循环依赖；mgr 层在调用前可先校验，或通过 IsWarFree 函数指针。
var WarFreeSeconds int64 = 0

func (this *MapRoleCity) IsWarFree() bool {
	return time.Now().Unix()-this.OccupyTime.Unix() < WarFreeSeconds
}

func (this *MapRoleCity) DurableChange(change int) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	t := this.CurDurable + change
	if t < 0 {
		this.CurDurable = 0
	} else {
		max := this.MaxDurable
		if GetMaxDurable != nil {
			if v := GetMaxDurable(this.CityId); v > 0 {
				max = v
			}
		}
		this.CurDurable = util.MinInt(max, t)
	}
}

func (this *MapRoleCity) Level() int8 {
	if GetCityLv != nil {
		return GetCityLv(this.CityId)
	}
	return 0
}

func (this *MapRoleCity) CellRadius() int { return 1 }

/* 推送同步 begin */
func (this *MapRoleCity) IsCellView() bool             { return true }
func (this *MapRoleCity) IsCanView(rid, x, y int) bool { return true }
func (this *MapRoleCity) BelongToRId() []int           { return []int{this.RId} }
func (this *MapRoleCity) PushMsgName() string          { return "roleCity.push" }
func (this *MapRoleCity) Position() (int, int)         { return this.X, this.Y }
func (this *MapRoleCity) TPosition() (int, int)        { return -1, -1 }

func (this *MapRoleCity) ToProto() interface{} {
	p := protocol.MapRoleCity{}
	p.X = this.X
	p.Y = this.Y
	p.CityId = this.CityId
	if GetUnionId != nil {
		p.UnionId = GetUnionId(this.RId)
	}
	if GetUnionName != nil {
		p.UnionName = GetUnionName(p.UnionId)
	}
	if GetParentId != nil {
		p.ParentId = GetParentId(this.RId)
	}
	if GetMaxDurable != nil {
		p.MaxDurable = GetMaxDurable(this.RId)
	} else {
		p.MaxDurable = this.MaxDurable
	}
	p.CurDurable = this.CurDurable
	p.Level = this.Level()
	p.RId = this.RId
	p.Name = this.Name
	p.IsMain = this.IsMain == 1
	p.OccupyTime = this.OccupyTime.UnixNano() / 1e6
	return p
}

func (this *MapRoleCity) Push() { pushIfHooked(this) }

/* 推送同步 end */

func (this *MapRoleCity) SyncExecute() {
	syncIfHooked(this)
	this.Push()
}
