package dseworker

import (
	"fmt"
	"strings"

	"github.com/johandrevandeventer/dse-worker/internal/workers"
	gensetworker "github.com/johandrevandeventer/dse-worker/internal/workers/dse_worker/genset"
	"github.com/johandrevandeventer/kafkaclient/payload"
	"go.uber.org/zap"
)

const (
	DSETopicPrefix = "Rubicon/DSE/"
	GatewayGenset  = "genset"
	Worker         = "DSE"
)

// Worker function mapping for gateways
var gatewayWorkers = map[string]func(payload.Payload, *zap.Logger) (*workers.DataStruct, *workers.DataStruct, error){
	GatewayGenset: gensetworker.GensetWorker,
}

func DSEWorker(msg []byte, logger *zap.Logger) (rawDataStruct, processedDataStruct *workers.DataStruct, err error) {
	p, err := payload.Deserialize(msg)
	if err != nil {
		return rawDataStruct, processedDataStruct, fmt.Errorf("failed to deserialize data: %w", err)
	}

	rawDataStruct = &workers.DataStruct{}
	processedDataStruct = &workers.DataStruct{}

	logger.Info(fmt.Sprintf("Running worker -> %s", Worker), zap.String("topic", p.MqttTopic))

	// Trim the worker prefix from the topic
	trimmedTopic := workers.TrimPrefix(p.MqttTopic, DSETopicPrefix)

	logger.Debug("Validating customer", zap.String("topic", trimmedTopic))

	customer, err := workers.GetValidCustomer(trimmedTopic)
	if err != nil {
		return rawDataStruct, processedDataStruct, fmt.Errorf("customer validation failed: %w", err)
	}

	logger.Debug("Validating gateway", zap.String("topic", trimmedTopic))

	gateway, err := workers.GetValidGateway(trimmedTopic)
	if err != nil {
		return rawDataStruct, processedDataStruct, fmt.Errorf("gateway validation failed: %w", err)
	}

	logger.Debug(fmt.Sprintf("%s :: %s :: %s", Worker, customer, gateway))

	// Convert gateway name to lowercase
	gateway = strings.ToLower(gateway)

	// Find the worker function and execute it
	if workerFunc, exists := gatewayWorkers[gateway]; exists {
		rawDataStruct, processedDataStruct, err = workerFunc(*p, logger)
		if err != nil {
			return rawDataStruct, processedDataStruct, fmt.Errorf("failed to process gateway data: %w", err)
		}

		return rawDataStruct, processedDataStruct, nil
	} else {
		return rawDataStruct, processedDataStruct, fmt.Errorf("no decoder function found for gateway: %s", gateway)
	}

}
