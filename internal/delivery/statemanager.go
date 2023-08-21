package delivery

import (
	"fmt"
	"log"
	"net/http"

	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/delivery/handler"
	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/internal/usecase"
)

type stateManagerServer struct {
	Port                int
	StateManagerUseCase usecase.StateManagerUseCase
}

func StateManager(port int, stateManagerUseCase usecase.StateManagerUseCase) *stateManagerServer {
	return &stateManagerServer{
		Port:                port,
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

	log.Printf("Listening on port %d\n", s.Port)
	return http.ListenAndServe(fmt.Sprintf(":%d", s.Port), mux)
}
