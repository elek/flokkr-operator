package stub

import (
	"context"

	"github.com/flokkr/flokkr-operator/pkg/apis/flokkr/v1alpha1"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
	"k8s.io/helm/pkg/helm"
	"fmt"
	"os/exec"
	"strings"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"bytes"
)

func NewHandler() sdk.Handler {
	host := helm.Host("localhost:44134")
	client := helm.NewClient(host)
	return &Handler{HelmClient: client}
}

type Handler struct {
	HelmClient *helm.Client
}

var lastResourceVersion string

var lastValues = make(map[string][]byte)

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *v1alpha1.Component:
		if event.Deleted {
			logrus.Info("Resource is deleted")
			_, err := executeAndReturn(fmt.Sprintf("helm delete --purge %s", o.Name))
			if err != nil {
				return err
			}
			lastValues[o.Name] = []byte("deleted")
			logrus.Infof("Helm delete is executed successfully")
		} else {
			if o.GetResourceVersion() == lastResourceVersion {
				logrus.Infof("skipping %s because resource version has not changed", strings.Join([]string{o.GetNamespace(), o.GetName()}, "/"))
				return nil
			}

			logrus.Infof("New resource is detected %s/%s", o.Namespace, o.Name)
			chartName := "flokkr/" + o.Spec.Type

			out, err := yaml.Marshal(&o.Spec.Values)
			if err != nil {
				return err
			}
			if val, ok := lastValues[o.Name]; !ok || !bytes.Equal(out, val){
				lastValues[o.Name] = out
				ioutil.WriteFile("/tmp/values.yaml", out, 0644)
				helmCommand := fmt.Sprintf("helm upgrade --values /tmp/values.yaml --install %s --namespace %s %s", o.Name, o.Namespace, chartName)
				_, err := executeAndReturn(helmCommand)
				if err != nil {
					return err
				}
				logrus.Infof("Helm install is executed successfully: %s", helmCommand)
			}

		}
		lastResourceVersion = o.GetResourceVersion()

	}

	return nil
}

func executeAndReturn(command string) (string, error) {
	out, err := exec.Command("bash", "-c", command).CombinedOutput()
	logrus.Infof("Helm output %s", out)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
