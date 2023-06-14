//package main
//
//import (
//	"context"
//	"fmt"
//	"k8s.io/apimachinery/pkg/fields"
//	"time"
//
//	corev1 "k8s.io/api/core/v1"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	"k8s.io/client-go/kubernetes"
//	"k8s.io/client-go/tools/cache"
//	"k8s.io/client-go/tools/clientcmd"
//)
//
//func main() {
//	// 加载 Kubernetes 配置
//	config, err := clientcmd.BuildConfigFromFlags("", "")
//	if err != nil {
//		panic(err)
//	}
//
//	// 创建 Kubernetes 客户端
//	clientset, err := kubernetes.NewForConfig(config)
//	if err != nil {
//		panic(err)
//	}
//	s, _ := fields.ParseSelector("")
//	podsListWatcher := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "pods", corev1.NamespaceAll, s)
//	// 创建 informer，监控 Pod 变化事件
//	informer := cache.NewSharedInformer(podsListWatcher, &corev1.Pod{}, 0)
//	eventHandlers := cache.ResourceEventHandlerFuncs{
//		AddFunc: func(obj interface{}) {
//			pod := obj.(*corev1.Pod)
//			createdTime := pod.CreationTimestamp.Time
//			for {
//				pod, err = clientset.CoreV1().Pods(pod.Namespace).Get(context.Background(), pod.Name, metav1.GetOptions{})
//				if err != nil {
//					break
//				}
//				if pod.Status.Phase != corev1.PodPending {
//					break
//				}
//			}
//			// 计算运行时间
//			endTime := time.Now()
//			for _, condition := range pod.Status.ContainerStatuses {
//				if condition.State.Terminated != nil {
//					endTime = condition.State.Terminated.FinishedAt.Time
//					break
//				}
//			}
//			runningTime := endTime.Sub(createdTime)
//
//			// 打印符合条件的 Pod 信息
//			if runningTime.Seconds() <= 30 {
//				fmt.Printf("Pod Name: %s, Namespace: %s, Running Time: %v, resource: %s\n", pod.Name, pod.Namespace, runningTime, pod.Spec.Containers[0].Resources)
//			}
//		},
//		DeleteFunc: func(obj interface{}) {
//			pod := obj.(*corev1.Pod)
//			fmt.Printf("%s Pod deleted\n", pod.Name)
//		},
//	}
//	informer.AddEventHandler(eventHandlers)
//	// 启动 informer
//	stopCh := make(chan struct{})
//	defer close(stopCh)
//	go informer.Run(stopCh)
//	select {
//	case <-stopCh:
//		// 当收到停止信号，继续往下执行
//	}
//
//}

package main

import (
	"fmt"
	"k8s.io/klog/v2"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
)

type podInfo struct {
	StartTime time.Time
	EndTime   time.Time
}

var num int

func main() {
	// Create a Kubernetes client.
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// Create a shared informer factory.
	informerFactory := informers.NewSharedInformerFactory(clientset, time.Second*30)

	// Create a Pod informer.
	podInformer := informerFactory.Core().V1().Pods()

	// Create a map to store pod information.
	podInfoMap := make(map[string]*podInfo)

	// Add an event handler to the Pod informer.
	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			podInfoMap[string(pod.GetUID())] = &podInfo{
				StartTime: time.Now(),
			}
			klog.V(4).Infof("Pod %s/%s created at %s\n", pod.GetNamespace(), pod.GetName(), podInfoMap[string(pod.GetUID())].StartTime.Format(time.RFC3339))
		},
		UpdateFunc: func(old, new interface{}) {
			oldPod := old.(*corev1.Pod)
			newPod := new.(*corev1.Pod)
			if oldPod.ResourceVersion == newPod.ResourceVersion {
				return
			}
			if newPod.Status.Phase == corev1.PodSucceeded || newPod.Status.Phase == corev1.PodFailed {
				if podInfoMap[string(newPod.GetUID())] == nil {
					return
				}
				podInfoMap[string(newPod.GetUID())].EndTime = time.Now()
				klog.V(6).Infof("Pod %s/%s finished at %s, total running time is %f seconds\n", newPod.GetNamespace(), newPod.GetName(), podInfoMap[string(newPod.GetUID())].EndTime.Format(time.RFC3339), podInfoMap[string(newPod.GetUID())].EndTime.Sub(podInfoMap[string(newPod.GetUID())].StartTime).Seconds())
				delete(podInfoMap, string(newPod.GetUID()))
			} else if newPod.Status.Phase == corev1.PodRunning {
				if podInfoMap[string(newPod.GetUID())] == nil {
					return
				}
				klog.V(6).Infof("Pod %s/%s updated at %s, current running time is %f seconds\n", newPod.GetNamespace(), newPod.GetName(), time.Now().Format(time.RFC3339), time.Now().Sub(podInfoMap[string(newPod.GetUID())].StartTime).Seconds())
			}
		},
		DeleteFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			if podInfoMap[string(pod.GetUID())] == nil {
				return
			}
			podInfoMap[string(pod.GetUID())].EndTime = time.Now()
			totalTime := podInfoMap[string(pod.GetUID())].EndTime.Sub(podInfoMap[string(pod.GetUID())].StartTime).Seconds()
			klog.V(4).Infof("Pod %s/%s deleted at %s, total running time is %f seconds\n", pod.GetNamespace(), pod.GetName(), podInfoMap[string(pod.GetUID())].EndTime.Format(time.RFC3339), totalTime)
			if totalTime <= 30 {
				cpuLimit := int64(0)
				memoryLimit := int64(0)
				cpuRequest := int64(0)
				memoryRequest := int64(0)
				containers := pod.Spec.Containers
				for _, container := range containers {
					resourceLimits := container.Resources.Limits
					resourceRequests := container.Resources.Requests
					cpuLimit += resourceLimits.Cpu().MilliValue()
					memoryLimit += resourceLimits.Memory().Value()
					cpuRequest += resourceRequests.Cpu().MilliValue()
					memoryRequest += resourceRequests.Memory().Value()

				}
				num++
				fmt.Printf("%d: Pod %s/%s total running %f seconds, cpuLimit:%dm, memoryLimit:%db, cpuRequest:%dm, memoryRequest: %db\n", num, pod.GetNamespace(), pod.GetName(), totalTime, cpuLimit, memoryLimit, cpuRequest, memoryRequest)
			}
			delete(podInfoMap, string(pod.GetUID()))
		},
	})

	// Start the informers.
	informerFactory.Start(nil)
	informerFactory.WaitForCacheSync(nil)

	// Run forever.
	<-make(chan struct{})
}
