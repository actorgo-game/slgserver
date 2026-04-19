package skill

import (
	"encoding/json"
	"io/ioutil"
	"path"

	clog "github.com/actorgo-game/actorgo/logger"
	cprofile "github.com/actorgo-game/actorgo/profile"
)

var Skill skill

type skill struct {
	skills       []Conf
	skillConfMap map[int]Conf
	outline      outline
}

func (this *skill) Load() {
	this.skills = make([]Conf, 0)
	this.skillConfMap = make(map[int]Conf)

	jsonDir := cprofile.ConfigPath()
	fileName := path.Join(jsonDir, "skill", "skill_outline.json")
	jdata, err := ioutil.ReadFile(fileName)
	if err != nil {
		clog.Panic("[skill] load outline error: file=%s, err=%v", fileName, err)
		return
	}
	if err := json.Unmarshal(jdata, &this.outline); err != nil {
		clog.Panic("[skill] unmarshal outline error: file=%s, err=%v", fileName, err)
		return
	}

	rd, err := ioutil.ReadDir(path.Join(jsonDir, "skill"))
	if err != nil {
		clog.Panic("[skill] readdir error: %v", err)
		return
	}

	for _, r := range rd {
		if r.IsDir() {
			this.readSkill(path.Join(jsonDir, "skill", r.Name()))
		}
	}

	clog.Info("[skill] Skill loaded: %d skills", len(this.skills))
}

func (this *skill) readSkill(dir string) {
	rd, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}
	for _, r := range rd {
		if r.IsDir() {
			continue
		}
		jdata, err := ioutil.ReadFile(path.Join(dir, r.Name()))
		if err != nil {
			continue
		}
		conf := Conf{}
		if err := json.Unmarshal(jdata, &conf); err == nil {
			this.skills = append(this.skills, conf)
			this.skillConfMap[conf.CfgId] = conf
		} else {
			clog.Warn("[skill] unmarshal error: file=%s, err=%v",
				path.Join(dir, r.Name()), err)
		}
	}
}

func (this *skill) GetCfg(cfgId int) (Conf, bool) {
	cfg, ok := this.skillConfMap[cfgId]
	return cfg, ok
}
