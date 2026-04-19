package mgr

import (
	"sync"

	clog "github.com/actorgo-game/actorgo/logger"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/llr104/slgserver/internal/data/model"
	"github.com/llr104/slgserver/internal/node/slgserver/static_conf/skill"
)

type skillMgr struct {
	mutex    sync.RWMutex
	skillMap map[int][]*model.Skill // key: rid
}

var SkillMgr = &skillMgr{
	skillMap: make(map[int][]*model.Skill),
}

func (this *skillMgr) Load() {
	c := coll(model.Skill{}.CollectionName())
	if c == nil {
		return
	}
	cx, cancel := ctx()
	defer cancel()
	cursor, err := c.Find(cx, withServer(bson.M{}))
	if err != nil {
		clog.Error("[SkillMgr] Load find err=%v", err)
		return
	}
	var rows []*model.Skill
	if err := cursor.All(cx, &rows); err != nil {
		clog.Error("[SkillMgr] Load decode err=%v", err)
		return
	}
	for _, v := range rows {
		if this.skillMap[v.RId] == nil {
			this.skillMap[v.RId] = make([]*model.Skill, 0)
		}
		this.skillMap[v.RId] = append(this.skillMap[v.RId], v)
	}
}

func (this *skillMgr) Get(rid int) ([]*model.Skill, bool) {
	this.mutex.RLock()
	r, ok := this.skillMap[rid]
	this.mutex.RUnlock()
	if ok {
		return r, true
	}

	c := coll(model.Skill{}.CollectionName())
	if c == nil {
		return nil, false
	}
	cx, cancel := ctx()
	defer cancel()
	cursor, err := c.Find(cx, withServer(bson.M{"rid": rid}))
	if err != nil {
		return nil, false
	}
	var m []*model.Skill
	if err := cursor.All(cx, &m); err != nil {
		return nil, false
	}
	if len(m) == 0 {
		return nil, false
	}
	this.mutex.Lock()
	this.skillMap[rid] = m
	this.mutex.Unlock()
	return m, true
}

func (this *skillMgr) GetSkillOrCreate(rid int, cfg int) (*model.Skill, bool) {
	if m, ok := this.Get(rid); ok {
		for _, v := range m {
			if v.CfgId == cfg {
				return v, true
			}
		}
	}

	c := coll(model.Skill{}.CollectionName())
	if c == nil {
		return nil, false
	}
	ret := model.NewSkill(rid, cfg)
	ret.ServerId = serverID()
	ret.Id = nextID(model.Skill{}.CollectionName())
	cx, cancel := ctx()
	defer cancel()
	if _, err := c.InsertOne(cx, ret); err != nil {
		clog.Warn("[SkillMgr] insert err=%v", err)
		return ret, false
	}
	this.mutex.Lock()
	if this.skillMap[rid] == nil {
		this.skillMap[rid] = make([]*model.Skill, 0)
	}
	this.skillMap[rid] = append(this.skillMap[rid], ret)
	this.mutex.Unlock()
	return ret, true
}

// SkillCfgInjector 注入到 model.GetSkillCfg。
func SkillCfgInjector(cfgId int) (model.SkillCfg, bool) {
	c, ok := skill.Skill.GetCfg(cfgId)
	if !ok {
		return model.SkillCfg{}, false
	}
	return model.SkillCfg{Limit: c.Limit, Arms: c.Arms}, true
}
