package model

import (
	"time"

	"github.com/llr104/slgserver/internal/protocol"
)

const (
	UnionDismiss = 0 // 解散
	UnionRunning = 1 // 运行中
)

const (
	UnionOpCreate    = 0 // 创建
	UnionOpDismiss   = 1 // 解散
	UnionOpJoin      = 2 // 加入
	UnionOpExit      = 3 // 退出
	UnionOpKick      = 4 // 踢出
	UnionOpAppoint   = 5 // 任命
	UnionOpAbdicate  = 6 // 禅让
	UnionOpModNotice = 7 // 修改公告
)

type Coalition struct {
	Id           int       `bson:"id" json:"id"`
	ServerId     int       `bson:"server_id" json:"serverId"`
	Name         string    `bson:"name" json:"name"`
	MemberArray  []int     `bson:"members" json:"members"`
	CreateId     int       `bson:"create_id" json:"createId"`
	Chairman     int       `bson:"chairman" json:"chairman"`
	ViceChairman int       `bson:"vice_chairman" json:"viceChairman"`
	Notice       string    `bson:"notice" json:"notice"`
	State        int8      `bson:"state" json:"state"`
	Ctime        time.Time `bson:"ctime" json:"ctime"`
}

func (Coalition) CollectionName() string {
	return "coalitions"
}

func (this *Coalition) Cnt() int { return len(this.MemberArray) }

func (this *Coalition) ToProto() interface{} {
	p := protocol.Union{}
	p.Id = this.Id
	p.Name = this.Name
	p.Notice = this.Notice
	p.Cnt = this.Cnt()
	return p
}

func (this *Coalition) SyncExecute() {
	syncIfHooked(this)
}

type CoalitionApply struct {
	Id       int       `bson:"id" json:"id"`
	ServerId int       `bson:"server_id" json:"serverId"`
	UnionId  int       `bson:"union_id" json:"unionId"`
	RId      int       `bson:"rid" json:"rid"`
	State    int8      `bson:"state" json:"state"`
	Ctime    time.Time `bson:"ctime" json:"ctime"`
}

func (CoalitionApply) CollectionName() string {
	return "coalition_applies"
}

/* 推送同步 begin */
func (this *CoalitionApply) IsCellView() bool             { return false }
func (this *CoalitionApply) IsCanView(rid, x, y int) bool { return false }

func (this *CoalitionApply) BelongToRId() []int {
	if GetMainMembers != nil {
		r := GetMainMembers(this.UnionId)
		return append(r, this.RId)
	}
	return []int{this.RId}
}

func (this *CoalitionApply) PushMsgName() string  { return "unionApply.push" }
func (this *CoalitionApply) Position() (int, int) { return -1, -1 }
func (this *CoalitionApply) TPosition() (int, int) { return -1, -1 }

func (this *CoalitionApply) ToProto() interface{} {
	p := protocol.ApplyItem{}
	p.RId = this.RId
	p.Id = this.Id
	if GetRoleNickName != nil {
		p.NickName = GetRoleNickName(this.RId)
	}
	return p
}

func (this *CoalitionApply) Push() { pushIfHooked(this) }

/* 推送同步 end */

func (this *CoalitionApply) SyncExecute() { this.Push() }

type CoalitionLog struct {
	Id       int       `bson:"id" json:"id"`
	ServerId int       `bson:"server_id" json:"serverId"`
	UnionId  int       `bson:"union_id" json:"unionId"`
	OPRId    int       `bson:"op_rid" json:"opRId"`
	TargetId int       `bson:"target_id" json:"targetId"`
	State    int8      `bson:"state" json:"state"`
	Des      string    `bson:"des" json:"des"`
	Ctime    time.Time `bson:"ctime" json:"ctime"`
}

func (CoalitionLog) CollectionName() string {
	return "coalition_logs"
}

func (this *CoalitionLog) ToProto() interface{} {
	p := protocol.UnionLog{}
	p.OPRId = this.OPRId
	p.TargetId = this.TargetId
	p.Des = this.Des
	p.State = this.State
	p.Ctime = this.Ctime.UnixNano() / 1e6
	return p
}

// CoalitionLogInsertHook 用于把日志写入 mongo（slgserver 节点注册）。
var CoalitionLogInsertHook func(*CoalitionLog)

func (this *CoalitionLog) Insert() {
	if CoalitionLogInsertHook != nil {
		CoalitionLogInsertHook(this)
	}
}

// 联盟操作日志辅助构造器（与原 slgserver 同语义，落库由 hook 负责）。

func NewCreate(opNickName string, unionId int, opRId int) {
	(&CoalitionLog{
		UnionId: unionId, OPRId: opRId,
		State: UnionOpCreate,
		Des:   opNickName + " 创建了联盟",
		Ctime: time.Now(),
	}).Insert()
}

func NewDismiss(opNickName string, unionId int, opRId int) {
	(&CoalitionLog{
		UnionId: unionId, OPRId: opRId,
		State: UnionOpDismiss,
		Des:   opNickName + " 解散了联盟",
		Ctime: time.Now(),
	}).Insert()
}

func NewJoin(targetNickName string, unionId int, opRId int, targetId int) {
	(&CoalitionLog{
		UnionId: unionId, OPRId: opRId, TargetId: targetId,
		State: UnionOpJoin,
		Des:   targetNickName + " 加入了联盟",
		Ctime: time.Now(),
	}).Insert()
}

func NewExit(opNickName string, unionId int, opRId int) {
	(&CoalitionLog{
		UnionId: unionId, OPRId: opRId, TargetId: opRId,
		State: UnionOpExit,
		Des:   opNickName + " 退出了联盟",
		Ctime: time.Now(),
	}).Insert()
}

func NewKick(opNickName, targetNickName string, unionId, opRId, targetId int) {
	(&CoalitionLog{
		UnionId: unionId, OPRId: opRId, TargetId: targetId,
		State: UnionOpKick,
		Des:   opNickName + " 将 " + targetNickName + " 踢出了联盟",
		Ctime: time.Now(),
	}).Insert()
}

func unionTitle(memberType int) string {
	switch memberType {
	case protocol.UnionChairman:
		return "盟主"
	case protocol.UnionViceChairman:
		return "副盟主"
	default:
		return "普通成员"
	}
}

func NewAppoint(opNickName, targetNickName string, unionId, opRId, targetId, memberType int) {
	(&CoalitionLog{
		UnionId: unionId, OPRId: opRId, TargetId: targetId,
		State: UnionOpAppoint,
		Des:   opNickName + " 将 " + targetNickName + " 任命为 " + unionTitle(memberType),
		Ctime: time.Now(),
	}).Insert()
}

func NewAbdicate(opNickName, targetNickName string, unionId, opRId, targetId, memberType int) {
	(&CoalitionLog{
		UnionId: unionId, OPRId: opRId, TargetId: targetId,
		State: UnionOpAbdicate,
		Des:   opNickName + " 将 " + unionTitle(memberType) + " 禅让给 " + targetNickName,
		Ctime: time.Now(),
	}).Insert()
}

func NewModNotice(opNickName string, unionId int, opRId int) {
	(&CoalitionLog{
		UnionId: unionId, OPRId: opRId,
		State: UnionOpModNotice,
		Des:   opNickName + " 修改了公告",
		Ctime: time.Now(),
	}).Insert()
}
