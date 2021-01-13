package main

import (
	"fmt"
	"log"
	"os"
	"pkg/k8sDiscovery"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/sirupsen/logrus"
	"github.com/zhiminwen/quote"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/resource"
)

func main() {
	namespace := os.Getenv("K8S_NAMESPACE")
	mode := os.Getenv("MODE")

	clientSet, _, err := k8sDiscovery.K8s()
	if err != nil {
		logrus.Fatalf("Failed to connect to K8s:%v", err)
	}

	pods, err := clientSet.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		logrus.Fatalf("Failed to connect to pods:%v", err)
	}

	switch mode {
	case "checkmk":
		var nsReqCPU int64 = 0
		var nsLimCPU int64 = 0
		var nsReqMem int64 = 0
		var nsLimMem int64 = 0
		var nsReqFS int64 = 0
		var nsLimFS int64 = 0
		for _, p := range pods.Items {
			for _, c := range p.Spec.Containers {
				nsReqCPU += c.Resources.Requests.Cpu().MilliValue()
				nsLimCPU += c.Resources.Limits.Cpu().MilliValue()
				nsReqMem += c.Resources.Requests.Memory().Value()
				nsLimMem += c.Resources.Limits.Memory().Value()
				nsReqFS += c.Resources.Requests.StorageEphemeral().Value()
				nsLimFS += c.Resources.Limits.StorageEphemeral().Value()
			}
		}
		if nsLimCPU == 0 {
			fmt.Printf("CPU load: %v\n", resource.NewMilliQuantity(nsReqCPU, "DecimalSI"))
		} else {
			fmt.Printf("CPU load: %v%%\n", nsReqCPU/nsLimCPU)
		}
		if nsLimMem == 0 {
			fmt.Printf("Memory load: %v\n", resource.NewQuantity(nsReqMem, "BinarySI"))
		} else {
			fmt.Printf("Memory load: %v%%\n", nsReqMem/nsLimMem)
		}
		if nsLimFS == 0 {
			fmt.Printf("FS load: %v\n", resource.NewQuantity(nsReqFS, "BinarySI"))
		} else {
			fmt.Printf("FS load: %v%%\n", nsReqFS/nsLimFS)
		}
		return
	}

	f := excelize.NewFile()
	f.SetActiveSheet(f.NewSheet("Sheet1"))

	header := quote.Word(`Namespace Pod Node Container Request.Cpu Request.Cpu(Canonical) Request.Mem Request.Mem(Canonical) Limits.Cpu Limits.Cpu(Canonical) Limits.Mem Limits.Mem(Canonical) `)
	err = f.SetSheetRow("Sheet1", "A2", &header)
	if err != nil {
		logrus.Fatalf("Failed to save title row:%v", err)
	}
	err = f.AutoFilter("Sheet1", "A2", "L2", "")
	if err != nil {
		logrus.Fatalf("Failed to set auto filter on title row:%v", err)
	}

	row := 3
	for _, p := range pods.Items {
		for _, c := range p.Spec.Containers {
			reqCPU := c.Resources.Requests.Cpu()
			reqMem := c.Resources.Requests.Memory()
			limCPU := c.Resources.Limits.Cpu()
			limMem := c.Resources.Limits.Memory()

			cellName, err := excelize.CoordinatesToCellName(1, row)
			if err != nil {
				log.Fatalf("Could not get cell name from row: %v", err)
			}
			err = f.SetSheetRow("Sheet1", cellName,
				&[]interface{}{
					p.Namespace,
					p.Name,
					p.Status.HostIP,
					c.Name,
					reqCPU.MilliValue(), reqCPU,
					reqMem.Value(), reqMem,
					limCPU.MilliValue(), limCPU,
					limMem.Value(), limMem,
				})
			if err != nil {
				logrus.Fatalf("Failed to save for pod:%v", p.Name)
			}
			row = row + 1

			// logrus.Infof("save as %s", cellName)
		}
	}

	f.SetCellFormula("Sheet1", "E1", fmt.Sprintf(`subtotal(109, E3:E%d)/1000`, row))           //cpu
	f.SetCellFormula("Sheet1", "G1", fmt.Sprintf(`subtotal(109, G3:G%d)/1024/1024/1024`, row)) // mem
	f.SetCellFormula("Sheet1", "I1", fmt.Sprintf(`subtotal(109, I3:I%d)/1000`, row))
	f.SetCellFormula("Sheet1", "K1", fmt.Sprintf(`subtotal(109, K3:K%d)/1024/1024/1024`, row))

	if err = f.SaveAs("resource.xlsx"); err != nil {
		logrus.Fatalf("Failed to save as xlsx:%v", err)
	}
}
