package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hellofresh/janus/pkg/errors"
	"github.com/hellofresh/janus/pkg/render"
	"github.com/hellofresh/janus/pkg/router"
	"github.com/hellofresh/janus/pkg/types"
)

// Controller is the api rest controller
type Controller struct {
	CurrentConfigurations *types.Configuration
}

// NewController creates a new instance of Controller
func NewController(currentConfigurations *types.Configuration) *Controller {
	return &Controller{currentConfigurations}
}

// Get is the find all handler
func (c *Controller) Get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, http.StatusOK, c.CurrentConfigurations.Backends)
	}
}

// GetBy is the find by handler
func (c *Controller) GetBy() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := router.URLParam(r, "name")
		backend := c.findByName(name)

		if backend != nil {
			render.JSON(w, http.StatusOK, backend)
		} else {
			render.JSON(w, http.StatusNotFound, "api definition not found")
		}
	}
}

// PutBy is the update handler
func (c *Controller) PutBy(changeChan chan<- types.ConfigurationEvent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := router.URLParam(r, "name")
		backend := c.findByName(name)

		if backend == nil {
			render.JSON(w, http.StatusNotFound, "api definition not found")
			return
		}

		err := json.NewDecoder(r.Body).Decode(backend)
		if err != nil {
			errors.Handler(w, err)
			return
		}

		// avoid situation when trying to update existing definition with new path
		// that is already registered with another name
		existingPathDefinition := c.findByListenPath(backend.Proxy.ListenPath)

		if existingPathDefinition == nil {
			render.JSON(w, http.StatusConflict, "api listen path is already registered")
			return
		}

		changeChan <- types.ConfigurationEvent{
			Type:    types.ConfigurationChanged,
			Backend: backend,
		}

		w.WriteHeader(http.StatusOK)
	}
}

// Post is the create handler
func (c *Controller) Post(changeChan chan<- types.ConfigurationEvent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		backend := types.NewBackend()

		err := json.NewDecoder(r.Body).Decode(backend)
		if nil != err {
			errors.Handler(w, err)
			return
		}

		exists := c.exists(backend)

		if exists {
			render.JSON(w, http.StatusConflict, "api name is already registered")
			return
		}

		changeChan <- types.ConfigurationEvent{
			Type:    types.ConfigurationChanged,
			Backend: backend,
		}

		w.Header().Add("Location", fmt.Sprintf("/apis/%s", backend.Name))
		w.WriteHeader(http.StatusCreated)
	}
}

// DeleteBy is the delete handler
func (c *Controller) DeleteBy(changeChan chan<- types.ConfigurationEvent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := router.URLParam(r, "name")

		changeChan <- types.ConfigurationEvent{
			Type:    types.ConfigurationChanged,
			Backend: &types.Backend{Name: name},
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (c *Controller) findByName(name string) *types.Backend {
	if name == "" {
		return nil
	}

	for _, backend := range c.CurrentConfigurations.Backends {
		if backend.Name == name {
			return backend
		}
	}

	return nil
}

func (c *Controller) findByListenPath(path string) *types.Backend {
	if path == "" {
		return nil
	}

	for _, backend := range c.CurrentConfigurations.Backends {
		if backend.Proxy.ListenPath == path {
			return backend
		}
	}

	return nil
}

func (c *Controller) exists(backend *types.Backend) bool {
	foundedBackend := c.findByName(backend.Name)
	if foundedBackend != nil {
		return true
	}

	foundedBackend = c.findByListenPath(backend.Proxy.ListenPath)
	if foundedBackend != nil {
		return true
	}

	return false
}
