package checkpoint

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	defaultLog "log"
	"net/http"
	"os"

	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/entity"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type KubernetesCheckpointService struct {
	nodeIP   string
	nodePort int
	client   *http.Client
}

func Kubernetes(nodeIP string, nodePort int) *KubernetesCheckpointService {
	client := getClient()
	return &KubernetesCheckpointService{
		client:   client,
		nodeIP:   nodeIP,
		nodePort: nodePort,
	}
}

func (service *KubernetesCheckpointService) Checkpoint(config *entity.CheckpointConfig) error {
	address := fmt.Sprintf(
		"https://%s:%d/checkpoint/%s/%s/%s",
		service.nodeIP,
		service.nodePort,
		"default",
		config.PodName,
		config.Container.Name,
	)
	defaultLog.Printf("address in use %q", address)
	res, err := service.client.Post(address, "application/json", nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		// TODO: extract content of image and save it in a volume.
		return nil
	}

	content, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("checkpoint failed with err %v", err)
	}

	return fmt.Errorf("checkpoint failed with status code %d and response %s", res.StatusCode, content)
}

func getClient() *http.Client {
	clientCertPrefix := "/var/run/secrets/kubelet-certs"
	clientCert, err := tls.LoadX509KeyPair(
		fmt.Sprintf("%s/client.crt", clientCertPrefix),
		fmt.Sprintf("%s/client.key", clientCertPrefix),
	)
	if err != nil {
		log.Log.Error(err, "could not read client cert key pair")
	}
	certs := x509.NewCertPool()

	pemData, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")
	if err != nil {
		log.Log.Error(err, "could not read ca file")
	}
	certs.AppendCertsFromPEM(pemData)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			RootCAs:            certs,
			Certificates:       []tls.Certificate{clientCert},
		},
	}
	return &http.Client{Transport: tr}
}
