package utils

import (
	"context"

	"github.com/godyy/ggs/app/internal/httputils"
	"github.com/godyy/ggs/app/login/httpproto"
)

func GetCharacterList(ctx context.Context, urlRoot string, token string) ([]httpproto.CharacterInfo, error) {
	resp := httpproto.GetCharacterListResp{}
	if err := httputils.GetJsonWithContext(ctx, urlRoot+"/character/list?token="+token, &resp, nil); err != nil {
		return nil, err
	}
	return resp.CharacterList, nil
}

func CreateCharacter(ctx context.Context, urlRoot string, token string, serverId int64) (int64, error) {
	req := httpproto.CreateCharacterReq{
		ServerID: serverId,
	}
	resp := httpproto.CreateCharacterResp{}
	if err := httputils.PostJsonWithContext(ctx, urlRoot+"/character/create?token="+token, &req, &resp, nil); err != nil {
		return 0, err
	}
	return resp.CharacterID, nil
}

func LoginCharacter(ctx context.Context, urlRoot string, token string, characterId int64) (string, error) {
	req := httpproto.CharacterLoginReq{
		CharacterID: characterId,
	}
	resp := httpproto.CharacterLoginResp{}
	if err := httputils.PostJsonWithContext(ctx, urlRoot+"/character/login?token="+token, &req, &resp, nil); err != nil {
		return "", err
	}
	return resp.Token, nil
}
