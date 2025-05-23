package main

import (
	"errors"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"log"
	"path/filepath"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	// 获取 Kubernetes 配置文件路径
	home := homedir.HomeDir()
	kubeconfig := filepath.Join(home, ".kube", "config")

	// 使用 Kubernetes 配置文件创建 Kubernetes 客户端
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	// 创建 Kubernetes 客户端
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	// 创建 Informer 工厂
	informerFactory := informers.NewSharedInformerFactory(clientset, 0)

	// 创建一个新的 Pod Informer
	podInformer := informerFactory.Core().V1().Pods().Informer()

	// 停止旧的 Pod Informer
	errors.As()
	podLister := clientset.CoreV1().Pods("namespace") // 根据实际情况选择正确的命名空间
	indexers := cache.NewIndexer(cache.DeletionHandlingMetaNamespaceKeyFunc, cache.CacheOpts{})
	//cache.NewSharedIndexInformer
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return podLister.List(context.TODO(), options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return podLister.Watch(context.TODO(), options)
		},
	}
	sharedInformers := informers.NewSharedInformerFactoryWithOptions(clientset, time.Second*30, informers.WithTweakListOptions(func(opt *metav1.ListOptions) {
		opt.LabelSelector = "create-by=panda"
	}))

	tmp := sharedInformers.Core().V1().Services().Informer()
	sharedInformers.Batch().V1().Jobs().Informer()
	re, _ := sharedInformers.ForResource(schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "deployment",
	})
	re.Informer()
	lister := sharedInformers.Core().V1().Pods().Lister()
	tmp.GetIndexer().AddIndexers()
	tmp.GetIndexer().ByIndex()
	podInformer = sharedInformers.Core().V1().Pods().Informer()
	controller := NewController(podInformer)

	// 定义并启动 Informer 处理函数
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*v1.Pod)
			log.Printf("Pod %s created", pod.Name)
			// 这里可以添加处理逻辑，例如检查是否与 Job 关联等
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			// 这里可以添加处理逻辑
		},
		DeleteFunc: func(obj interface{}) {
			pod := obj.(*v1.Pod)
			log.Printf("Pod %s deleted", pod.Name)
			// 这里可以添加处理逻辑，例如检查是否与 Job 关联等
		},
	})

	// 启动 Informer
	stopCh := make(chan struct{})
	defer close(stopCh)
	informerFactory.Start(stopCh)

	// 等待 Informer 结束
	select {}
}

func NewController(podInformer cache.SharedIndexInformer) {
	eventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// 处理添加事件
			pod := obj.(*v1.Pod)
			//service := obj.(*v1.Endpoints)
			//service.
			fmt.Printf("添加了一个Pod: %s\n", pod.Name)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			// 处理更新事件
			oldPod := oldObj.(*v1.Pod)
			newPod := newObj.(*v1.Pod)
			fmt.Printf("更新了Pod: %s\n", newPod.Name)
		},
		DeleteFunc: func(obj interface{}) {
			// 处理删除事件
			pod := obj.(*v1.Pod)
			fmt.Printf("删除了Pod: %s\n", pod.Name)
		},
	}
	podInformer.AddEventHandler(eventHandler)
}
