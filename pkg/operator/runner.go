package flokkroperator

import (
	"github.com/spotahome/kooper/operator"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"github.com/spotahome/kooper/log"
	"github.com/spotahome/kooper/operator/controller"
	"k8s.io/client-go/tools/cache"
	"k8s.io/apimachinery/pkg/runtime"
	"time"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiextensionscli "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"context"
	"fmt"
	"os"
	"k8s.io/apimachinery/pkg/watch"
	owncli "github.com/flokkr/flokkr-operator/pkg/clientset/versioned/typed/flokkr/v1alpha1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	ownapi "github.com/flokkr/flokkr-operator/pkg/api/flokkr/v1alpha1"

	"github.com/sirupsen/logrus"
	"errors"
)

func Run() {

	ownCli, axeCli, _, err := createKubernetesClients();
	if err != nil {
		panic(err)
	}
	std := &log.Std{}
	// Create crd.
	crd := newComponentCRD(ownCli, axeCli)

	helmHandler := newHelmHandler()

	// Create handler.
	handler := newComponentHandler(&helmHandler)

	ctrl := controller.NewSequential(30*time.Second, handler, crd, nil, std)
	signalC := make(chan os.Signal, 1)
	stopC := make(chan struct{})
	finishC := make(chan error)

	operator := operator.NewOperator(crd, ctrl, std)

	// Run in background the operator.
	go func() {
		finishC <- operator.Run(stopC)
	}()

	select {
	case err := <-finishC:
		if err != nil {
			fmt.Fprintf(os.Stderr, "error running operator: %s", err)
			os.Exit(1)
		}
	case <-signalC:
		fmt.Println("Signal captured, exiting...")
	}
	close(stopC)
	time.Sleep(5 * time.Second)

}

type ComponentCrd struct {
	crdCli *owncli.FlokkrV1alpha1Client
	axeCli *apiextensionscli.Clientset
}

type ComponentHandler struct {
	handler *HelmHandler
}

func newComponentHandler(handler *HelmHandler) *ComponentHandler {
	return &ComponentHandler{
		handler: handler,
	}
}

func (h *ComponentHandler) Add(_ context.Context, obj runtime.Object) error {
	logrus.Infof("resource is added")
	component := obj.(*ownapi.Component)
	return h.handler.Install(component)

}
func (h *ComponentHandler) Delete(_ context.Context, name string) error {
	logrus.Infof("resource is deleted")
	return h.handler.Delete(name)
}

func newComponentCRD(crdCli *owncli.FlokkrV1alpha1Client, axeCli *apiextensionscli.Clientset) *ComponentCrd {
	return &ComponentCrd{
		crdCli: crdCli,
		axeCli: axeCli,
	}
}

func (crd *ComponentCrd) GetListerWatcher() cache.ListerWatcher {
	return &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {

			return crd.crdCli.Components("ozoneweekly").List(options);
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return crd.crdCli.Components("ozoneweekly").Watch(options);
		},
	}
}
func (*ComponentCrd) GetObject() runtime.Object {
	return &ownapi.Component{}
}

func (crd *ComponentCrd) Initialize() error {

	name := ownapi.ComponentNamePlural + "." + ownapi.SchemeGroupVersion.Group
	config := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   ownapi.SchemeGroupVersion.Group,
			Version: ownapi.SchemeGroupVersion.Version,
			Scope:   ownapi.ComponentScope,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural: ownapi.ComponentNamePlural,
				Kind:   ownapi.ComponentKind,
			},
		},
	}

	_, err := crd.axeCli.ApiextensionsV1beta1().CustomResourceDefinitions().Create(config)
	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return fmt.Errorf("error creating crd %s: %s", name, err)
		}
		return nil
	}

	logrus.Infof("crd %s created, waiting to be ready...", name)
	if err := crd.WaitToBePresent(name, 3*time.Minute); err != nil {
		return err
	}
	logrus.Infof("crd %s ready", name)

	return nil
}

func (c *ComponentCrd) WaitToBePresent(name string, timeout time.Duration) error {

	tout := time.After(timeout)
	t := time.NewTicker(2 * time.Second)

	for {
		select {
		case <-t.C:
			_, err := c.axeCli.ApiextensionsV1beta1().CustomResourceDefinitions().Get(name, metav1.GetOptions{})
			// Is present, finish.
			if err == nil {
				return nil
			}
		case <-tout:
			return errors.New("timeout waiting for CRD")
		}
	}
}
