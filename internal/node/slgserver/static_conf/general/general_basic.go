package general

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"path"

	clog "github.com/actorgo-game/actorgo/logger"
	cprofile "github.com/actorgo-game/actorgo/profile"
)

var GenBasic Basic

type gLevel struct {
	Level    int8 `json:"level"`
	Exp      int  `json:"exp"`
	Soldiers int  `json:"soldiers"`
}

type Basic struct {
	Title  string   `json:"title"`
	Levels []gLevel `json:"levels"`
}

func (this *Basic) Load() {
	fileName := path.Join(cprofile.ConfigPath(), "general", "general_basic.json")
	jdata, err := ioutil.ReadFile(fileName)
	if err != nil {
		clog.Panic("[general] load basic error: file=%s, err=%v", fileName, err)
		return
	}
	if err := json.Unmarshal(jdata, this); err != nil {
		clog.Panic("[general] unmarshal basic error: file=%s, err=%v", fileName, err)
		return
	}
	clog.Info("[general] GenBasic loaded: %d levels", len(this.Levels))

	General.Load()
	GenArms.Load()
}

func (this *Basic) GetLevel(l int8) (*gLevel, error) {
	if l <= 0 {
		return nil, errors.New("level error")
	}
	if int(l) <= len(this.Levels) {
		return &this.Levels[l-1], nil
	}
	return nil, errors.New("level error")
}

func (this *Basic) ExpToLevel(exp int) (int8, int) {
	var level int8 = 0
	limitExp := this.Levels[len(this.Levels)-1].Exp
	for _, v := range this.Levels {
		if exp >= v.Exp && v.Level > level {
			level = v.Level
		}
	}

	if limitExp < exp {
		return level, limitExp
	}
	return level, exp
}
