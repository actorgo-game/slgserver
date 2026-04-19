package model

import (
	"errors"
	"time"

	"github.com/llr104/slgserver/internal/protocol"
)

const (
	GeneralNormal      = 0 // 正常
	GeneralComposeStar = 1 // 星级合成
	GeneralConvert     = 2 // 转换
)

const SkillLimit = 3

// GeneralCfg 由 mgr 在启动时注入：通过 cfgId 查询配置（属性、阵营、成长）。
type GeneralCfg struct {
	Force        int
	ForceGrow    int
	Strategy     int
	StrategyGrow int
	Defense      int
	DefenseGrow  int
	Speed        int
	SpeedGrow    int
	Destroy      int
	DestroyGrow  int
	Camp         int8
	Star         int8
	Arms         []int
}

var GetGeneralCfg func(cfgId int) (GeneralCfg, bool)

type GeneralSkill struct {
	Id    int `bson:"id" json:"id"`
	Lv    int `bson:"lv" json:"lv"`
	CfgId int `bson:"cfgId" json:"cfgId"`
}

type General struct {
	Id            int             `bson:"id" json:"id"`
	ServerId      int             `bson:"server_id" json:"serverId"`
	RId           int             `bson:"rid" json:"rid"`
	CfgId         int             `bson:"cfgId" json:"cfgId"`
	PhysicalPower int             `bson:"physical_power" json:"physicalPower"`
	Level         int8            `bson:"level" json:"level"`
	Exp           int             `bson:"exp" json:"exp"`
	Order         int8            `bson:"order" json:"order"`
	CityId        int             `bson:"cityId" json:"cityId"`
	CurArms       int             `bson:"arms" json:"curArms"`
	HasPrPoint    int             `bson:"has_pr_point" json:"hasPrPoint"`
	UsePrPoint    int             `bson:"use_pr_point" json:"usePrPoint"`
	AttackDis     int             `bson:"attack_distance" json:"attackDis"`
	ForceAdded    int             `bson:"force_added" json:"forceAdded"`
	StrategyAdded int             `bson:"strategy_added" json:"strategyAdded"`
	DefenseAdded  int             `bson:"defense_added" json:"defenseAdded"`
	SpeedAdded    int             `bson:"speed_added" json:"speedAdded"`
	DestroyAdded  int             `bson:"destroy_added" json:"destroyAdded"`
	StarLv        int8            `bson:"star_lv" json:"starLv"`
	Star          int8            `bson:"star" json:"star"`
	ParentId      int             `bson:"parentId" json:"parentId"`
	SkillsArray   []*protocol.GSkill `bson:"skills" json:"skills"`
	State         int8            `bson:"state" json:"state"`
	CreatedAt     time.Time       `bson:"created_at" json:"createdAt"`
}

func (General) CollectionName() string {
	return "generals"
}

// NewGeneralBaseInit 由 mgr 在启动时注入：返回指定 cfgId 武将的初始
// (curArms, star) 与体力上限。解耦 model 不直接依赖 static_conf。
var NewGeneralBaseInit func(cfgId int) (curArms int, star int8, physicalLimit int, ok bool)

// NewGeneral 创建一个空白武将（不写库），由 mgr 决定后续 Insert/缓存。
// 与原 model.NewGeneral 同语义，但去掉了直接写库部分。
func NewGeneral(cfgId int, rid int, level int8) (*General, bool) {
	if NewGeneralBaseInit == nil {
		return nil, false
	}
	curArms, star, plimit, ok := NewGeneralBaseInit(cfgId)
	if !ok {
		return nil, false
	}
	g := &General{
		PhysicalPower: plimit,
		RId:           rid,
		CfgId:         cfgId,
		Order:         0,
		CityId:        0,
		Level:         level,
		CreatedAt:     time.Now(),
		CurArms:       curArms,
		Star:          star,
		StarLv:        0,
		SkillsArray:   make([]*protocol.GSkill, SkillLimit),
		State:         GeneralNormal,
	}
	return g, true
}

func (this *General) GetForce() int {
	if GetGeneralCfg == nil {
		return 0
	}
	if cfg, ok := GetGeneralCfg(this.CfgId); ok {
		return cfg.Force + cfg.ForceGrow*int(this.Level) + this.ForceAdded
	}
	return 0
}

func (this *General) GetStrategy() int {
	if GetGeneralCfg == nil {
		return 0
	}
	if cfg, ok := GetGeneralCfg(this.CfgId); ok {
		return cfg.Strategy + cfg.StrategyGrow*int(this.Level) + this.StrategyAdded
	}
	return 0
}

func (this *General) GetDefense() int {
	if GetGeneralCfg == nil {
		return 0
	}
	if cfg, ok := GetGeneralCfg(this.CfgId); ok {
		return cfg.Defense + cfg.DefenseGrow*int(this.Level) + this.DefenseAdded
	}
	return 0
}

func (this *General) GetSpeed() int {
	if GetGeneralCfg == nil {
		return 0
	}
	if cfg, ok := GetGeneralCfg(this.CfgId); ok {
		return cfg.Speed + cfg.SpeedGrow*int(this.Level) + this.SpeedAdded
	}
	return 0
}

func (this *General) GetDestroy() int {
	if GetGeneralCfg == nil {
		return 0
	}
	if cfg, ok := GetGeneralCfg(this.CfgId); ok {
		return cfg.Destroy + cfg.DestroyGrow*int(this.Level) + this.DestroyAdded
	}
	return 0
}

func (this *General) GetCamp() int8 {
	if GetGeneralCfg == nil {
		return 0
	}
	if cfg, ok := GetGeneralCfg(this.CfgId); ok {
		return cfg.Camp
	}
	return 0
}

func (this *General) IsActive() bool { return this.State == GeneralNormal }

func (this *General) ensureSkillSlots() {
	if len(this.SkillsArray) < SkillLimit {
		s := make([]*protocol.GSkill, SkillLimit)
		copy(s, this.SkillsArray)
		this.SkillsArray = s
	}
}

func (this *General) UpSkill(skillId int, cfgId int, pos int) bool {
	if pos < 0 || pos >= SkillLimit {
		return false
	}
	this.ensureSkillSlots()
	for _, s := range this.SkillsArray {
		if s != nil && s.Id == skillId {
			return false
		}
	}
	s := this.SkillsArray[pos]
	if s == nil {
		this.SkillsArray[pos] = &protocol.GSkill{Id: skillId, Lv: 1, CfgId: cfgId}
		return true
	}
	if s.Id == 0 {
		s.Id = skillId
		s.CfgId = cfgId
		s.Lv = 1
		return true
	}
	return false
}

func (this *General) DownSkill(skillId int, pos int) bool {
	if pos < 0 || pos >= SkillLimit {
		return false
	}
	this.ensureSkillSlots()
	s := this.SkillsArray[pos]
	if s != nil && s.Id == skillId {
		s.Id = 0
		s.Lv = 0
		s.CfgId = 0
		return true
	}
	return false
}

func (this *General) PosSkill(pos int) (*protocol.GSkill, error) {
	this.ensureSkillSlots()
	if pos >= len(this.SkillsArray) {
		return nil, errors.New("skill index out of range")
	}
	return this.SkillsArray[pos], nil
}

/* 推送同步 begin */
func (this *General) IsCellView() bool             { return false }
func (this *General) IsCanView(rid, x, y int) bool { return false }
func (this *General) BelongToRId() []int           { return []int{this.RId} }
func (this *General) PushMsgName() string          { return "general.push" }
func (this *General) Position() (int, int)         { return -1, -1 }
func (this *General) TPosition() (int, int)        { return -1, -1 }

func (this *General) ToProto() interface{} {
	p := protocol.General{}
	p.CityId = this.CityId
	p.Order = this.Order
	p.PhysicalPower = this.PhysicalPower
	p.Id = this.Id
	p.CfgId = this.CfgId
	p.Level = this.Level
	p.Exp = this.Exp
	p.CurArms = this.CurArms
	p.HasPrPoint = this.HasPrPoint
	p.UsePrPoint = this.UsePrPoint
	p.AttackDis = this.AttackDis
	p.ForceAdded = this.ForceAdded
	p.StrategyAdded = this.StrategyAdded
	p.DefenseAdded = this.DefenseAdded
	p.SpeedAdded = this.SpeedAdded
	p.DestroyAdded = this.DestroyAdded
	p.StarLv = this.StarLv
	p.Star = this.Star
	p.State = this.State
	p.ParentId = this.ParentId
	p.Skills = this.SkillsArray
	return p
}

func (this *General) Push() { pushIfHooked(this) }

/* 推送同步 end */

func (this *General) SyncExecute() {
	syncIfHooked(this)
	this.Push()
}
