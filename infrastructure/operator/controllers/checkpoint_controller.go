/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"io/fs"
	defaultLog "log"
	"net/http"
	"os"
	"regexp"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/infrastructure/operator/api/v1alpha1"
	"github.com/containers/buildah"
	"github.com/containers/common/pkg/config"
	is "github.com/containers/image/v5/storage"
	"github.com/containers/storage"
)

// CheckpointReconciler reconciles a Checkpoint object
type CheckpointReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps.crsc.io,resources=checkpoints,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.crsc.io,resources=checkpoints/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.crsc.io,resources=checkpoints/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *CheckpointReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	checkpointName := req.NamespacedName
	checkpoint := v1alpha1.Checkpoint{}
	if err := r.Get(ctx, checkpointName, &checkpoint); err != nil {
		return ctrl.Result{Requeue: false}, err
	}

	ticker := time.NewTicker(time.Duration(checkpoint.Spec.Interval * 1000000000))
	go func(ticker *time.Ticker) {
		for range ticker.C {
			// Call http to make checkpoint
			url := fmt.Sprintf("http://%s:8002/checkpoint", checkpoint.Spec.PodIP)
			res, err := http.Get(url)
			if err != nil {
				logger.Error(err, "failed to make checkpoint")
				continue
			}
			if res.StatusCode != http.StatusOK {
				logger.Error(fmt.Errorf("checkpoint returned status %d", res.StatusCode), "")
				continue
			}
			if err := r.generateCheckpointImage(checkpoint.Spec.ContainerName); err != nil {
				logger.Error(err, "failed to generate checkpoint image")
			}
		}
	}(ticker)

	return ctrl.Result{}, nil
}

func (r *CheckpointReconciler) generateCheckpointImage(containerName string) error {
	buildOptions, err := storage.DefaultStoreOptionsAutoDetectUID()
	if err != nil {
		return err
	}

	conf, err := config.Default()
	if err != nil {
		panic(err)
	}
	capabilitiesForRoot, err := conf.Capabilities("root", nil, nil)
	if err != nil {
		panic(err)
	}

	buildStore, err := storage.GetStore(buildOptions)
	if err != nil {
		return err
	}

	builderOptions := buildah.BuilderOptions{
		FromImage:    "scratch",
		Capabilities: capabilitiesForRoot,
	}

	builder, err := buildah.NewBuilder(context.TODO(), buildStore, builderOptions)
	if err != nil {
		return err
	}

	checkpointPath := "/var/lib/kubelet/checkpoints"
	checkpointPrefix := fmt.Sprintf(
		"checkpoint-(%s)-[0-9a-f]+-[0-9a-f]+_(%s)-%s-",
		containerName,
		"default",
		containerName,
	)
	defaultLog.Println("Checkpoint prefix regex", checkpointPrefix)
	files, err := os.ReadDir(checkpointPath)
	if err != nil {
		return err
	}

	var fileWithTheLatestTimestamp fs.DirEntry
	var latestTimestamp time.Time
	for _, file := range files {
		defaultLog.Printf("File: %s", file.Name())
		if !file.IsDir() {
			regexPattern := checkpointPrefix
			fileMatchContainerCheckpointFilename, _ := regexp.MatchString(regexPattern, file.Name())
			defaultLog.Printf("File %s is not directory", file.Name())
			if fileMatchContainerCheckpointFilename {
				defaultLog.Printf("File %s matchs a checkpoint filename", file.Name())
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

	builder.SetAnnotation("io.kubernetes.cri-o.annotations.checkpoint.name", containerName)

	checkpointFullPath := fmt.Sprintf("%s/%s", checkpointPath, fileWithTheLatestTimestamp.Name())
	defaultLog.Printf("This is the file selected %s", checkpointFullPath)
	if err := builder.Add("/", false, buildah.AddAndCopyOptions{}, checkpointFullPath); err != nil {
		defaultLog.Printf("Failed to add to builder: %v", err)
		return err
	}

	imageRef, err := is.Transport.ParseStoreReference(buildStore, "docker.io/gianaortiz/crsc-interceptor")
	if err != nil {
		defaultLog.Printf("Failed to parse store reference: %v", err)
		return err
	}

	_, _, _, err = builder.Commit(context.TODO(), imageRef, buildah.CommitOptions{})
	if err != nil {
		defaultLog.Printf("Failed to commit: %v", err)
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CheckpointReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Checkpoint{}).
		Complete(r)
}
