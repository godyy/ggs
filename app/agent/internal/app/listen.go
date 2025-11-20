package app

import (
	"crypto/rand"
	"fmt"
	"net"

	inet "github.com/godyy/ggs/app/internal/net"

	"github.com/godyy/ggs/app/agent/internal/app/internal/agent"
	"github.com/godyy/ggs/app/agent/internal/config"
	icrypto "github.com/godyy/ggs/app/internal/crypto"
	"github.com/godyy/ggs/internal/libs/logger"
	pkgerrors "github.com/pkg/errors"
)

// startListen 启动监听.
func (a *app) startListen() error {
	port := config.GetConfig().Port
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return pkgerrors.WithMessagef(err, "listening at :%d", port)
	}

	logger.GetLogger().Infof("agent listening at :%d", port)
	a.listener = l
	go a.listenLoop()

	return nil
}

// stopListen 停止监听.
func (a *app) stopListen() {
	if err := a.listener.Close(); err != nil {
		logger.GetLogger().Errorf("close listener failed, %v", err)
	}
}

// listenLoop 监听循环.
func (a *app) listenLoop() {
	for {
		conn, err := a.listener.Accept()
		if err != nil {
			break
		}

		go a.handleConn(conn)
	}
}

// handleConn 处理连接.
func (a *app) handleConn(conn net.Conn) {
	// 交换密钥
	sessionKey, err := a.exchangeSecretKey(conn)
	if err != nil {
		logger.GetLogger().Errorf("exchange secret key failed, remote=%s, %v", conn.RemoteAddr().String(), err)
		conn.Close()
		return
	}

	// 创建 Agent
	agent, err := agent.NewAgent(a, conn, sessionKey)
	if err != nil {
		logger.GetLogger().Errorf("create agent failed, remote=%s, %v", conn.RemoteAddr().String(), err)
		conn.Close()
		return
	}

	// 启动 Agent
	agent.Start(false)
}

// exchangeSecretKey 交换密钥.
func (a *app) exchangeSecretKey(conn net.Conn) ([]byte, error) {
	// 读取临时密钥
	encryptedTmpSecret, err := inet.ReadPacket(conn)
	if err != nil {
		return nil, pkgerrors.WithMessage(err, "read tmp secret")
	}

	// 解密临时密钥
	tmpSecret, err := a.secretDecryptor.Decrypt(encryptedTmpSecret)
	if err != nil {
		return nil, pkgerrors.WithMessage(err, "decrypt tmp secret")
	}

	// 生成会话密钥
	sessionKey := make([]byte, 16)
	if _, err := rand.Read(sessionKey); err != nil {
		return nil, pkgerrors.WithMessage(err, "generate session key")
	}

	// 利用临时密钥创建会话密钥加密器
	sessionKeyEncryptor, err := icrypto.CreateAESCrypto(tmpSecret)
	if err != nil {
		return nil, pkgerrors.WithMessage(err, "create session key encryptor")
	}

	// 加密会话密钥
	encryptedSessionKey, err := sessionKeyEncryptor.Encrypt(sessionKey)
	if err != nil {
		return nil, pkgerrors.WithMessage(err, "encrypt session key")
	}

	// 发送加密后的会话密钥
	if err := inet.WritePacket(conn, encryptedSessionKey); err != nil {
		return nil, pkgerrors.WithMessage(err, "write encrypted session key")
	}

	return sessionKey, nil
}
