package model

// 资源产出
type Yield struct {
	Wood  int
	Iron  int
	Stone int
	Grain int
	Gold  int
}

// IPushable 用于把 model 推给在线客户端，由 slgserver 节点在启动时实现 Push。
// 与原 net.ConnMgr.Push 接口同语义。
type IPushable interface {
	IsCellView() bool
	IsCanView(rid, x, y int) bool
	BelongToRId() []int
	PushMsgName() string
	Position() (int, int)
	TPosition() (int, int)
	ToProto() interface{}
}

// ===== 跨包/跨节点的回调钩子 =====
//
// 原 slgserver 在 `model/util.go` 用包级函数指针解耦 model -> mgr。
// 这里完整保留这一模式，slgserver 启动时（logic.BeforeInit）把这些指针指向
// 对应的 mgr 函数。

var ArmyIsInView func(rid, x, y int) bool
var GetUnionId func(rid int) int
var GetUnionName func(unionId int) string
var GetRoleNickName func(rid int) string
var GetParentId func(rid int) int
var GetMainMembers func(unionId int) []int
var GetYield func(rid int) Yield
var GetDepotCapacity func(rid int) int
var GetCityCost func(cid int) int8
var GetMaxDurable func(cid int) int
var GetCityLv func(cid int) int8
var MapResTypeLevel func(x, y int) (bool, int8, int8)

// SyncWriter 是 SyncExecute 的异步落库钩子。slgserver 节点在启动时把它指向
// 一个能根据具体类型分发到对应 mongo collection 的 worker。
// 其它节点（httpserver / loginserver）不设置该钩子，model 上的 SyncExecute
// 退化成只 Push 不落库。
var SyncWriter func(any)

// PushHook 是 Push() 方法的回调钩子。slgserver 节点把它指向 gateserver/
// online 的推送函数；其它节点保持 nil 即可。
var PushHook func(IPushable)

// ServerId 与原 model.ServerId 同语义。slgserver 节点启动时由
// run.Init -> db.Setup 间接设置成当前 serverId。
var ServerId = 0

// SetServerId 由 db.Setup 调用，向 model 同步 serverId。
func SetServerId(sid int) { ServerId = sid }

// pushIfHooked 为 Push() 提供统一入口，避免 nil deref。
func pushIfHooked(p IPushable) {
	if PushHook != nil {
		PushHook(p)
	}
}

// syncIfHooked 为 SyncExecute 提供统一入口，避免 nil deref。
func syncIfHooked(v any) {
	if SyncWriter != nil {
		SyncWriter(v)
	}
}
