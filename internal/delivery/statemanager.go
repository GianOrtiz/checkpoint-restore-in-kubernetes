package delivery

import (
	"fmt"
	"log"
	"net/http"

	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/delivery/handler"
	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/entity"
	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/usecase"
)

type stateManagerServer struct {
	Port                int
	Config              entity.StateManagerConfig
	StateManagerUseCase usecase.StateManagerUseCase
}

func StateManager(port int, stateManagerUseCase usecase.StateManagerUseCase, stateManagerConfig entity.StateManagerConfig) *stateManagerServer {
	return &stateManagerServer{
		Port:                port,
		Config:              stateManagerConfig,
		StateManagerUseCase: stateManagerUseCase,
	}
}

func (s *stateManagerServer) Run() error {
	mux := http.NewServeMux()

	saveImageMetadataHandler := handler.SaveImageMetadata(s.StateManagerUseCase)
	getImageMetdataHandler := handler.GetImageMetadata(s.StateManagerUseCase)

	mux.HandleFunc("/containers/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			saveImageMetadataHandler.ServeHTTP(w, r)
		} else if r.Method == http.MethodGet {
			getImageMetdataHandler.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	if s.Config.DevelopmentFeaturesEnabled {
		mux.HandleFunc("/checkpoint", func(w http.ResponseWriter, r *http.Request) {
			containerName := r.URL.Query().Get("name")
			containerHash := r.URL.Query().Get("hash")
			if err := s.StateManagerUseCase.DevelopmentRestore(containerName, containerHash); err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
			}
		})
	}

	log.Printf("Listening on port %d\n", s.Port)
	return http.ListenAndServe(fmt.Sprintf(":%d", s.Port), mux)
}
