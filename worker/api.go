package worker

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type API struct {
	Address string
	Port    int
	Worker  *Worker
	Router  *chi.Mux
}

func (a *API) initRouter() {
	a.Router = chi.NewRouter()
	a.Router.Route("/tasks", func(r chi.Router) {
		r.Post("/", a.StartTaskHandler)
		r.Get("/", a.GetTaskHandler)
		r.Route("/{taskID}", func(r chi.Router) {
			r.Delete("/", a.StopTaskHandler)
		})
	})
	a.Router.Route("/stats", func(r chi.Router) {
		r.Get("/", a.GetStatsHandler)
	})
}

func (a *API) Start() {
	log.Printf("Starting API on %s:%d\n", a.Address, a.Port)
	a.initRouter()
	log.Fatal(
		http.ListenAndServe(
			fmt.Sprintf("%s:%d", a.Address, a.Port),
			a.Router,
		),
	)

}
