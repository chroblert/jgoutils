package jasync

import (
	"log"
	"sync"
)

// 全局配置
type globalConfig struct {
	MaxGoroutinCount      int
	realtimeGoroutinCount int
	mu                    *sync.RWMutex
}

func (p *globalConfig) AddGlobalGoroutinCount() {
	p.mu.Lock()
	p.realtimeGoroutinCount++
	p.mu.Unlock()
}
func (p *globalConfig) SubGlobalGoroutinCount() {
	p.mu.Lock()
	p.realtimeGoroutinCount--
	p.mu.Unlock()
}

func (p *globalConfig) Wait() {
	p.mu.RLock()
	for {
		if p.realtimeGoroutinCount < p.MaxGoroutinCount {
			break
		}
		log.Println("达到最大协程数量限制")
	}
	p.mu.RUnlock()
}

// async配置
// TaskMaxLimit 使用最大允许协程数量
type asyncConfig struct {
	TaskMaxLimit int
	mu           *sync.RWMutex
}

var (
	jasyncConf = asyncConfig{
		TaskMaxLimit: 200,
		mu:           &sync.RWMutex{},
	}
)
