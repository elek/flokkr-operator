package flokkroperator

import (
	"fmt"
	ownapi "github.com/flokkr/flokkr-operator/pkg/api/flokkr/v1alpha1"
	owncli "github.com/flokkr/flokkr-operator/pkg/clientset/versioned/typed/flokkr/v1alpha1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionscli "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"time"

	"errors"
	"github.com/sirupsen/logrus"
)

type ComponentCrd struct {
	crdCli *owncli.FlokkrV1alpha1Client
	axeCli *apiextensionscli.Clientset
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

			return crd.crdCli.Components("").List(options);
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return crd.crdCli.Components("").Watch(options);
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
				Singular: ownapi.ComponentName,
				Plural:   ownapi.ComponentNamePlural,
				Kind:     ownapi.ComponentKind,
				ListKind: ownapi.ComponentListKind,
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
