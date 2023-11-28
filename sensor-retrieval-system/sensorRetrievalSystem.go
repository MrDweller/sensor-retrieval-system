package sensorretrievalsystem

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	orchestrator_models "github.com/MrDweller/orchestrator-connection/models"
	"github.com/MrDweller/orchestrator-connection/orchestrator"
	"github.com/MrDweller/service-registry-connection/models"

	serviceregistry "github.com/MrDweller/service-registry-connection/service-registry"
)

type SensorRetrievalSystem struct {
	models.SystemDefinition
	ServiceRegistryConnection serviceregistry.ServiceRegistryConnection
	OrchestrationConnection   orchestrator.OrchestratorConnection
}

func NewSensorRetrievalSystem(address string, port int, systemName string, serviceRegistryAddress string, serviceRegistryPort int) (*SensorRetrievalSystem, error) {
	systemDefinition := models.SystemDefinition{
		Address:    address,
		Port:       port,
		SystemName: systemName,
	}

	serviceRegistryConnection, err := serviceregistry.NewConnection(serviceregistry.ServiceRegistry{
		Address: serviceRegistryAddress,
		Port:    serviceRegistryPort,
	}, serviceregistry.SERVICE_REGISTRY_ARROWHEAD_4_6_1, models.CertificateInfo{
		CertFilePath: os.Getenv("CERT_FILE_PATH"),
		KeyFilePath:  os.Getenv("KEY_FILE_PATH"),
		Truststore:   os.Getenv("TRUSTSTORE_FILE_PATH"),
	})
	if err != nil {
		return nil, err
	}

	serviceQueryResult, err := serviceRegistryConnection.Query(models.ServiceDefinition{
		ServiceDefinition: "orchestration-service",
	})
	if err != nil {
		return nil, err
	}

	serviceQueryData := serviceQueryResult.ServiceQueryData[0]

	orchestrationConnection, err := orchestrator.NewConnection(orchestrator.Orchestrator{
		Address: serviceQueryData.Provider.Address,
		Port:    serviceQueryData.Provider.Port,
	}, orchestrator.ORCHESTRATION_ARROWHEAD_4_6_1, orchestrator_models.CertificateInfo{
		CertFilePath: os.Getenv("CERT_FILE_PATH"),
		KeyFilePath:  os.Getenv("KEY_FILE_PATH"),
		Truststore:   os.Getenv("TRUSTSTORE_FILE_PATH"),
	})
	if err != nil {
		return nil, err
	}

	return &SensorRetrievalSystem{
		SystemDefinition:          systemDefinition,
		ServiceRegistryConnection: serviceRegistryConnection,
		OrchestrationConnection:   orchestrationConnection,
	}, nil
}

func (sensorRetrievalSystem *SensorRetrievalSystem) StartSensorRetrievalSystem() error {
	_, err := sensorRetrievalSystem.ServiceRegistryConnection.RegisterSystem(sensorRetrievalSystem.SystemDefinition)
	if err != nil {
		return err
	}
	return nil
}

func (sensorRetrievalSystem *SensorRetrievalSystem) StopSensorRetrievalSystem() error {
	err := sensorRetrievalSystem.ServiceRegistryConnection.UnRegisterSystem(sensorRetrievalSystem.SystemDefinition)
	if err != nil {
		return err
	}
	return nil
}

func (sensorRetrievalSystem *SensorRetrievalSystem) GetSensorData(requestedService orchestrator_models.ServiceDefinition) ([]byte, error) {
	orchestrationResponse, err := sensorRetrievalSystem.OrchestrationConnection.Orchestration(requestedService, orchestrator_models.SystemDefinition{
		Address:    sensorRetrievalSystem.Address,
		Port:       sensorRetrievalSystem.Port,
		SystemName: sensorRetrievalSystem.SystemName,
	})
	if err != nil {
		return nil, err
	}

	if len(orchestrationResponse.Response) <= 0 {
		return nil, errors.New("found no providers")
	}
	provider := orchestrationResponse.Response[0]

	req, err := http.NewRequest("GET", "https://"+provider.Provider.Address+":"+strconv.Itoa(provider.Provider.Port)+provider.ServiceUri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client, err := sensorRetrievalSystem.getClient()
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		errorString := fmt.Sprintf("status: %s, body: %s", resp.Status, string(body))
		return nil, errors.New(errorString)
	}

	return body, nil

}

func (sensorRetrievalSystem *SensorRetrievalSystem) getClient() (*http.Client, error) {
	cert, err := tls.LoadX509KeyPair(os.Getenv("CERT_FILE_PATH"), os.Getenv("KEY_FILE_PATH"))
	if err != nil {
		return nil, err
	}

	// Load truststore.p12
	truststoreData, err := os.ReadFile(os.Getenv("TRUSTSTORE_FILE_PATH"))
	if err != nil {
		return nil, err

	}

	// Extract the root certificate(s) from the truststore
	pool := x509.NewCertPool()
	if ok := pool.AppendCertsFromPEM(truststoreData); !ok {
		return nil, err
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates:       []tls.Certificate{cert},
				RootCAs:            pool,
				InsecureSkipVerify: false,
			},
		},
	}
	return client, nil
}
