package httputils

import (
	"context"
	"errors"
	"net/http"

	"github.com/godyy/ggs/internal/utils/httputils"
)

type commonResp struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func checkCommonResp(resp *commonResp) error {
	if resp.Code != 0 {
		return errors.New(resp.Msg)
	}
	return nil
}

func GetJson(url string, resp any, headerFunc func(http.Header)) error {
	return GetJsonWithContext(context.Background(), url, resp, headerFunc)
}

func GetJsonWithContext(ctx context.Context, url string, resp any, headerFunc func(http.Header)) error {
	commonResp := &commonResp{
		Data: resp,
	}
	if err := httputils.GetJsonWithContext(ctx, url, &commonResp, headerFunc); err != nil {
		return err
	}
	if err := checkCommonResp(commonResp); err != nil {
		return err
	}
	return nil
}

func PostJson(url string, req any, resp any, headerFunc func(http.Header)) error {
	return PostJsonWithContext(context.Background(), url, req, resp, headerFunc)
}

func PostJsonWithContext(ctx context.Context, url string, req any, resp any, headerFunc func(http.Header)) error {
	commonResp := &commonResp{
		Data: resp,
	}
	if err := httputils.PostJsonWithContext(ctx, url, req, &commonResp, headerFunc); err != nil {
		return err
	}
	if err := checkCommonResp(commonResp); err != nil {
		return err
	}
	return nil
}
