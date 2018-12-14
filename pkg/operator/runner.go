package flokkroperator

import (
	"fmt"
	"github.com/spotahome/kooper/log"
	"github.com/spotahome/kooper/operator"
	"github.com/spotahome/kooper/operator/controller"
	"os"
	"time"
)

func Run() {

	ownCli, axeCli, k8sCli, err := createKubernetesClients();
	if err != nil {
		panic(err)
	}
	std := &log.Std{}

	// Create crd.
	crd := newComponentCRD(ownCli, axeCli)

	installHandler := newJobHandler(k8sCli)

	// Create handler.
	handler := newComponentHandler(&installHandler)

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

