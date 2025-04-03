package dseworker

import (
	"fmt"

	"github.com/johandrevandeventer/dse-worker/internal/workers"
	"github.com/johandrevandeventer/dse-worker/internal/workers/dse_worker/dse890"
	"github.com/johandrevandeventer/dse-worker/internal/workers/types"
	"github.com/johandrevandeventer/kafkaclient/payload"
	"go.uber.org/zap"
)

const (
	DSETopicPrefix = "Rubicon/DSE/"
	WorkerTitle    = "DSE"
)

type Worker struct {
	decoder   *Decoder
	processor *Processor
	logger    *zap.Logger
}

func NewWorker(logger *zap.Logger) *Worker {
	decoder := NewDecoder()
	processor := NewProcessor(logger)

	// Priority 1
	decoder.RegisterDecoder("DSE890", dse890.Decoder)
	processor.RegisterProcessor("DSE890", dse890.Processor)

	return &Worker{
		decoder:   decoder,
		processor: processor,
		logger:    logger,
	}
}

func (w *Worker) RunWorker(msg []byte) (messageInfo *types.MessageInfo, err error) {
	p, err := payload.Deserialize(msg)
	if err != nil {
		return messageInfo, fmt.Errorf("failed to deserialize data: %w", err)
	}

	w.logger.Info("Running worker", zap.String("worker", WorkerTitle), zap.String("topic", p.MqttTopic), zap.String("id", p.ID.String()))

	trimmedTopic := workers.TrimPrefix(p.MqttTopic, DSETopicPrefix)
	w.logger.Debug("Validating customer", zap.String("topic", trimmedTopic))

	customer, err := workers.GetValidCustomer(trimmedTopic)
	if err != nil {
		return messageInfo, fmt.Errorf("customer validation failed: %w", err)
	}

	decodedPayloadInfo, err := w.decoder.DecodePayload(p.Message)
	if err != nil {
		return messageInfo, fmt.Errorf("failed to decode payload: %w", err)
	}

	w.logger.Debug(fmt.Sprintf("%s :: %s", WorkerTitle, customer))

	messageInfo, err = w.processor.ProcessPayload(decodedPayloadInfo.Type, *p)
	if err != nil {
		return messageInfo, fmt.Errorf("failed to process payload: %w", err)
	}

	return messageInfo, nil
}
