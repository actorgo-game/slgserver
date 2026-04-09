package warmgr

import (
	"math"
	"time"

	clog "github.com/actorgo-game/actorgo/logger"
	cactor "github.com/actorgo-game/actorgo/net/actor"
	"github.com/llr104/slgserver/internal/code"
	"github.com/llr104/slgserver/internal/component"
	"github.com/llr104/slgserver/internal/data/entry"
	"github.com/llr104/slgserver/internal/data/model"
)

type ArmyMovement struct {
	Army      *model.Army
	ArriveAt  int64
	OwnerPath string
}

type ActorWar struct {
	cactor.Base

	movements      map[int]*ArmyMovement
	warReportEntry *entry.WarReportEntry
	serverId       int
}

func NewActorWar() *ActorWar {
	return &ActorWar{
		movements: make(map[int]*ArmyMovement),
	}
}

func (p *ActorWar) AliasID() string {
	return "war"
}

func (p *ActorWar) OnInit() {
	p.initDB()

	p.Remote().Register("scheduleArmy", p.scheduleArmy)
	p.Remote().Register("cancelArmy", p.cancelArmy)

	p.Timer().Add(1*time.Second, p.checkArmyArrive)

	clog.Info("[ActorWar] initialized")
}

func (p *ActorWar) initDB() {
	mongoComp := p.App().Find(component.MongoComponentName)
	if mongoComp == nil {
		return
	}
	mc := mongoComp.(*component.MongoComponent)
	db := mc.GetDb("slg_db")
	if db == nil {
		return
	}
	p.serverId = p.App().Settings().GetInt("server_id", 1)
	p.warReportEntry = entry.NewWarReportEntry(db.Collection("war_reports"), p.serverId)
}

func (p *ActorWar) scheduleArmy(army *model.Army) int32 {
	if army == nil {
		return code.InvalidParam
	}

	now := time.Now().Unix()
	arriveAt := army.End
	if arriveAt <= now {
		arriveAt = now + 1
	}

	p.movements[army.Id] = &ArmyMovement{
		Army:     army,
		ArriveAt: arriveAt,
	}

	clog.Debug("[ActorWar] scheduled army=%d, arrive=%d, cmd=%d", army.Id, arriveAt, army.Cmd)
	return code.OK
}

func (p *ActorWar) cancelArmy(army *model.Army) int32 {
	if _, ok := p.movements[army.Id]; ok {
		delete(p.movements, army.Id)
		army.Cmd = model.ArmyCmdBack
		army.End = time.Now().Unix() + 10
		p.movements[army.Id] = &ArmyMovement{
			Army:     army,
			ArriveAt: army.End,
		}
		return code.OK
	}
	return code.ArmyNotFound
}

func (p *ActorWar) checkArmyArrive() {
	now := time.Now().Unix()
	arrived := make([]*ArmyMovement, 0)

	for id, m := range p.movements {
		if now >= m.ArriveAt {
			arrived = append(arrived, m)
			delete(p.movements, id)
		}
	}

	for _, m := range arrived {
		p.processArrival(m)
	}
}

func (p *ActorWar) processArrival(m *ArmyMovement) {
	army := m.Army
	clog.Debug("[ActorWar] army arrived: id=%d, cmd=%d, to=(%d,%d)", army.Id, army.Cmd, army.ToX, army.ToY)

	switch army.Cmd {
	case model.ArmyCmdAttack:
		p.processAttack(m)
	case model.ArmyCmdDefend:
		p.processDefend(m)
	case model.ArmyCmdReclamation:
		p.processReclamation(m)
	case model.ArmyCmdBack:
		p.processBack(m)
	case model.ArmyCmdTransfer:
		p.processTransfer(m)
	default:
		p.processBack(m)
	}
}

func (p *ActorWar) processAttack(m *ArmyMovement) {
	army := m.Army

	attackSoldiers := 0
	for _, s := range army.Soldiers {
		attackSoldiers += s
	}

	defenseSoldiers := attackSoldiers / 2
	rounds := 3

	attackLoss := int(math.Ceil(float64(attackSoldiers) * 0.2))
	defenseLoss := int(math.Ceil(float64(defenseSoldiers) * 0.5))

	result := int8(2)
	if attackLoss >= attackSoldiers {
		result = 0
	}

	for i := range army.Soldiers {
		loss := int(math.Ceil(float64(army.Soldiers[i]) * 0.2))
		army.Soldiers[i] -= loss
		if army.Soldiers[i] < 0 {
			army.Soldiers[i] = 0
		}
	}

	report := &model.WarReport{
		Id:             int(time.Now().UnixNano() % 1000000),
		AttackRid:      army.RId,
		DefenseRid:     0,
		Result:         result,
		Rounds:         rounds,
		DestroyDurable: defenseLoss,
		X:              army.ToX,
		Y:              army.ToY,
		CreatedAt:      time.Now(),
	}
	if p.warReportEntry != nil {
		_ = p.warReportEntry.Insert(report)
	}

	army.Cmd = model.ArmyCmdBack
	now := time.Now().Unix()
	army.Start = now
	army.End = now + 30
	army.FromX = army.ToX
	army.FromY = army.ToY

	p.movements[army.Id] = &ArmyMovement{
		Army:     army,
		ArriveAt: army.End,
	}

	playerPath := p.NewChildPath("player", army.RId)
	p.Call(playerPath, "updateArmyResult", army)
}

func (p *ActorWar) processDefend(m *ArmyMovement) {
	army := m.Army
	army.Cmd = model.ArmyCmdIdle

	playerPath := p.NewChildPath("player", army.RId)
	p.Call(playerPath, "updateArmyResult", army)
}

func (p *ActorWar) processReclamation(m *ArmyMovement) {
	army := m.Army
	army.Cmd = model.ArmyCmdBack
	now := time.Now().Unix()
	army.Start = now
	army.End = now + 30

	p.movements[army.Id] = &ArmyMovement{
		Army:     army,
		ArriveAt: army.End,
	}

	playerPath := p.NewChildPath("player", army.RId)
	p.Call(playerPath, "updateArmyResult", army)
}

func (p *ActorWar) processBack(m *ArmyMovement) {
	army := m.Army
	army.Cmd = model.ArmyCmdIdle
	army.Start = 0
	army.End = 0

	playerPath := p.NewChildPath("player", army.RId)
	p.Call(playerPath, "updateArmyResult", army)
}

func (p *ActorWar) processTransfer(m *ArmyMovement) {
	p.processDefend(m)
}
