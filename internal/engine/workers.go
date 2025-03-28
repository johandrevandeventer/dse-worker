package engine

import (
	"encoding/json"

	"github.com/johandrevandeventer/dse-worker/internal/flags"
	"github.com/johandrevandeventer/dse-worker/internal/workers"
	dseworker "github.com/johandrevandeventer/dse-worker/internal/workers/dse_worker"
	"github.com/johandrevandeventer/kafkaclient/payload"
	"github.com/johandrevandeventer/logging"
	"go.uber.org/zap"
)

func (e *Engine) startWorker() {
	e.logger.Info("Starting DSE workers")

	var workersLogger *zap.Logger
	var kafkaProducerLogger *zap.Logger
	if flags.FlagWorkersLogging {
		workersLogger = logging.GetLogger("workers")
		kafkaProducerLogger = logging.GetLogger("kafka.producer")
	} else {
		workersLogger = zap.NewNop()
		kafkaProducerLogger = zap.NewNop()
	}

	for {
		select {
		case <-e.ctx.Done(): // Handle context cancellation (e.g., Ctrl+C)
			e.logger.Info("Stopping worker due to context cancellation")
			return
		case data, ok := <-e.kafkaConsumer.GetOutputChannel():
			if !ok { // Channel is closed
				e.logger.Info("Kafka consumer output channel closed, stopping worker")
				return
			}
			// Process the data (e.g., call DSEWorker)
			rawData, processedData, err := dseworker.DSEWorker(data, workersLogger)
			if err != nil {
				workersLogger.Error("Failed to process data", zap.Error(err))
				continue
			}

			if workers.IsEmpty(*rawData) && workers.IsEmpty(*processedData) {
				workersLogger.Warn("Empty data received, skipping")
				continue
			}

			if workers.IsEmpty(*rawData) {
				workersLogger.Warn("Empty raw data received, skipping")
				continue
			}

			if workers.IsEmpty(*processedData) {
				workersLogger.Warn("Empty processed data received, skipping")
				continue
			}

			serializedRawData, err := json.Marshal(rawData)
			if err != nil {
				workersLogger.Error("Failed to serialize raw data", zap.Error(err))
				return
			}

			serializedProcessedData, err := json.Marshal(processedData)
			if err != nil {
				workersLogger.Error("Failed to serialize processed data", zap.Error(err))
				return
			}

			rp := payload.Payload{
				Message:          serializedRawData,
				MessageTimestamp: rawData.Timestamp,
			}

			pp := payload.Payload{
				Message:          serializedProcessedData,
				MessageTimestamp: processedData.Timestamp,
			}

			serializedRp, err := rp.Serialize()
			if err != nil {
				workersLogger.Error("Failed to serialize raw payload", zap.Error(err))
				return
			}

			serializedPp, err := pp.Serialize()
			if err != nil {
				workersLogger.Error("Failed to serialize processed payload", zap.Error(err))
				return
			}

			// Send the processed data to the Kafka producer
			err = e.kafkaProducerPool.SendMessage(e.ctx, "rubicon_kafka_influxdb", serializedRp)
			if err != nil {
				kafkaProducerLogger.Error("Failed to send raw data to Kafka", zap.Error(err))
				return
			}

			err = e.kafkaProducerPool.SendMessage(e.ctx, "rubicon_kafka_influxdb", serializedPp)
			if err != nil {
				kafkaProducerLogger.Error("Failed to send processed data to Kafka", zap.Error(err))
				return
			}

			err = e.kafkaProducerPool.SendMessage(e.ctx, "rubicon_kafka_kodelabs", serializedPp)
			if err != nil {
				kafkaProducerLogger.Error("Failed to send processed data to Kafka", zap.Error(err))
				return
			}
		}
	}
}
