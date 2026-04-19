package mgr

import (
	"sync"
	"time"

	clog "github.com/actorgo-game/actorgo/logger"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/llr104/slgserver/internal/code"
	"github.com/llr104/slgserver/internal/data/model"
	"github.com/llr104/slgserver/internal/node/slgserver/global"
	"github.com/llr104/slgserver/internal/node/slgserver/static_conf"
	"github.com/llr104/slgserver/internal/util"
)

type roleBuildMgr struct {
	baseMutex    sync.RWMutex
	giveUpMutex  sync.RWMutex
	destroyMutex sync.RWMutex
	dbRB         map[int]*model.MapRoleBuild           // key: id
	posRB        map[int]*model.MapRoleBuild           // key: posId
	roleRB       map[int][]*model.MapRoleBuild         // key: rid
	giveUpRB     map[int64]map[int]*model.MapRoleBuild // key: time
	destroyRB    map[int64]map[int]*model.MapRoleBuild // key: time
}

var RBMgr = &roleBuildMgr{
	dbRB:      make(map[int]*model.MapRoleBuild),
	posRB:     make(map[int]*model.MapRoleBuild),
	roleRB:    make(map[int][]*model.MapRoleBuild),
	giveUpRB:  make(map[int64]map[int]*model.MapRoleBuild),
	destroyRB: make(map[int64]map[int]*model.MapRoleBuild),
}

func (this *roleBuildMgr) Load() {
	c := coll(model.MapRoleBuild{}.CollectionName())
	if c == nil {
		return
	}
	cx, cancel := ctx()
	defer cancel()

	// 检查并初始化系统建筑：若系统建筑数量与 NMMgr.sysBuild 不一致，重置后重建
	cnt, _ := c.CountDocuments(cx, withServer(bson.M{"type": bson.M{"$in": []int8{model.MapBuildSysCity, model.MapBuildSysFortress}}}))
	sysBuilds := NMMgr.SysBuilds()
	if int(cnt) != len(sysBuilds) {
		_, _ = c.DeleteMany(cx, withServer(bson.M{"type": bson.M{"$in": []int8{model.MapBuildSysCity, model.MapBuildSysFortress}}}))
		for _, sb := range sysBuilds {
			b := &model.MapRoleBuild{
				ServerId: serverID(),
				Id:       nextID(model.MapRoleBuild{}.CollectionName()),
				RId:      0,
				Type:     sb.Type,
				Level:    sb.Level,
				X:        sb.X,
				Y:        sb.Y,
			}
			b.Init()
			if _, err := c.InsertOne(cx, b); err != nil {
				clog.Error("[RBMgr] init sys build err=%v", err)
			}
		}
	}

	cursor, err := c.Find(cx, withServer(bson.M{}))
	if err != nil {
		clog.Error("[RBMgr] Load find err=%v", err)
		return
	}
	var rows []*model.MapRoleBuild
	if err := cursor.All(cx, &rows); err != nil {
		clog.Error("[RBMgr] Load decode err=%v", err)
		return
	}

	curTime := time.Now().Unix()

	for _, v := range rows {
		this.dbRB[v.Id] = v
		v.Init()

		if v.GiveUpTime != 0 {
			if _, ok := this.giveUpRB[v.GiveUpTime]; !ok {
				this.giveUpRB[v.GiveUpTime] = make(map[int]*model.MapRoleBuild)
			}
			this.giveUpRB[v.GiveUpTime][v.Id] = v
		}

		if v.OPLevel == 0 && v.Level != v.OPLevel {
			t := v.EndTime.Unix()
			if curTime >= t {
				v.ConvertToRes()
			} else {
				if _, ok := this.destroyRB[t]; !ok {
					this.destroyRB[t] = make(map[int]*model.MapRoleBuild)
				}
				this.destroyRB[t][v.Id] = v
			}
		}

		posId := global.ToPosition(v.X, v.Y)
		this.posRB[posId] = v
		if _, ok := this.roleRB[v.RId]; !ok {
			this.roleRB[v.RId] = make([]*model.MapRoleBuild, 0)
		}
		this.roleRB[v.RId] = append(this.roleRB[v.RId], v)

		if v.GiveUpTime != 0 && v.GiveUpTime <= curTime {
			this.RemoveFromRole(v)
		}
	}
}

func (this *roleBuildMgr) CheckGiveUp() []int {
	var ret []int
	var builds []*model.MapRoleBuild

	curTime := time.Now().Unix()
	this.giveUpMutex.Lock()
	for i := curTime - 10; i <= curTime; i++ {
		if gs, ok := this.giveUpRB[i]; ok {
			for _, g := range gs {
				builds = append(builds, g)
				ret = append(ret, global.ToPosition(g.X, g.Y))
			}
		}
	}
	this.giveUpMutex.Unlock()

	for _, build := range builds {
		this.RemoveFromRole(build)
	}
	return ret
}

func (this *roleBuildMgr) CheckDestroy() []int {
	var ret []int
	var builds []*model.MapRoleBuild

	curTime := time.Now().Unix()
	this.destroyMutex.Lock()
	for i := curTime - 10; i <= curTime; i++ {
		if gs, ok := this.destroyRB[i]; ok {
			for _, g := range gs {
				builds = append(builds, g)
				ret = append(ret, global.ToPosition(g.X, g.Y))
			}
		}
	}
	this.destroyMutex.Unlock()

	for _, b := range builds {
		b.ConvertToRes()
		b.SyncExecute()
	}
	return ret
}

func (this *roleBuildMgr) IsEmpty(x, y int) bool {
	this.baseMutex.RLock()
	defer this.baseMutex.RUnlock()
	posId := global.ToPosition(x, y)
	_, ok := this.posRB[posId]
	return !ok
}

func (this *roleBuildMgr) PositionBuild(x, y int) (*model.MapRoleBuild, bool) {
	this.baseMutex.RLock()
	defer this.baseMutex.RUnlock()
	posId := global.ToPosition(x, y)
	b, ok := this.posRB[posId]
	if ok {
		return b, true
	}
	return nil, false
}

func (this *roleBuildMgr) RoleFortressCnt(rid int) int {
	bs, ok := this.GetRoleBuild(rid)
	if !ok {
		return 0
	}
	cnt := 0
	for _, b := range bs {
		if b.IsRoleFortress() {
			cnt++
		}
	}
	return cnt
}

func (this *roleBuildMgr) AddBuild(rid, x, y int) (*model.MapRoleBuild, bool) {
	posId := global.ToPosition(x, y)
	this.baseMutex.Lock()
	rb, ok := this.posRB[posId]
	this.baseMutex.Unlock()
	if ok {
		rb.RId = rid
		this.baseMutex.Lock()
		if _, ok := this.roleRB[rid]; !ok {
			this.roleRB[rid] = make([]*model.MapRoleBuild, 0)
		}
		this.roleRB[rid] = append(this.roleRB[rid], rb)
		this.baseMutex.Unlock()
		return rb, true
	}

	if b, ok := NMMgr.PositionBuild(x, y); ok {
		if cfg, _ := static_conf.MapBuildConf.BuildConfig(b.Type, b.Level); cfg != nil {
			rb := &model.MapRoleBuild{
				ServerId:   serverID(),
				Id:         nextID(model.MapRoleBuild{}.CollectionName()),
				RId:        rid, X: x, Y: y,
				Type: b.Type, Level: b.Level, OPLevel: b.Level,
				Name: cfg.Name, CurDurable: cfg.Durable,
				MaxDurable: cfg.Durable,
			}
			rb.Init()

			c := coll(model.MapRoleBuild{}.CollectionName())
			if c == nil {
				return nil, false
			}
			cx, cancel := ctx()
			defer cancel()
			if _, err := c.InsertOne(cx, rb); err != nil {
				clog.Warn("[RBMgr] insert build err=%v", err)
				return nil, false
			}
			this.baseMutex.Lock()
			this.posRB[posId] = rb
			this.dbRB[rb.Id] = rb
			if _, ok := this.roleRB[rid]; !ok {
				this.roleRB[rid] = make([]*model.MapRoleBuild, 0)
			}
			this.roleRB[rid] = append(this.roleRB[rid], rb)
			this.baseMutex.Unlock()
			return rb, true
		}
	}
	return nil, false
}

func (this *roleBuildMgr) RemoveFromRole(build *model.MapRoleBuild) {
	this.baseMutex.Lock()
	if rb, ok := this.roleRB[build.RId]; ok {
		for i, v := range rb {
			if v.Id == build.Id {
				this.roleRB[build.RId] = append(rb[:i], rb[i+1:]...)
				break
			}
		}
	}
	this.baseMutex.Unlock()

	t := build.GiveUpTime
	this.giveUpMutex.Lock()
	if ms, ok := this.giveUpRB[t]; ok {
		delete(ms, build.Id)
	}
	this.giveUpMutex.Unlock()

	t = build.EndTime.Unix()
	this.destroyMutex.Lock()
	if ms, ok := this.destroyRB[t]; ok {
		delete(ms, build.Id)
	}
	this.destroyMutex.Unlock()

	build.Reset()
	build.SyncExecute()
}

func (this *roleBuildMgr) GetRoleBuild(rid int) ([]*model.MapRoleBuild, bool) {
	this.baseMutex.RLock()
	defer this.baseMutex.RUnlock()
	ra, ok := this.roleRB[rid]
	return ra, ok
}

func (this *roleBuildMgr) BuildCnt(rid int) int {
	bs, ok := this.GetRoleBuild(rid)
	if ok {
		return len(bs)
	}
	return 0
}

func (this *roleBuildMgr) Scan(x, y int) []*model.MapRoleBuild {
	if x < 0 || x >= global.MapWith || y < 0 || y >= global.MapHeight {
		return nil
	}
	this.baseMutex.RLock()
	defer this.baseMutex.RUnlock()

	minX := util.MaxInt(0, x-ScanWith)
	maxX := util.MinInt(global.MapWith, x+ScanWith)
	minY := util.MaxInt(0, y-ScanHeight)
	maxY := util.MinInt(global.MapHeight, y+ScanHeight)

	rb := make([]*model.MapRoleBuild, 0)
	for i := minX; i <= maxX; i++ {
		for j := minY; j <= maxY; j++ {
			if v, ok := this.posRB[global.ToPosition(i, j)]; ok && v.RId != 0 {
				rb = append(rb, v)
			}
		}
	}
	return rb
}

func (this *roleBuildMgr) ScanBlock(x, y, length int) []*model.MapRoleBuild {
	if x < 0 || x >= global.MapWith || y < 0 || y >= global.MapHeight {
		return nil
	}
	this.baseMutex.RLock()
	defer this.baseMutex.RUnlock()

	maxX := util.MinInt(global.MapWith, x+length-1)
	maxY := util.MinInt(global.MapHeight, y+length-1)

	rb := make([]*model.MapRoleBuild, 0)
	for i := x; i <= maxX; i++ {
		for j := y; j <= maxY; j++ {
			if v, ok := this.posRB[global.ToPosition(i, j)]; ok && (v.RId != 0 || v.IsSysCity() || v.IsSysFortress()) {
				rb = append(rb, v)
			}
		}
	}
	return rb
}

func (this *roleBuildMgr) BuildIsRId(x, y, rid int) bool {
	b, ok := this.PositionBuild(x, y)
	if ok {
		return b.RId == rid
	}
	return false
}

func (this *roleBuildMgr) GetYield(rid int) model.Yield {
	builds, ok := this.GetRoleBuild(rid)
	var y model.Yield
	if ok {
		for _, b := range builds {
			y.Iron += b.Iron
			y.Wood += b.Wood
			y.Grain += b.Grain
			y.Stone += b.Stone
		}
	}
	return y
}

func (this *roleBuildMgr) GiveUp(x, y int) int32 {
	b, ok := this.PositionBuild(x, y)
	if !ok {
		return code.CannotGiveUp
	}
	if b.IsWarFree() {
		return code.BuildWarFree
	}
	if b.GiveUpTime > 0 {
		return code.BuildGiveUpAlready
	}

	b.GiveUpTime = time.Now().Unix() + static_conf.Basic.Build.GiveUpTime
	b.SyncExecute()

	this.giveUpMutex.Lock()
	if _, ok := this.giveUpRB[b.GiveUpTime]; !ok {
		this.giveUpRB[b.GiveUpTime] = make(map[int]*model.MapRoleBuild)
	}
	this.giveUpRB[b.GiveUpTime][b.Id] = b
	this.giveUpMutex.Unlock()

	return code.OK
}

func (this *roleBuildMgr) Destroy(x, y int) int32 {
	b, ok := this.PositionBuild(x, y)
	if !ok {
		return code.BuildNotMe
	}
	if !b.IsHaveModifyLVAuth() || b.IsInGiveUp() || b.IsBusy() {
		return code.CanNotDestroy
	}
	cfg, ok := static_conf.MapBCConf.BuildConfig(b.Type, b.Level)
	if !ok {
		return code.InvalidParam
	}

	if c := RResMgr.TryUseNeed(b.RId, cfg.Need); c != code.OK {
		return c
	}

	b.EndTime = time.Now().Add(time.Duration(cfg.Time) * time.Second)
	this.destroyMutex.Lock()
	t := b.EndTime.Unix()
	if _, ok := this.destroyRB[t]; !ok {
		this.destroyRB[t] = make(map[int]*model.MapRoleBuild)
	}
	this.destroyRB[t][b.Id] = b
	this.destroyMutex.Unlock()

	mbCfg := model.MapBuildCfg{
		Type:     cfg.Type,
		Level:    cfg.Level,
		Name:     cfg.Name,
		Durable:  cfg.Durable,
		Defender: cfg.Defender,
		Time:     cfg.Time,
	}
	b.DelBuild(mbCfg)
	b.SyncExecute()
	return code.OK
}
