package model

import (
	"github.com/llr104/slgserver/internal/protocol"
)

type RoleRes struct {
	Id       int `bson:"id" json:"id"`
	ServerId int `bson:"server_id" json:"serverId"`
	RId      int `bson:"rid" json:"rid"`
	Wood     int `bson:"wood" json:"wood"`
	Iron     int `bson:"iron" json:"iron"`
	Stone    int `bson:"stone" json:"stone"`
	Grain    int `bson:"grain" json:"grain"`
	Gold     int `bson:"gold" json:"gold"`
	Decree   int `bson:"decree" json:"decree"`
}

func (RoleRes) CollectionName() string {
	return "role_resources"
}

/* 推送同步 begin */
func (this *RoleRes) IsCellView() bool             { return false }
func (this *RoleRes) IsCanView(rid, x, y int) bool { return false }
func (this *RoleRes) BelongToRId() []int           { return []int{this.RId} }
func (this *RoleRes) PushMsgName() string          { return "roleRes.push" }
func (this *RoleRes) Position() (int, int)         { return -1, -1 }
func (this *RoleRes) TPosition() (int, int)        { return -1, -1 }

func (this *RoleRes) ToProto() interface{} {
	p := protocol.RoleRes{}
	p.Gold = this.Gold
	p.Grain = this.Grain
	p.Stone = this.Stone
	p.Iron = this.Iron
	p.Wood = this.Wood
	p.Decree = this.Decree

	if GetYield != nil {
		y := GetYield(this.RId)
		p.GoldYield = y.Gold
		p.GrainYield = y.Grain
		p.StoneYield = y.Stone
		p.IronYield = y.Iron
		p.WoodYield = y.Wood
	}
	if GetDepotCapacity != nil {
		p.DepotCapacity = GetDepotCapacity(this.RId)
	}
	return p
}

func (this *RoleRes) Push() { pushIfHooked(this) }

/* 推送同步 end */

func (this *RoleRes) SyncExecute() {
	syncIfHooked(this)
	this.Push()
}
