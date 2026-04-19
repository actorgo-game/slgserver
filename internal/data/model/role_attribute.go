package model

import (
	"time"

	"github.com/llr104/slgserver/internal/protocol"
)

type PosTag struct {
	X    int    `bson:"x" json:"x"`
	Y    int    `bson:"y" json:"y"`
	Name string `bson:"name" json:"name"`
}

type RoleAttribute struct {
	Id              int       `bson:"id" json:"id"`
	ServerId        int       `bson:"server_id" json:"serverId"`
	RId             int       `bson:"rid" json:"rid"`
	UnionId         int       `bson:"-" json:"unionId"` // 不入库，启动时由 mgr 计算
	ParentId        int       `bson:"parent_id" json:"parentId"`
	CollectTimes    int8      `bson:"collect_times" json:"collectTimes"`
	LastCollectTime time.Time `bson:"last_collect_time" json:"lastCollectTime"`
	PosTagArray     []PosTag  `bson:"pos_tags" json:"posTags"`
}

func (RoleAttribute) CollectionName() string {
	return "role_attributes"
}

func (this *RoleAttribute) RemovePosTag(x, y int) {
	tags := make([]PosTag, 0)
	for _, tag := range this.PosTagArray {
		if tag.X != x || tag.Y != y {
			tags = append(tags, tag)
		}
	}
	this.PosTagArray = tags
}

func (this *RoleAttribute) AddPosTag(x, y int, name string) {
	for _, tag := range this.PosTagArray {
		if tag.X == x && tag.Y == y {
			return
		}
	}
	this.PosTagArray = append(this.PosTagArray, PosTag{X: x, Y: y, Name: name})
}

func (this *RoleAttribute) ProtoPosTags() []protocol.PosTag {
	r := make([]protocol.PosTag, 0, len(this.PosTagArray))
	for _, t := range this.PosTagArray {
		r = append(r, protocol.PosTag{X: t.X, Y: t.Y, Name: t.Name})
	}
	return r
}

/* 推送同步 begin */
func (this *RoleAttribute) IsCellView() bool             { return false }
func (this *RoleAttribute) IsCanView(rid, x, y int) bool { return false }
func (this *RoleAttribute) BelongToRId() []int           { return []int{this.RId} }
func (this *RoleAttribute) PushMsgName() string          { return "roleAttr.push" }
func (this *RoleAttribute) Position() (int, int)         { return -1, -1 }
func (this *RoleAttribute) TPosition() (int, int)        { return -1, -1 }
func (this *RoleAttribute) ToProto() interface{}         { return nil }
func (this *RoleAttribute) Push()                        { pushIfHooked(this) }

/* 推送同步 end */

func (this *RoleAttribute) SyncExecute() {
	syncIfHooked(this)
}
