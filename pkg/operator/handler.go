package flokkroperator

import (
	"bytes"
	"context"
	ownapi "github.com/flokkr/flokkr-operator/pkg/api/flokkr/v1alpha1"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/runtime"
)

type ComponentHandler struct {
	handler    *JobHander
	lastValues map[string][]byte
}

func newComponentHandler(handler *JobHander) *ComponentHandler {
	return &ComponentHandler{
		handler:    handler,
		lastValues: make(map[string][]byte),
	}
}

func (h *ComponentHandler) Add(_ context.Context, obj runtime.Object) error {
	component := obj.(*ownapi.Component)
	out, err := yaml.Marshal(&component.Spec.Values)
	if err != nil {
		return err
	}
	if val, ok := h.lastValues[component.Name]; !ok || !bytes.Equal(out, val) {
		h.lastValues[component.Name] = out
		return h.handler.Install(component)
	} else {
		return nil
	}

}
func (h *ComponentHandler) Delete(_ context.Context, name string) error {
	err := h.handler.Delete(name)
	h.lastValues[name] = []byte("deleted")
	return err;
}
