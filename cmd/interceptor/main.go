package main

import (
	"github.com/GianOrtiz/k8s-transparent-checkpoint-restore/cmd/interceptor/interceptor"
)

func main() {
	interceptor, err := interceptor.New()
	if err != nil {
		panic(err)
	}

	if err := interceptor.Run(); err != nil {
		panic(err)
	}
}
