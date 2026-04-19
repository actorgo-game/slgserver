package model

import (
	"time"

	"github.com/llr104/slgserver/internal/protocol"
	"github.com/llr104/slgserver/internal/util"
)

const (
	ArmyCmdIdle        = 0 // 空闲
	ArmyCmdAttack      = 1 // 攻击
	ArmyCmdDefend      = 2 // 驻守
	ArmyCmdReclamation = 3 // 屯垦
	ArmyCmdBack        = 4 // 撤退
	ArmyCmdConscript   = 5 // 征兵
	ArmyCmdTransfer    = 6 // 调动
)

const (
	ArmyStop    = 0
	ArmyRunning = 1
)

const ArmyGCnt = 3

// MapBound 由 slgserver 启动后注入：限制军队 Position() 的边界，避免 model
// 直接依赖 global 包导致循环。
var MapBound = struct{ W, H int }{200, 200}

type Army struct {
	Id                 int                  `bson:"id" json:"id"`
	ServerId           int                  `bson:"server_id" json:"serverId"`
	RId                int                  `bson:"rid" json:"rid"`
	CityId             int                  `bson:"cityId" json:"cityId"`
	Order              int8                 `bson:"order" json:"order"`
	GeneralArray       [ArmyGCnt]int        `bson:"generals" json:"generals"`
	SoldierArray       [ArmyGCnt]int        `bson:"soldiers" json:"soldiers"`
	ConscriptTimeArray [ArmyGCnt]int64      `bson:"conscript_times" json:"conscriptTimes"`
	ConscriptCntArray  [ArmyGCnt]int        `bson:"conscript_cnts" json:"conscriptCnts"`
	Cmd                int8                 `bson:"cmd" json:"cmd"`
	FromX              int                  `bson:"from_x" json:"fromX"`
	FromY              int                  `bson:"from_y" json:"fromY"`
	ToX                int                  `bson:"to_x" json:"toX"`
	ToY                int                  `bson:"to_y" json:"toY"`
	Start              time.Time            `bson:"start" json:"-"`
	End                time.Time            `bson:"end" json:"-"`
	State              int8                 `bson:"-" json:"state"`
	Gens               [ArmyGCnt]*General   `bson:"-" json:"-"`
	CellX              int                  `bson:"-" json:"-"`
	CellY              int                  `bson:"-" json:"-"`
}

func (Army) CollectionName() string {
	return "armies"
}

func (this *Army) IsCanOutWar() bool {
	return this.Gens[0] != nil && this.Cmd == ArmyCmdIdle
}

func (this *Army) GetCamp() int8 {
	var camp int8 = 0
	for _, g := range this.Gens {
		if g == nil {
			return 0
		}
		if camp == 0 {
			camp = g.GetCamp()
		} else if camp != g.GetCamp() {
			return 0
		}
	}
	return camp
}

// CheckConscript 服务器不做定时任务，用到时再检查征兵是否完成
func (this *Army) CheckConscript() {
	if this.Cmd != ArmyCmdConscript {
		return
	}
	curTime := time.Now().Unix()
	finish := true
	for i, endTime := range this.ConscriptTimeArray {
		if endTime > 0 {
			if endTime <= curTime {
				this.SoldierArray[i] += this.ConscriptCntArray[i]
				this.ConscriptCntArray[i] = 0
				this.ConscriptTimeArray[i] = 0
			} else {
				finish = false
			}
		}
	}
	if finish {
		this.Cmd = ArmyCmdIdle
	}
}

func (this *Army) PositionCanModify(position int) bool {
	if position >= ArmyGCnt || position < 0 {
		return false
	}
	switch this.Cmd {
	case ArmyCmdIdle:
		return true
	case ArmyCmdConscript:
		return this.ConscriptTimeArray[position] == 0
	default:
		return false
	}
}

func (this *Army) ClearConscript() {
	if this.Cmd == ArmyCmdConscript {
		for i := range this.ConscriptTimeArray {
			this.ConscriptCntArray[i] = 0
			this.ConscriptTimeArray[i] = 0
		}
		this.Cmd = ArmyCmdIdle
	}
}

func (this *Army) IsIdle() bool { return this.Cmd == ArmyCmdIdle }

/* 推送同步 begin */
func (this *Army) IsCellView() bool { return true }

func (this *Army) IsCanView(rid, x, y int) bool {
	if ArmyIsInView != nil {
		return ArmyIsInView(rid, x, y)
	}
	return false
}

func (this *Army) BelongToRId() []int  { return []int{this.RId} }
func (this *Army) PushMsgName() string { return "army.push" }

func (this *Army) Position() (int, int) {
	diffTime := this.End.Unix() - this.Start.Unix()
	if diffTime <= 0 {
		return util.MinInt(util.MaxInt(this.FromX, 0), MapBound.W),
			util.MinInt(util.MaxInt(this.FromY, 0), MapBound.H)
	}
	passTime := time.Now().Unix() - this.Start.Unix()
	rate := float32(passTime) / float32(diffTime)
	x, y := 0, 0
	if this.Cmd == ArmyCmdBack {
		diffX := this.FromX - this.ToX
		diffY := this.FromY - this.ToY
		x = int(rate*float32(diffX)) + this.ToX
		y = int(rate*float32(diffY)) + this.ToY
	} else {
		diffX := this.ToX - this.FromX
		diffY := this.ToY - this.FromY
		x = int(rate*float32(diffX)) + this.FromX
		y = int(rate*float32(diffY)) + this.FromY
	}
	x = util.MinInt(util.MaxInt(x, 0), MapBound.W)
	y = util.MinInt(util.MaxInt(y, 0), MapBound.H)
	return x, y
}

func (this *Army) TPosition() (int, int) { return this.ToX, this.ToY }

func (this *Army) ToProto() interface{} {
	p := protocol.Army{}
	p.CityId = this.CityId
	p.Id = this.Id
	if GetUnionId != nil {
		p.UnionId = GetUnionId(this.RId)
	}
	p.Order = this.Order
	p.Generals = this.GeneralArray
	p.Soldiers = this.SoldierArray
	p.ConTimes = this.ConscriptTimeArray
	p.ConCnts = this.ConscriptCntArray
	p.Cmd = this.Cmd
	p.State = this.State
	p.FromX = this.FromX
	p.FromY = this.FromY
	p.ToX = this.ToX
	p.ToY = this.ToY
	p.Start = this.Start.Unix()
	p.End = this.End.Unix()
	return p
}

func (this *Army) Push() { pushIfHooked(this) }

/* 推送同步 end */

func (this *Army) SyncExecute() {
	syncIfHooked(this)
	this.Push()
	this.CellX, this.CellY = this.Position()
}

func (this *Army) CheckSyncCell() {
	x, y := this.Position()
	if x != this.CellX || y != this.CellY {
		this.SyncExecute()
	}
}
