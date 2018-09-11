package flokkroperator

import (
	"github.com/flokkr/flokkr-operator/pkg/api/flokkr/v1alpha1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/api/core/v1"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math/rand"
	"github.com/sirupsen/logrus"
	"time"
	"strings"
	"fmt"
)

type JobHander struct {
	K8sCli *kubernetes.Clientset
}

func newJobHandler(clientset *kubernetes.Clientset) JobHander {
	return JobHander{
		K8sCli: clientset,
	}
}

func (handler *JobHander) Install(component *v1alpha1.Component) error {
	job := createJob(component.ObjectMeta.Namespace, component.ObjectMeta.Name, "install.sh")
	logrus.Infof("Starting job %s", job.ObjectMeta.Name)
	_, err := handler.K8sCli.BatchV1().Jobs(component.ObjectMeta.Namespace).Create(&job)
	return err;
}

func createJob(namespace string, name string, action string) batchv1.Job {
	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("flokkr-%s-%s-%s", action, name, RandStringRunes(5)),
			Namespace: namespace,
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"flokkr.github.io/task":     "job",
						"flokkr.github.io/instance": name,
					},
				},
				Spec: v1.PodSpec{
					ServiceAccountName: "flokkr-operator",
					Containers: []v1.Container{
						{
							Name:  "installer",
							Image: "flokkr/flokkr-operator",
							Command: []string{
								fmt.Sprintf("/go/bin/%s", action),
								namespace,
								name,
							},
						},
					},
					RestartPolicy: v1.RestartPolicyNever,
				},
			},
		},
	}
	return job
}

func (handler *JobHander) Delete(name string) error {
	parts := strings.Split(name, "/")
	job := createJob(parts[0], parts[1], "delete.sh")
	logrus.Infof("Starting job %s", job.ObjectMeta.Name)
	_, err := handler.K8sCli.BatchV1().Jobs(parts[0]).Create(&job)
	return err;
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
