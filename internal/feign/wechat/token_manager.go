// Package wechat @Author larry
// @Date 2026/06/15
// @Desc access_token 内存缓存管理（框架无 Redis，单实例部署下安全）
//
// 微信 access_token 有效期 7200s、新 token 顶掉旧的。这里内存缓存 + 到期前刷新。
// 双主体：小程序与服务号是两个独立 token（各自 appid/secret），按 account 分桶缓存、互不顶掉。
// ⚠️ 多实例部署会各自刷新互相顶掉，需引入外部存储/分布式锁（见计划 §6.1 多实例前提）。

package wechat

import (
	"fmt"
	"sync"
	"time"

	"warm-nest/pkg/tool/client/utils"

	"warm-nest/internal/config"
	"warm-nest/internal/feign/wechat/model"
)

// tokenCache 单个主体的 access_token 内存缓存
type tokenCache struct {
	mu       sync.Mutex
	token    string
	expireAt int64 // 毫秒，本地推算的过期时刻
}

// tokenStores 按 SignAccount 分桶（小程序/服务号各一份），各自独立锁与过期
var tokenStores = map[string]*tokenCache{
	SignAccountMini:     {},
	SignAccountOfficial: {},
}

// appCredential 取某主体的 appid/secret（小程序 vs 服务号）
func appCredential(account string) (appId, secret string, err error) {
	conf := config.WeChatConf()
	switch account {
	case SignAccountMini:
		return conf.AppId, conf.Secret, nil
	case SignAccountOfficial:
		return conf.OfficialAppId, conf.OfficialSecret, nil
	default:
		return "", "", fmt.Errorf("unknown wechat sign account: %s", account)
	}
}

// GetAccessToken 取某主体可用 access_token（缓存有效直接返回，否则刷新）
func GetAccessToken(account string) (string, error) {
	store, ok := tokenStores[account]
	if !ok {
		return "", fmt.Errorf("unknown wechat sign account: %s", account)
	}
	store.mu.Lock()
	defer store.mu.Unlock()

	if store.token != "" && nowMs() < store.expireAt {
		return store.token, nil
	}
	return refreshLocked(account, store)
}

// Refresh 主动刷新某主体 token（定时任务调用）
func Refresh(account string) error {
	store, ok := tokenStores[account]
	if !ok {
		return fmt.Errorf("unknown wechat sign account: %s", account)
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	_, err := refreshLocked(account, store)
	return err
}

// refreshLocked 拉取某主体新 token，调用方必须持有该 store 的锁
func refreshLocked(account string, store *tokenCache) (string, error) {
	conf := config.WeChatConf()
	if conf.Mock {
		store.token = "mock_access_token_" + account
		store.expireAt = nowMs() + conf.TokenRefreshSec*1000
		return store.token, nil
	}
	appId, secret, err := appCredential(account)
	if err != nil {
		return "", err
	}
	resp, err := utils.ResultPtrV2[model.AccessTokenResp](
		Client().Request().
			SetQueryParam("grant_type", "client_credential").
			SetQueryParam("appid", appId).
			SetQueryParam("secret", secret).
			SetResult(model.AccessTokenResp{}).
			Get("/cgi-bin/token"),
	)
	if err != nil {
		return "", fmt.Errorf("fetch wechat access_token (%s): %w", account, err)
	}
	if resp.ErrCode != 0 || resp.AccessToken == "" {
		return "", fmt.Errorf("wechat token (%s) errcode=%d errmsg=%s", account, resp.ErrCode, resp.ErrMsg)
	}
	store.token = resp.AccessToken
	// 用配置的刷新间隔推算过期（< 微信 7200s 留余量）
	store.expireAt = nowMs() + conf.TokenRefreshSec*1000
	return store.token, nil
}

func nowMs() int64 {
	return time.Now().UnixMilli()
}
