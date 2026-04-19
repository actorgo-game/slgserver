package facility

import (
	"encoding/json"
	"io/ioutil"

	clog "github.com/actorgo-game/actorgo/logger"
)

type conditions struct {
	Type  int `json:"type"`
	Level int `json:"level"`
}

type facility struct {
	Title      string       `json:"title"`
	Des        string       `json:"des"`
	Name       string       `json:"name"`
	Type       int8         `json:"type"`
	Additions  []int8       `json:"additions"`
	Conditions []conditions `json:"conditions"`
	Levels     []fLevel     `json:"levels"`
}

type NeedRes struct {
	Decree int `json:"decree"`
	Grain  int `json:"grain"`
	Wood   int `json:"wood"`
	Iron   int `json:"iron"`
	Stone  int `json:"stone"`
	Gold   int `json:"gold"`
}

type fLevel struct {
	Level  int     `json:"level"`
	Values []int   `json:"values"`
	Need   NeedRes `json:"need"`
	Time   int     `json:"time"` //升级需要的时间
}

func NewFacility(jsonName string) *facility {
	f := &facility{}
	f.load(jsonName)
	return f
}

func (this *facility) load(jsonName string) {
	jdata, err := ioutil.ReadFile(jsonName)
	if err != nil {
		clog.Panic("[facility] load file error: file=%s, err=%v", jsonName, err)
		return
	}
	if err := json.Unmarshal(jdata, this); err != nil {
		clog.Panic("[facility] unmarshal error: file=%s, err=%v", jsonName, err)
		return
	}
}
