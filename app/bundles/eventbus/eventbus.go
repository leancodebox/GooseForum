package eventbus

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/leancodebox/GooseForum/app/bundles/closer"
	"github.com/leancodebox/GooseForum/app/service/eventhandlers"
)

var (
	pubSub         *gochannel.GoChannel
	router         *message.Router
	eventBus       *cqrs.EventBus
	eventProcessor *cqrs.EventProcessor
	once           sync.Once
	logger         = &dynamicSlogLogger{}
	cancelFunc     context.CancelFunc
	stopWg         sync.WaitGroup
)

// Init 初始化事件总线并注册处理器
func Init(handlers ...cqrs.EventHandler) {
	InitEventBus()
	if len(handlers) > 0 {
		AddHandlers(handlers...)
	}
}

// InitEventBus 初始化事件总线
func InitEventBus() {
	once.Do(func() {
		pubSub = gochannel.NewGoChannel(
			gochannel.Config{
				// 设为 false 以实现真正的异步发布：发布者不需要等待订阅者处理完成
				BlockPublishUntilSubscriberAck: false,
				// 设置缓冲区大小，防止短时间内大量消息导致发布阻塞
				OutputChannelBuffer: 1024,
			},
			logger,
		)

		var err error
		router, err = message.NewRouter(message.RouterConfig{}, logger)
		if err != nil {
			panic(fmt.Errorf("failed to create router: %w", err))
		}

		// 注册到全局关闭管理器
		closer.Register(func() error {
			Close()
			return nil
		})

		// 添加中间件
		router.AddMiddleware(
			// 捕获处理器中的 panic
			middleware.Recoverer,

			// 最终错误处理：如果重试后仍然失败，记录错误并 Ack 消息，防止 gochannel 无限重试
			func(h message.HandlerFunc) message.HandlerFunc {
				return func(msg *message.Message) ([]*message.Message, error) {
					hEvents, hErr := h(msg)
					if hErr != nil {
						slog.Error("eventbus: message failed after retries, dropping it",
							"err", hErr,
							"msg_uuid", msg.UUID,
							"topic", message.SubscribeTopicFromCtx(msg.Context()),
						)
						// 返回 nil 错误表示 Ack，停止 gochannel 的重试循环
						return nil, nil
					}
					return hEvents, nil
				}
			},

			// 重试机制：最大重试 3 次，初始间隔 2s，指数退避
			middleware.Retry{
				MaxRetries:      3,
				InitialInterval: time.Second * 2,
				MaxInterval:     time.Second * 10,
				Multiplier:      2,
				Logger:          logger,
			}.Middleware,
		)

		marshaler := cqrs.JSONMarshaler{}

		eventBus, err = cqrs.NewEventBusWithConfig(pubSub, cqrs.EventBusConfig{
			GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
				return params.EventName, nil
			},
			Marshaler: marshaler,
		})
		if err != nil {
			panic(fmt.Errorf("failed to create event bus: %w", err))
		}

		eventProcessor, err = cqrs.NewEventProcessorWithConfig(router, cqrs.EventProcessorConfig{
			GenerateSubscribeTopic: func(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
				return params.EventName, nil
			},
			SubscriberConstructor: func(params cqrs.EventProcessorSubscriberConstructorParams) (message.Subscriber, error) {
				return pubSub, nil
			},
			Marshaler: marshaler,
			Logger:    logger,
		})
		if err != nil {
			panic(fmt.Errorf("failed to create event processor: %w", err))
		}
	})
}

// Publish 发布事件
func Publish(ctx context.Context, event any) error {
	if eventBus == nil {
		return fmt.Errorf("event bus not initialized")
	}
	return eventBus.Publish(ctx, event)
}

// AddHandlers 添加事件处理器
func AddHandlers(handlers ...cqrs.EventHandler) {
	if eventProcessor == nil {
		panic("event processor not initialized")
	}
	err := eventProcessor.AddHandlers(handlers...)
	if err != nil {
		panic(err)
	}
}

// Start 启动事件处理器
func Start() error {
	Init(eventhandlers.Handlers()...)
	if router == nil {
		return fmt.Errorf("router not initialized")
	}

	var ctx context.Context
	ctx, cancelFunc = context.WithCancel(context.Background())

	stopWg.Go(func() {
		slog.Info("starting event bus...")
		err := router.Run(ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			slog.Error("event bus router run error", "err", err)
		}
	})

	return nil
}

// Close 关闭事件总线
func Close() error {
	if cancelFunc != nil {
		cancelFunc()
	}

	// 等待 router 停止
	done := make(chan struct{})
	go func() {
		stopWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		slog.Info("event bus stopped")
	case <-time.After(3 * time.Second):
		slog.Warn("event bus shutdown timed out")
	}

	if pubSub != nil {
		return pubSub.Close()
	}
	return nil
}
