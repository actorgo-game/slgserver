package model

import (
	"github.com/llr104/slgserver/internal/protocol"
)

// SkillCfg 由 mgr 注入：根据 cfgId 查询技能限制。
type SkillCfg struct {
	Limit int
	Arms  []int
}

var GetSkillCfg func(cfgId int) (SkillCfg, bool)

type Skill struct {
	Id       int   `bson:"id" json:"id"`
	ServerId int   `bson:"server_id" json:"serverId"`
	RId      int   `bson:"rid" json:"rid"`
	CfgId    int   `bson:"cfgId" json:"cfgId"`
	Generals []int `bson:"belong_generals" json:"generals"`
}

func (Skill) CollectionName() string {
	return "skills"
}

func NewSkill(rid int, cfgId int) *Skill {
	return &Skill{
		CfgId:    cfgId,
		RId:      rid,
		Generals: []int{},
	}
}

func (this *Skill) Limit() int {
	if GetSkillCfg == nil {
		return 0
	}
	cfg, _ := GetSkillCfg(this.CfgId)
	return cfg.Limit
}

func (this *Skill) IsInLimit() bool { return len(this.Generals) < this.Limit() }

func (this *Skill) ArmyIsIn(armId int) bool {
	if GetSkillCfg == nil {
		return false
	}
	cfg, _ := GetSkillCfg(this.CfgId)
	for _, arm := range cfg.Arms {
		if arm == armId {
			return true
		}
	}
	return false
}

func (this *Skill) UpSkill(gId int) {
	this.Generals = append(this.Generals, gId)
}

func (this *Skill) DownSkill(gId int) {
	gs := make([]int, 0, len(this.Generals))
	for _, g := range this.Generals {
		if g != gId {
			gs = append(gs, g)
		}
	}
	this.Generals = gs
}

/* 推送同步 begin */
func (this *Skill) IsCellView() bool             { return false }
func (this *Skill) IsCanView(rid, x, y int) bool { return false }
func (this *Skill) BelongToRId() []int           { return []int{this.RId} }
func (this *Skill) PushMsgName() string          { return "skill.push" }
func (this *Skill) Position() (int, int)         { return -1, -1 }
func (this *Skill) TPosition() (int, int)        { return -1, -1 }

func (this *Skill) ToProto() interface{} {
	p := protocol.Skill{}
	p.Id = this.Id
	p.CfgId = this.CfgId
	p.Generals = this.Generals
	return p
}

func (this *Skill) Push() { pushIfHooked(this) }

/* 推送同步 end */

func (this *Skill) SyncExecute() {
	syncIfHooked(this)
	this.Push()
}
