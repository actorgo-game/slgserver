package model

import (
	"time"

	"github.com/llr104/slgserver/internal/protocol"
	"github.com/llr104/slgserver/internal/util"
)

const (
	MapBuildSysFortress = 50 // 系统要塞
	MapBuildSysCity     = 51 // 系统城市
	MapBuildFortress    = 56 // 玩家要塞
)

// MapBuildCfg 是 mgr 层向 model 层注入的建筑配置查询结果。
// 解耦 model 不直接依赖 static_conf。
type MapBuildCfg struct {
	Type     int8
	Level    int8
	Name     string
	Wood     int
	Iron     int
	Stone    int
	Grain    int
	Durable  int
	Defender int
	Time     int
}

// 由 slgserver/logic.BeforeInit 注入：根据 type+level 查询建筑配置。
var GetMapBuildCfg func(t, level int8) (MapBuildCfg, bool)

type MapRoleBuild struct {
	Id         int       `bson:"id" json:"id"`
	ServerId   int       `bson:"server_id" json:"serverId"`
	RId        int       `bson:"rid" json:"rid"`
	Type       int8      `bson:"type" json:"type"`
	Level      int8      `bson:"level" json:"level"`
	OPLevel    int8      `bson:"op_level" json:"opLevel"`
	X          int       `bson:"x" json:"x"`
	Y          int       `bson:"y" json:"y"`
	Name       string    `bson:"name" json:"name"`
	Wood       int       `bson:"-" json:"-"`
	Iron       int       `bson:"-" json:"-"`
	Stone      int       `bson:"-" json:"-"`
	Grain      int       `bson:"-" json:"-"`
	Defender   int       `bson:"-" json:"-"`
	CurDurable int       `bson:"cur_durable" json:"curDurable"`
	MaxDurable int       `bson:"max_durable" json:"maxDurable"`
	OccupyTime time.Time `bson:"occupy_time" json:"occupyTime"`
	EndTime    time.Time `bson:"end_time" json:"endTime"`
	GiveUpTime int64     `bson:"give_up_time" json:"giveUpTime"`
}

func (MapRoleBuild) CollectionName() string {
	return "builds"
}

// Init 用于按当前 Type+Level 从配置回填字段。
func (this *MapRoleBuild) Init() {
	if GetMapBuildCfg == nil {
		return
	}
	if cfg, ok := GetMapBuildCfg(this.Type, this.Level); ok {
		this.Name = cfg.Name
		this.Level = cfg.Level
		this.Type = cfg.Type
		this.Wood = cfg.Wood
		this.Iron = cfg.Iron
		this.Stone = cfg.Stone
		this.Grain = cfg.Grain
		this.MaxDurable = cfg.Durable
		this.CurDurable = cfg.Durable
		this.Defender = cfg.Defender
	}
}

// Reset 把建筑还原成系统资源点（用于放弃/被销毁后回收）。
func (this *MapRoleBuild) Reset() {
	if MapResTypeLevel != nil && GetMapBuildCfg != nil {
		ok, t, level := MapResTypeLevel(this.X, this.Y)
		if ok {
			if cfg, found := GetMapBuildCfg(t, level); found {
				this.Name = cfg.Name
				this.Level = cfg.Level
				this.Type = cfg.Type
				this.Wood = cfg.Wood
				this.Iron = cfg.Iron
				this.Stone = cfg.Stone
				this.Grain = cfg.Grain
				this.MaxDurable = cfg.Durable
				this.Defender = cfg.Defender
			}
		}
	}

	this.GiveUpTime = 0
	this.RId = 0
	this.EndTime = time.Time{}
	this.OPLevel = this.Level
	this.CurDurable = util.MinInt(this.MaxDurable, this.CurDurable)
}

func (this *MapRoleBuild) ConvertToRes() {
	rid := this.RId
	giveUp := this.GiveUpTime
	this.Reset()
	this.RId = rid
	this.GiveUpTime = giveUp
}

func (this *MapRoleBuild) IsInGiveUp() bool { return this.GiveUpTime != 0 }

func (this *MapRoleBuild) IsWarFree() bool {
	return time.Now().Unix()-this.OccupyTime.Unix() < WarFreeSeconds
}

func (this *MapRoleBuild) IsResBuild() bool {
	return this.Grain > 0 || this.Stone > 0 || this.Iron > 0 || this.Wood > 0
}

func (this *MapRoleBuild) IsHaveModifyLVAuth() bool { return this.Type == MapBuildFortress }

func (this *MapRoleBuild) IsBusy() bool { return this.Level != this.OPLevel }

func (this *MapRoleBuild) IsRoleFortress() bool { return this.Type == MapBuildFortress }

func (this *MapRoleBuild) IsSysFortress() bool { return this.Type == MapBuildSysFortress }

func (this *MapRoleBuild) IsSysCity() bool { return this.Type == MapBuildSysCity }

func (this *MapRoleBuild) CellRadius() int {
	if this.IsSysCity() {
		switch {
		case this.Level >= 8:
			return 3
		case this.Level >= 5:
			return 2
		default:
			return 1
		}
	}
	return 0
}

func (this *MapRoleBuild) IsHasTransferAuth() bool {
	return this.Type == MapBuildFortress || this.Type == MapBuildSysFortress
}

func (this *MapRoleBuild) BuildOrUp(cfg MapBuildCfg) {
	this.Type = cfg.Type
	this.Level = cfg.Level - 1
	this.Name = cfg.Name
	this.OPLevel = cfg.Level
	this.GiveUpTime = 0

	this.Wood = 0
	this.Iron = 0
	this.Stone = 0
	this.Grain = 0
	this.EndTime = time.Now().Add(time.Duration(cfg.Time) * time.Second)
}

func (this *MapRoleBuild) DelBuild(cfg MapBuildCfg) {
	this.OPLevel = 0
	this.EndTime = time.Now().Add(time.Duration(cfg.Time) * time.Second)
}

/* 推送同步 begin */
func (this *MapRoleBuild) IsCellView() bool             { return true }
func (this *MapRoleBuild) IsCanView(rid, x, y int) bool { return true }
func (this *MapRoleBuild) BelongToRId() []int           { return []int{this.RId} }
func (this *MapRoleBuild) PushMsgName() string          { return "roleBuild.push" }
func (this *MapRoleBuild) Position() (int, int)         { return this.X, this.Y }
func (this *MapRoleBuild) TPosition() (int, int)        { return -1, -1 }

func (this *MapRoleBuild) ToProto() interface{} {
	p := protocol.MapRoleBuild{}
	if GetRoleNickName != nil {
		p.RNick = GetRoleNickName(this.RId)
	}
	if GetUnionId != nil {
		p.UnionId = GetUnionId(this.RId)
	}
	if GetUnionName != nil {
		p.UnionName = GetUnionName(p.UnionId)
	}
	if GetParentId != nil {
		p.ParentId = GetParentId(this.RId)
	}
	p.X = this.X
	p.Y = this.Y
	p.Type = this.Type
	p.RId = this.RId
	p.Name = this.Name
	p.OccupyTime = this.OccupyTime.UnixNano() / 1e6
	p.GiveUpTime = this.GiveUpTime * 1000
	p.EndTime = this.EndTime.UnixNano() / 1e6

	if !this.EndTime.IsZero() && this.IsHasTransferAuth() {
		if !time.Now().Before(this.EndTime) {
			if this.OPLevel == 0 {
				this.ConvertToRes()
			} else {
				this.Level = this.OPLevel
				this.EndTime = time.Time{}
				if GetMapBuildCfg != nil {
					if cfg, ok := GetMapBuildCfg(this.Type, this.Level); ok {
						this.MaxDurable = cfg.Durable
						this.CurDurable = util.MinInt(this.MaxDurable, this.CurDurable)
						this.Defender = cfg.Defender
					}
				}
			}
		}
	}

	p.CurDurable = this.CurDurable
	p.MaxDurable = this.MaxDurable
	p.Defender = this.Defender
	p.Level = this.Level
	p.OPLevel = this.OPLevel
	return p
}

func (this *MapRoleBuild) Push() { pushIfHooked(this) }

/* 推送同步 end */

func (this *MapRoleBuild) SyncExecute() {
	syncIfHooked(this)
	this.Push()
}
