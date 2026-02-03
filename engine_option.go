package appx

import (
	"github.com/go-xuan/appx/serverx"
)

// Option 配置选项
type Option = func(e *Engine)

// SetConfig 设置预制服务配置
func SetConfig(config *serverx.Config) Option {
	return func(e *Engine) {
		e.SetConfig(config)
	}
}

// AddServer 添加服务
func AddServer(servers ...serverx.Server) Option {
	return func(e *Engine) {
		e.AddServer(servers...)
	}
}
