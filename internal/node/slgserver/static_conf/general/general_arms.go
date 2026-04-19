package general

import (
	"encoding/json"
	"io/ioutil"
	"path"

	clog "github.com/actorgo-game/actorgo/logger"
	cprofile "github.com/actorgo-game/actorgo/profile"
)

var GenArms Arms

type gArmsCondition struct {
	Level     int `json:"level"`
	StarLevel int `json:"star_lv"`
}

type gArmsCost struct {
	Gold int `json:"gold"`
}

type gArms struct {
	Id         int            `json:"id"`
	Name       string         `json:"name"`
	Condition  gArmsCondition `json:"condition"`
	ChangeCost gArmsCost      `json:"change_cost"`
	HarmRatio  []int          `json:"harm_ratio"`
}

type Arms struct {
	Title string  `json:"title"`
	Arms  []gArms `json:"arms"`
	AMap  map[int]gArms
}

func (this *Arms) Load() {
	fileName := path.Join(cprofile.ConfigPath(), "general", "general_arms.json")
	jdata, err := ioutil.ReadFile(fileName)
	if err != nil {
		clog.Panic("[general] load arms error: file=%s, err=%v", fileName, err)
		return
	}
	if err := json.Unmarshal(jdata, this); err != nil {
		clog.Panic("[general] unmarshal arms error: file=%s, err=%v", fileName, err)
		return
	}

	this.AMap = make(map[int]gArms)
	for _, v := range this.Arms {
		this.AMap[v.Id] = v
	}
	clog.Info("[general] GenArms loaded: %d arms", len(this.Arms))
}

func (this *Arms) GetArm(id int) (gArms, error) {
	return this.AMap[id], nil
}

func (this *Arms) GetHarmRatio(attId, defId int) float64 {
	attArm, ok1 := this.AMap[attId]
	_, ok2 := this.AMap[defId]
	if ok1 && ok2 {
		return float64(attArm.HarmRatio[defId-1]) / 100.0
	}
	return 1.0
}
