package httpproto

type ServerCreateReq struct {
	ID     int64  `json:"id"`     // 服务器ID
	Name   string `json:"name"`   // 服务器名称
	NodeId string `json:"nodeId"` // 服务器所在节点ID
}
