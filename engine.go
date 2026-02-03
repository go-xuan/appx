package appx

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-xuan/configx"
	"github.com/go-xuan/utilx/errorx"
	log "github.com/sirupsen/logrus"

	"github.com/go-xuan/appx/serverx"
)

var engine *Engine // 服务

// NewEngine 初始化Engine
func NewEngine(options ...Option) *Engine {
	e := &Engine{
		servers: make([]serverx.Server, 0),
		running: false,
	}
	// 添加配置选项
	for _, option := range options {
		option(e)
	}
	return e
}

// GetEngine 获取当前Engine
func GetEngine() *Engine {
	if engine == nil {
		engine = NewEngine(SetConfig(serverx.DefaultConfig()))
	}
	return engine
}

// GetConfig 获取当前配置
func GetConfig() *serverx.Config {
	return GetEngine().config
}

// Engine 服务Engine
type Engine struct {
	config  *serverx.Config  // 服务启动配置
	servers []serverx.Server // http/grpc或者其他服务
	running bool             // 运行标识
}

// RUN 运行应用
func (e *Engine) RUN(ctx context.Context) {
	e.checkRunning()   // 检查服务是否已运行
	e.startServer(ctx) // 启动服务
	e.keepRunning(ctx) // 保持服务运行
	e.Shutdown(ctx)    // 关闭服务
}

// Shutdown 关闭服务
func (e *Engine) Shutdown(ctx context.Context) {
	serverx.Shutdown(ctx, e.servers...)
	e.servers = make([]serverx.Server, 0)
	e.running = false
	log.WithContext(ctx).Info("shutdown complete")
}

// SetConfig 设置服务配置
func (e *Engine) SetConfig(config *serverx.Config) {
	if config != nil {
		e.config = config
	}
}

// AddServer 添加服务
func (e *Engine) AddServer(servers ...serverx.Server) {
	e.servers = append(e.servers, servers...)
}

// 检查服务运行状态
func (e *Engine) checkRunning() {
	if e.running {
		panic("engine has already running")
	}
}

// 启动服务
func (e *Engine) startServer(ctx context.Context) {
	config := &serverx.Config{}
	if err := configx.LoadConfigurator(config); err != nil && e.config == nil {
		return
	}
	if e.config != nil {
		config.Cover(e.config)
	}
	errorx.Panic(serverx.Start(ctx, config, e.servers...))

}

// 保持服务运行，等待信号量关闭服务
func (e *Engine) keepRunning(ctx context.Context) {
	e.running = true

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	<-quit

	// 设定超时时间，确保服务有足够时间关闭
	_, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
}
