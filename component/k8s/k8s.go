package k8s

import (
	"context"
	"errors"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"panda-go/common/log"
	"panda-go/common/utils"
	"sort"
	"sync"

	"k8s.io/client-go/kubernetes"
)

var (
	once                 sync.Once
	k8sClient            *kubernetes.Clientset
	mapLock              sync.Mutex
	service2NamespacesMp = make(map[string][]string, 0)
	config               *rest.Config
)

func GetK8sConfig() (*rest.Config, error) {
	if config == nil {
		return nil, errors.New("not found k8s config")
	}
	return config, nil
}

func GetK8sClient() *kubernetes.Clientset {
	once.Do(func() {
		ctx := context.TODO()
		var err error
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Error(ctx, "rest.InClusterConfig failed ", zap.Error(err))
		} else {
			//tmp, err := discovery.NewDiscoveryClient()
			//_, ll, err := tmp.ServerGroupsAndResources()
			//for _, l := range ll {
			//	version, err := schema.ParseGroupVersion()
			//	version.
			//}
			//index := cache.NewIndexer(cache.MetaNamespaceKeyFunc)
			//index.Add()
			//index.ByIndex()
			//fa := informers.NewSharedInformerFactory()
			//podInformer := fa.Core().V1().Pods()
			//indexer := podInformer.Lister()
			//indexer.List()

			k8sClient, err = kubernetes.NewForConfig(config)
			if err != nil {
				log.Error(ctx, "kubernetes.NewForConfig failed ", zap.Error(err))
			}
		}
	})
	return k8sClient
}

func GetNewServiceEntity(ctx context.Context, serviceName, originNamespace string) (isChange bool, namespace string, err error) {
	// originNamespace为空时候，在nacos没有查询到或者配置为空，启用自动发现服务
	sortNamespacesFunc := func(namespaces []string, headNamespace string) {
		sort.Slice(namespaces, func(i, j int) bool {
			f := func(specialNamespace string) bool {
				if namespaces[i] == specialNamespace {
					return true
				} else if namespaces[j] == specialNamespace {
					return false
				} else {
					return namespaces[i] < namespaces[j]
				}
			}
			if headNamespace != "" {
				return f(headNamespace)
			}
			return f("default")
		})
	}
	mapLock.Lock()
	defer mapLock.Unlock()
	if namespaces, ok := service2NamespacesMp[serviceName]; ok && len(namespaces) > 0 {
		if !utils.InArray(originNamespace, namespaces) {
			return true, namespaces[0], nil
		}
	}
	clientset := GetK8sClient()
	if clientset != nil {
		services, err := clientset.CoreV1().Services("").List(ctx, metav1.ListOptions{})
		if err != nil {
			log.Error(ctx, "[GetNewServiceEntity] [CoreV1().Services(\"\").List] call to k8s api fail", zap.Error(err))
			return false, originNamespace, err
		}
		if services != nil {
			service2NamespacesMp[serviceName] = []string{}
			for _, service := range services.Items {
				singleServiceName := service.GetName()
				if singleServiceName != serviceName {
					continue
				}
				service2NamespacesMp[singleServiceName] = append(service2NamespacesMp[singleServiceName], service.GetNamespace())
			}
			if namespaces, ok := service2NamespacesMp[serviceName]; ok {
				var headNamespace string
				if utils.InArray(originNamespace, namespaces) {
					headNamespace = originNamespace
					isChange = false
				} else {
					isChange = true
				}
				sortNamespacesFunc(namespaces, headNamespace)
				return isChange, namespaces[0], nil
			} else {
				log.Error(ctx, "[GetNewServiceEntity] serviceName not found in k8s", zap.Error(errors.New("services is nil")))
				return false, originNamespace, errors.New("serviceName not found in k8s")
			}
		} else {
			log.Error(ctx, "[GetNewServiceEntity] call to k8s api fail", zap.Error(errors.New("services is nil")))
			return false, originNamespace, err
		}
	}
	err = errors.New("[GetNewServiceEntity] [GetK8sClient] is nil in function GetService")
	return
}
