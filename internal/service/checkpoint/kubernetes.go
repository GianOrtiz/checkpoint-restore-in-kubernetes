package checkpoint

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/fs"
	defaultLog "log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/entity"
	"github.com/containers/buildah"
	is "github.com/containers/image/v5/storage"
	"github.com/containers/storage"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type KubernetesCheckpointService struct {
	nodeIP         string
	nodePort       int
	client         *http.Client
	buildahBuilder buildah.Builder
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
		if err := service.generateCheckpointImage(config.PodName, config.Container.Name, config.CheckpointHash); err != nil {
			return err
		}

		// TODO: Get the image path as a tar format.
		// TODO: Get the contents of the image.
		// TODO: Generates an OCI compliant image based on the restore.
		// TODO: Annotate the generated image with io.kubernetes.cri-o.annotations.checkpoint.name=<container> to indicate it is a checkpoint.
		// TOOD: Now, the node should have the generated checkpoint to use.
		return nil
	}

	content, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("checkpoint failed with err %v", err)
	}

	return fmt.Errorf("checkpoint failed with status code %d and response %s", res.StatusCode, content)
}

func (service *KubernetesCheckpointService) generateCheckpointImage(podName, containerName, imageHash string) error {
	buildOptions, err := storage.DefaultStoreOptionsAutoDetectUID()
	if err != nil {
		return err
	}

	buildStore, err := storage.GetStore(buildOptions)
	if err != nil {
		return err
	}

	builderOptions := buildah.BuilderOptions{
		FromImage: "scratch",
	}

	builder, err := buildah.NewBuilder(context.TODO(), buildStore, builderOptions)
	if err != nil {
		return err
	}

	checkpointPath := "/var/lib/kubelet/checkpoints"
	checkpointPrefix := fmt.Sprintf("%s_%s-%s-", podName, "default", containerName)
	files, err := os.ReadDir(checkpointPath)
	if err != nil {
		return err
	}

	var fileWithTheLatestTimestamp fs.DirEntry
	var latestTimestamp time.Time
	for _, file := range files {
		if !file.IsDir() {
			regexPattern := fmt.Sprintf("%s*.tar", checkpointPrefix)
			fileMatchContainerCheckpointFilename, _ := regexp.MatchString(regexPattern, file.Name())
			if fileMatchContainerCheckpointFilename {
				timestampRegex := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z)`)
				matches := timestampRegex.FindStringSubmatch(file.Name())
				if len(matches) > 1 {
					timestampStr := matches[1]
					timestamp, err := time.Parse(time.RFC3339, timestampStr)
					if err != nil {
						return err
					}
					if timestamp.After(latestTimestamp) {
						fileWithTheLatestTimestamp = file
						latestTimestamp = timestamp
					}
				}
			}
		}
	}

	checkpointFullPath := fmt.Sprintf("%s/%s", checkpointPath, fileWithTheLatestTimestamp.Name())
	if err := builder.Add("/", false, buildah.AddAndCopyOptions{}, checkpointFullPath); err != nil {
		return err
	}

	builder.SetAnnotation("io.kubernetes.cri-o.annotations.checkpoint.name", containerName)
	imageRef, err := is.Transport.ParseStoreReference(buildStore, "docker.io/gianaortiz/crsc-interceptor")
	if err != nil {
		return err
	}

	_, _, _, err = builder.Commit(context.TODO(), imageRef, buildah.CommitOptions{})
	if err != nil {
		return err
	}

	return nil
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
