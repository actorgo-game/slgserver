package util

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/forgoer/openssl"
)

const validTime = 30 * 24 * time.Hour
const sessionKey = "1234567890123456"

type Session struct {
	MTime time.Time
	Id    int
}

func NewSession(id int, t time.Time) *Session {
	return &Session{Id: id, MTime: t}
}

func ParseSession(session string) (*Session, error) {
	if session == "" {
		return nil, errors.New("session is empty")
	}
	decode, err := base64.StdEncoding.DecodeString(session)
	if err != nil {
		return nil, err
	}

	data, _ := AesCBCDecrypt(decode, []byte(sessionKey), []byte(sessionKey), openssl.ZEROS_PADDING)
	arr := strings.Split(string(data), "|")
	if len(arr) != 2 {
		return nil, errors.New("session format error")
	}

	id, err := strconv.Atoi(arr[0])
	if err != nil {
		return nil, err
	}

	t, err := time.Parse("2006-01-02 15:04:05", arr[1])
	if err != nil {
		return nil, err
	}

	return &Session{Id: id, MTime: t}, nil
}

func (s *Session) String() string {
	timeStr := s.MTime.Format("2006-01-02 15:04:05")
	str := fmt.Sprintf("%d|%s", s.Id, timeStr)
	data, _ := AesCBCEncrypt([]byte(str), []byte(sessionKey), []byte(sessionKey), openssl.ZEROS_PADDING)
	encode := base64.StdEncoding.EncodeToString(data)
	return encode
}

func (s *Session) IsValid() bool {
	diff := time.Since(s.MTime)
	return diff-validTime < 0
}
