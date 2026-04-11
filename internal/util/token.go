package util

import (
	"encoding/json"
	"fmt"

	ccrypto "github.com/actorgo-game/actorgo/extend/crypto"
	ctime "github.com/actorgo-game/actorgo/extend/time"
	clogger "github.com/actorgo-game/actorgo/logger"
	"github.com/llr104/slgserver/internal/code"
)

const (
	hashFormat      = "pid:%d,openid:%s,timestamp:%d"
	tokenExpiredDay = 3
)

type Token struct {
	PID       int32  `json:"pid"`
	OpenID    string `json:"open_id"`
	Timestamp int64  `json:"tt"`
	Hash      string `json:"hash"`
}

func New(pid int32, openId string, appKey string) *Token {
	token := &Token{
		PID:       pid,
		OpenID:    openId,
		Timestamp: ctime.Now().ToMillisecond(),
	}

	token.Hash = BuildHash(token, appKey)
	return token
}

func (t *Token) ToBase64() string {
	bytes, _ := json.Marshal(t)
	return ccrypto.Base64Encode(string(bytes))
}

func DecodeToken(base64Token string) (*Token, bool) {
	if len(base64Token) < 1 {
		return nil, false
	}

	token := &Token{}
	bytes, err := ccrypto.Base64DecodeBytes(base64Token)
	if err != nil {
		clogger.Warn("base64Token = %s, validate error = %v", base64Token, err)
		return nil, false
	}

	err = json.Unmarshal(bytes, token)
	if err != nil {
		clogger.Warn("base64Token = %s, unmarshal error = %v", base64Token, err)
		return nil, false
	}

	return token, true
}

func Validate(token *Token, appKey string) (int32, bool) {
	now := ctime.Now()
	now.AddDays(tokenExpiredDay)

	if token.Timestamp > now.ToMillisecond() {
		clogger.Warn("token is expired, token = %s", token)
		return code.AccountTokenValidateFail, false
	}

	newHash := BuildHash(token, appKey)
	if newHash != token.Hash {
		clogger.Warn("hash validate fail. newHash = %s, token = %s", token)
		return code.AccountTokenValidateFail, false
	}

	return code.OK, true
}

func BuildHash(t *Token, appKey string) string {
	value := fmt.Sprintf(hashFormat, t.PID, t.OpenID, t.Timestamp)
	return ccrypto.MD5(value + appKey)
}
