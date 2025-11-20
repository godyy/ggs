package models

// Server 服务器
type Server struct {
	ID     int64  `bson:"id"`     // 服务器ID
	Name   string `bson:"name"`   // 服务器名称
	NodeId string `bson:"nodeId"` // 服务器所在节点ID
}
