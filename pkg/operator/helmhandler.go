package flokkroperator

import (
	"os/exec"
	"github.com/Sirupsen/logrus"
	"fmt"
	"github.com/flokkr/flokkr-operator/pkg/api/flokkr/v1alpha1"
	"bytes"
	"io/ioutil"
	"gopkg.in/yaml.v2"
)

type HelmHandler struct {
	lastValues map[string][]byte
}

func newHelmHandler() HelmHandler {
	return HelmHandler{
		lastValues: make(map[string][]byte),
	}
}

func (helm *HelmHandler) Install(component *v1alpha1.Component) error {

	logrus.Infof("New resource is detected %s/%s", component.Namespace, component.Name)
	chartName := "flokkr/" + component.Spec.Type

	out, err := yaml.Marshal(&component.Spec.Values)
	if err != nil {
		return err
	}
	if val, ok := helm.lastValues[component.Name]; !ok || !bytes.Equal(out, val) {
		helm.lastValues[component.Name] = out
		ioutil.WriteFile("/tmp/values.yaml", out, 0644)
		helmCommand := fmt.Sprintf("helm upgrade --values /tmp/values.yaml --install %s --namespace %s %s", component.Name, component.Namespace, chartName)
		_, err := helm.executeAndReturn(helmCommand)
		if err != nil {
			return err
		}
		logrus.Infof("Helm install is executed successfully: %s", helmCommand)
	}
	return nil

}
func (helm *HelmHandler) Delete(name string) error {
	logrus.Info("Resource is deleted")
	_, err := helm.executeAndReturn(fmt.Sprintf("helm delete --purge %s", name))
	if err != nil {
		return err
	}
	helm.lastValues[name] = []byte("deleted")
	logrus.Infof("Helm delete is executed successfully")
	return nil

}
func (*HelmHandler) executeAndReturn(command string) (string, error) {
	logrus.Infof("Executing command %s", command)
	out, err := exec.Command("bash", "-c", command).CombinedOutput()
	logrus.Infof("Helm output %s", out)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
