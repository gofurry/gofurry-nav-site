package service

import (
	"fmt"
	"github.com/gofurry/gofurry-game-backend/common"
	"github.com/gofurry/gofurry-game-backend/common/log"
	"github.com/gofurry/gofurry-game-backend/roof/env"
	"github.com/sourcegraph/conc/pool"
	"sync"
)

// 事件总线
var EB = &EventBus{
	Subscribers: map[string]DataChannelSlice{},
}

type EventData struct {
	Data  any
	Topic string
}

// 接收数据的通道
type DataChannel chan EventData

// 包含接收数据的切片
type DataChannelSlice []DataChannel

type EventBus struct {
	Subscribers map[string]DataChannelSlice
	rwm         sync.RWMutex
}

// 发送线程池
var threadPool = pool.New().WithMaxGoroutines(env.GetServerConfig().Thread.EventPublishThread)

func (eb *EventBus) PublishGlobalMsg(data any) {
	eb.Publish(common.GLOBAL_MSG, data)
}

func (eb *EventBus) PublishCommonMsg(data any) {
	eb.Publish(common.COMMON_MSG, data)
}

func (eb *EventBus) PublishStatusMsg(data any) {
	eb.Publish(common.EVENT_STATUS_REPORT, data)
}

// 发布主题
func (eb *EventBus) Publish(topic string, data any) {
	eb.rwm.Lock()
	defer eb.rwm.Unlock()
	if chs, found := eb.Subscribers[topic]; found {
		channels := append(DataChannelSlice{}, chs...)
		threadPool.Go(func() {
			func(data EventData, dataChannelSlices DataChannelSlice) {
				defer func() {
					if e := recover(); e != nil {
						log.Error(fmt.Sprintf("Publish event error recover: %v", e))
					}
				}()
				for _, ch := range dataChannelSlices {
					ch <- data
				}
			}(EventData{Data: data, Topic: topic}, channels)
		})
	}
}

// 订阅主题
func (eb *EventBus) Subscribe(topic string, ch DataChannel) {
	eb.rwm.Lock()
	defer eb.rwm.Unlock()
	if prev, found := eb.Subscribers[topic]; found {
		eb.Subscribers[topic] = append(prev, ch)
	} else {
		eb.Subscribers[topic] = append([]DataChannel{}, ch)
	}
}

// 取消主题
func (eb *EventBus) UnSubscribe(topic string, ch DataChannel) {
	eb.rwm.Lock()
	defer eb.rwm.Unlock()
	if prev, found := eb.Subscribers[topic]; found {
		for i, channel := range prev {
			if channel == ch {
				news := append(prev[:i], prev[i+1:]...)
				eb.Subscribers[topic] = news
			}
		}
	}
}
