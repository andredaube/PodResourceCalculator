package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"pkg/k8sDiscovery"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/sirupsen/logrus"
	"github.com/zhiminwen/quote"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/resource"
)

func main() {
	namespace := os.Getenv("K8S_NAMESPACE")
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	checkmk := flag.Bool("checkmk", false, "Checkmk mode for local checks")
	serviceName := flag.String("service-name", "K8sNsRes_"+namespace, "Service name for Checkmk")
	threshCPUAbsWarn := flag.Int64("thresh-cpu-abs-warn", 0, "Threshold value for CPU load warning alert (millis)")
	threshCPUAbsCrit := flag.Int64("thresh-cpu-abs-crit", 0, "Threshold value for CPU load critical alert (millis)")
	threshCPUPercWarn := flag.Float64("thresh-cpu-perc-warn", 0, "Threshold value for CPU load warning alert - only if limits are specified (%)")
	threshCPUPercCrit := flag.Float64("thresh-cpu-perc-crit", 0, "Threshold value for CPU load critical alert - only if limits are specified (%)")
	threshMemAbsWarn := flag.Int64("thresh-mem-abs-warn", 0, "Threshold value for memory load warning alert (bytes)")
	threshMemAbsCrit := flag.Int64("thresh-mem-abs-crit", 0, "Threshold value for memory load critical alert (bytes)")
	threshMemPercWarn := flag.Float64("thresh-mem-perc-warn", 0, "Threshold value for memory load warning alert - only if limits are specified (%)")
	threshMemPercCrit := flag.Float64("thresh-mem-perc-crit", 0, "Threshold value for memory load critical alert - only if limits are specified (%)")
	threshFSAbsWarn := flag.Int64("thresh-fs-abs-warn", 0, "Threshold value for file system load warning alert (bytes)")
	threshFSAbsCrit := flag.Int64("thresh-fs-abs-crit", 0, "Threshold value for file system load critical alert (bytes)")
	threshFSPercWarn := flag.Float64("thresh-fs-perc-warn", 0, "Threshold value for file system load warning alert - only if limits are specified (%)")
	threshFSPercCrit := flag.Float64("thresh-fs-perc-crit", 0, "Threshold value for file system load critical alert - only if limits are specified (%)")
	flag.Parse()

	/*fmt.Printf("checkmk: %v\n"+
		"service-name: %v\n"+
		"thresh-cpu-abs-warn: %v\n"+
		"thresh-cpu-abs-crit: %v\n"+
		"thresh-cpu-perc-warn: %v\n"+
		"thresh-cpu-perc-crit: %v\n"+
		"thresh-mem-abs-warn: %v\n"+
		"thresh-mem-abs-crit: %v\n"+
		"thresh-mem-perc-warn: %v\n"+
		"thresh-mem-perc-crit: %v\n"+
		"thresh-fs-abs-warn: %v\n"+
		"thresh-fs-abs-crit: %v\n"+
		"thresh-fs-perc-warn: %v\n"+
		"thresh-fs-perc-crit: %v\n",
		*checkmk,
		*serviceName,
		*threshCPUAbsWarn,
		*threshCPUAbsCrit,
		*threshCPUPercWarn,
		*threshCPUPercCrit,
		*threshMemAbsWarn,
		*threshMemAbsCrit,
		*threshMemPercWarn,
		*threshMemPercCrit,
		*threshFSAbsWarn,
		*threshFSAbsCrit,
		*threshFSPercWarn,
		*threshFSPercCrit)
	return*/

	clientSet, _, err := k8sDiscovery.K8s(*kubeconfig)
	if err != nil {
		logrus.Fatalf("Failed to connect to K8s:%v", err)
	}

	pods, err := clientSet.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		logrus.Fatalf("Failed to connect to pods:%v", err)
	}

	if *checkmk {
		var nsReqCPU, nsLimCPU, nsReqMem, nsLimMem, nsReqFS, nsLimFS int64 = 0, 0, 0, 0, 0, 0
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
		statusCPU, statusMem, statusFS := 2, 2, 2
		statusDetailCPU, statusDetailMem, statusDetailFS := "", "", ""
		var nsUsedCPUPerc, nsUsedMemPerc, nsUsedFSPerc float64 = 0, 0, 0
		if nsLimCPU == 0 {
			statusDetailCPU = fmt.Sprintf("CPU load: %v < %v <= (%v)",
				resource.NewMilliQuantity(*threshCPUAbsWarn, "DecimalSI"),
				resource.NewMilliQuantity(*threshCPUAbsCrit, "DecimalSI"),
				resource.NewMilliQuantity(nsReqCPU, "DecimalSI"))
			if nsReqCPU < *threshCPUAbsWarn {
				statusCPU = 0
				statusDetailCPU = fmt.Sprintf("CPU load: (%v) < %v < %v",
					resource.NewMilliQuantity(nsReqCPU, "DecimalSI"),
					resource.NewMilliQuantity(*threshCPUAbsWarn, "DecimalSI"),
					resource.NewMilliQuantity(*threshCPUAbsCrit, "DecimalSI"))
			} else if nsReqCPU < *threshCPUAbsCrit {
				statusCPU = 1
				statusDetailCPU = fmt.Sprintf("CPU load: %v <= (%v) < %v",
					resource.NewMilliQuantity(*threshCPUAbsWarn, "DecimalSI"),
					resource.NewMilliQuantity(nsReqCPU, "DecimalSI"),
					resource.NewMilliQuantity(*threshCPUAbsCrit, "DecimalSI"))
			}
		} else {
			nsUsedCPUPerc = math.Ceil(float64(nsReqCPU)/float64(nsLimCPU)*10000) / 100
			statusDetailCPU = fmt.Sprintf("CPU load: %v%% < %v%% <= (%v%%)",
				*threshCPUPercWarn,
				*threshCPUPercCrit,
				nsUsedCPUPerc)
			if nsUsedCPUPerc < *threshCPUPercWarn {
				statusCPU = 0
				statusDetailCPU = fmt.Sprintf("CPU load: (%v%%) < %v%% < %v%%",
					nsUsedCPUPerc,
					*threshCPUPercWarn,
					*threshCPUPercCrit)
			} else if nsUsedCPUPerc < *threshCPUPercCrit {
				statusCPU = 1
				statusDetailCPU = fmt.Sprintf("CPU load: %v%% <= (%v%%) < %v%%",
					*threshCPUPercWarn,
					nsUsedCPUPerc,
					*threshCPUPercCrit)
			}
		}
		if nsLimMem == 0 {
			statusDetailMem = fmt.Sprintf("Memory load: %v < %v <= (%v)",
				resource.NewQuantity(*threshMemAbsWarn, "BinarySI"),
				resource.NewQuantity(*threshMemAbsCrit, "BinarySI"),
				resource.NewQuantity(nsReqMem, "BinarySI"))
			if nsReqMem < *threshMemAbsWarn {
				statusMem = 0
				statusDetailMem = fmt.Sprintf("Memory load: (%v) < %v < %v",
					resource.NewQuantity(nsReqMem, "BinarySI"),
					resource.NewQuantity(*threshMemAbsWarn, "BinarySI"),
					resource.NewQuantity(*threshMemAbsCrit, "BinarySI"))
			} else if nsReqMem < *threshMemAbsCrit {
				statusMem = 1
				statusDetailMem = fmt.Sprintf("Memory load: %v <= (%v) < %v",
					resource.NewQuantity(*threshMemAbsWarn, "BinarySI"),
					resource.NewQuantity(nsReqMem, "BinarySI"),
					resource.NewQuantity(*threshMemAbsCrit, "BinarySI"))
			}
		} else {
			nsUsedMemPerc = math.Ceil(float64(nsReqMem)/float64(nsLimMem)*10000) / 100
			statusDetailMem = fmt.Sprintf("Memory load: %v%% < %v%% <= (%v%%)",
				*threshMemPercWarn,
				*threshMemPercCrit,
				nsUsedMemPerc)
			if nsUsedMemPerc < *threshMemPercWarn {
				statusMem = 0
				statusDetailMem = fmt.Sprintf("Memory load: (%v%%) < %v%% < %v%%",
					nsUsedMemPerc,
					*threshMemPercWarn,
					*threshMemPercCrit)
			} else if nsUsedMemPerc < *threshMemPercCrit {
				statusMem = 1
				statusDetailMem = fmt.Sprintf("Memory load: %v%% <= (%v%%) < %v%%",
					*threshMemPercWarn,
					nsUsedMemPerc,
					*threshMemPercCrit)
			}
		}
		if nsLimFS == 0 {
			statusDetailFS = fmt.Sprintf("FS load: %v < %v <= (%v)",
				resource.NewQuantity(*threshFSAbsWarn, "BinarySI"),
				resource.NewQuantity(*threshFSAbsCrit, "BinarySI"),
				resource.NewQuantity(nsReqFS, "BinarySI"))
			if nsReqFS < *threshFSAbsWarn {
				statusFS = 0
				statusDetailFS = fmt.Sprintf("FS load: (%v) < %v < %v",
					resource.NewQuantity(nsReqFS, "BinarySI"),
					resource.NewQuantity(*threshFSAbsWarn, "BinarySI"),
					resource.NewQuantity(*threshFSAbsCrit, "BinarySI"))
			} else if nsReqFS < *threshFSAbsCrit {
				statusFS = 1
				statusDetailFS = fmt.Sprintf("FS load: %v <= (%v) < %v",
					resource.NewQuantity(*threshFSAbsWarn, "BinarySI"),
					resource.NewQuantity(nsReqFS, "BinarySI"),
					resource.NewQuantity(*threshFSAbsCrit, "BinarySI"))
			}
		} else {
			nsUsedFSPerc = math.Ceil(float64(nsReqFS)/float64(nsLimFS)*10000) / 100
			statusDetailFS = fmt.Sprintf("FS load: %v%% < %v%% <= (%v%%)",
				*threshFSPercWarn,
				*threshFSPercCrit,
				nsUsedFSPerc)
			if nsUsedFSPerc < *threshFSPercWarn {
				statusFS = 0
				statusDetailFS = fmt.Sprintf("FS load: (%v%%) < %v%% < %v%%",
					nsUsedFSPerc,
					*threshFSPercWarn,
					*threshFSPercCrit)
			} else if nsUsedFSPerc < *threshFSPercCrit {
				statusFS = 1
				statusDetailFS = fmt.Sprintf("FS load: %v%% <= (%v%%) < %v%%",
					*threshFSPercWarn,
					nsUsedFSPerc,
					*threshFSPercCrit)
			}
		}
		status := 2
		if statusCPU == 2 || statusMem == 2 || statusFS == 2 {
		} else if statusCPU == 1 || statusMem == 1 || statusFS == 1 {
			status = 1
		} else {
			status = 0
		}
		statusDetail := fmt.Sprintf("%v, %v, %v", statusDetailCPU, statusDetailMem, statusDetailFS)
		fmt.Printf("%v %v - %v\n", status, *serviceName, statusDetail)
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

	_ = f.SetCellFormula("Sheet1", "E1", fmt.Sprintf(`subtotal(109, E3:E%d)/1000`, row))           //cpu
	_ = f.SetCellFormula("Sheet1", "G1", fmt.Sprintf(`subtotal(109, G3:G%d)/1024/1024/1024`, row)) // mem
	_ = f.SetCellFormula("Sheet1", "I1", fmt.Sprintf(`subtotal(109, I3:I%d)/1000`, row))
	_ = f.SetCellFormula("Sheet1", "K1", fmt.Sprintf(`subtotal(109, K3:K%d)/1024/1024/1024`, row))

	if err = f.SaveAs("resource.xlsx"); err != nil {
		logrus.Fatalf("Failed to save as xlsx:%v", err)
	}
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE")
}
