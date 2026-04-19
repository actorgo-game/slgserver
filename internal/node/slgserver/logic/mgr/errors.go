package mgr

import (
	"errors"

	"github.com/llr104/slgserver/internal/node/slgserver/db"
)

// ErrDBNotReady 当 mongo 还未注入时，所有 mgr 写入都返回它。
var ErrDBNotReady = errors.New("slgserver/mgr: mongo not ready")

// nextID 是 db.NextID 的简短别名，便于 mgr 包内引用。
func nextID(name string) int { return db.NextID(name) }

// serverID 是 db.ServerId 的简短别名。
func serverID() int { return db.ServerId() }
