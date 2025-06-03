package consumer

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/IBM/sarama"
	"github.com/hoangdv99/morgana/internal/configs"
	"github.com/hoangdv99/morgana/internal/utils"
	"go.uber.org/zap"
)

type HandlerFunc func(ctx context.Context, queueName string, payload []byte) error

type Consumer interface {
	RegisterHandler(queueName string, handleFunc HandlerFunc)
	Start(ctx context.Context) error
}
type consumer struct {
	saramaConsumer            sarama.ConsumerGroup
	queueNameToHandlerFuncMap map[string]HandlerFunc
	logger                    *zap.Logger
}

type consumerHandler struct {
	handlerFunc       HandlerFunc
	exitSignalChannel chan os.Signal
}

func newConsumerHanler(
	handlerFunc HandlerFunc,
	exitSignalChannel chan os.Signal,
) *consumerHandler {
	return &consumerHandler{
		handlerFunc:       handlerFunc,
		exitSignalChannel: exitSignalChannel,
	}
}

func (h consumerHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h consumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h consumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				session.Commit()
				return nil
			}

			if err := h.handlerFunc(session.Context(), message.Topic, message.Value); err != nil {
				return err
			}

		case <-h.exitSignalChannel:
			session.Commit()
			break
		}
	}
}

func newSaramaConfig(mqConfig configs.MQ) *sarama.Config {
	saramaConfig := sarama.NewConfig()
	saramaConfig.ClientID = mqConfig.ClientID
	saramaConfig.Metadata.Full = true

	return saramaConfig
}

func NewConsumer(
	mqConfig configs.MQ,
	logger *zap.Logger,
) (Consumer, error) {
	saramaConsumer, err := sarama.NewConsumerGroup(mqConfig.Addresses, mqConfig.ClientID, newSaramaConfig(mqConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to create sarama consumer: %w", err)
	}

	return &consumer{
		saramaConsumer:            saramaConsumer,
		logger:                    logger,
		queueNameToHandlerFuncMap: make(map[string]HandlerFunc),
	}, nil
}

func (c *consumer) RegisterHandler(queueName string, handlerFunc HandlerFunc) {
	c.queueNameToHandlerFuncMap[queueName] = handlerFunc
}

func (c consumer) Start(ctx context.Context) error {
	logger := utils.LoggerWithContext(ctx, c.logger)

	exitSignalChannel := make(chan os.Signal, 1)
	signal.Notify(exitSignalChannel, os.Interrupt)

	for queueName, handlerFunc := range c.queueNameToHandlerFuncMap {
		go func(queueName string, handlerFunc HandlerFunc) {
			err := c.saramaConsumer.Consume(
				context.Background(),
				[]string{queueName},
				newConsumerHanler(handlerFunc, exitSignalChannel),
			)
			if err != nil {
				logger.
					With(zap.String("queue_name", queueName)).
					With(zap.Error(err)).
					Error("failed to consume message from queue")
			}
		}(queueName, handlerFunc)
	}
	<-exitSignalChannel

	return nil
}
