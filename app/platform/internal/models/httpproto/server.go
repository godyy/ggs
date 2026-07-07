package httpproto

type ServerCreateReq struct {
	ID   int64  `json:"id"`   // 服务器ID
	Name string `json:"name"` // 服务器名称
}

type ServerReloadGDConfReq struct {
	AllServers bool     `json:"all_servers"` // 是否重载所有服务器配置
	ServerIds  []int64  `json:"server_ids"`  // 服务器ID列表
	All        bool     `json:"all"`         // 是否加载所有配置
	Tables     []string `json:"tables"`      // 要加载的配置表
}

type ServerReloadFailed struct {
	ServerId int64  `json:"server_id"` // 服务器ID
	Error    string `json:"error"`     // 错误信息
}

type ServerReloadGDConfResp struct {
	Success      bool                 `json:"success"`       // 是否成功
	ServerFailed []ServerReloadFailed `json:"server_failed"` // 失败的服务器ID列表
}
