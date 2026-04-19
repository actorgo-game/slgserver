package npc

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"path"

	clog "github.com/actorgo-game/actorgo/logger"
	cprofile "github.com/actorgo-game/actorgo/profile"
)

var Cfg npc

type ArmyCfg struct {
	Lvs    []int8 `json:"lvs"`
	CfgIds []int  `json:"cfgIds"`
}

type armyArray struct {
	Des      string    `json:"des"`
	Soldiers int       `json:"soldiers"`
	ArmyCfg  []ArmyCfg `json:"army"`
}

type npc struct {
	Des   string      `json:"des"`
	Armys []armyArray `json:"armys"`
}

func (this *npc) Load() {
	fileName := path.Join(cprofile.ConfigPath(), "npc", "npc_army.json")
	jdata, err := ioutil.ReadFile(fileName)
	if err != nil {
		clog.Panic("[npc] load file error: file=%s, err=%v", fileName, err)
		return
	}
	if err := json.Unmarshal(jdata, this); err != nil {
		clog.Panic("[npc] unmarshal error: file=%s, err=%v", fileName, err)
		return
	}
	clog.Info("[npc] Cfg loaded: %d levels", len(this.Armys))
}

func (this *npc) NPCSoilder(level int8) int {
	if int(level) > len(this.Armys) || level <= 0 {
		return 0
	}
	return this.Armys[level-1].Soldiers
}

func (this *npc) RandomOne(level int8) (bool, *ArmyCfg) {
	if int(level) > len(this.Armys) || level <= 0 {
		return false, nil
	}
	r := rand.Intn(len(this.Armys[level-1].ArmyCfg))
	return true, &this.Armys[level-1].ArmyCfg[r]
}
