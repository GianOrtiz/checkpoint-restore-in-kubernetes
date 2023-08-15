package delivery

import (
	"fmt"
	"log"
	"net/http"

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
	mux.HandleFunc("/containers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.StateManagerUseCase.SaveImageMetadata()
		} else if r.Method == http.MethodGet {

		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	log.Printf("Listening on port %d\n", s.Port)
	return http.ListenAndServe(fmt.Sprintf(":%d", s.Port), mux)
}
