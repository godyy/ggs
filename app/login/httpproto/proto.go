package httpproto

// LoginReq 登录请求
type LoginReq struct {
	UID      int64  `json:"uid"`       // 用户ID.
	Token    string `json:"token"`     // 登陆令牌.
	Device   string `json:"device"`    // 设备名称.
	DeviceOS string `json:"device_os"` // 设备操作系统.
	DeviceID string `json:"device_id"` // 设备ID.
}

// LoginResp 登录响应
type LoginResp struct {
}

// ServerInfo 服务器信息
type ServerInfo struct {
	ID   int64  `json:"id"`   // 服务器ID.
	Name string `json:"name"` // 服务器名称.
}

// GetServerListReq 获取服务器列表请求
type GetServerListReq struct {
}

// GetServerListResp 获取服务器列表响应
type GetServerListResp struct {
	ServerList []ServerInfo `json:"server_list"` // 服务器列表.
}

// CharacterInfo 角色信息
type CharacterInfo struct {
	ID       int64  `json:"id"`        // 角色ID.
	Name     string `json:"name"`      // 角色名称.
	ServerID int64  `json:"server_id"` // 服务器ID
}

// GetCharacterListReq 获取角色列表请求
type GetCharacterListReq struct {
}

// GetCharacterListResp 获取角色列表响应
type GetCharacterListResp struct {
	CharacterList []CharacterInfo `json:"character_list"` // 角色列表.
}

// CreateCharacterReq 创建角色请求
type CreateCharacterReq struct {
	ServerID int64 `json:"server_id"` // 服务器ID.
}

// CreateCharacterResp 创建角色响应
type CreateCharacterResp struct {
	CharacterID int64 `json:"character_id"` // 角色ID.
}

// CharacterLoginReq 角色登录请求.
type CharacterLoginReq struct {
	CharacterID int64 `json:"character_id"` // 角色ID.
}

// CharacterLoginResp
type CharacterLoginResp struct {
	Token string `json:"token"` // 用于登录网管的令牌.
}
