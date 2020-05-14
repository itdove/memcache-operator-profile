package memcache_operator_profile_test

import (
	"fmt"
	"os/user"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	NAMESPACE                = "memcached"
	MEMCACHE_DEPLOYMENT_NAME = "example-memcached"
)

var _ = Describe("Memcache", func() {
	var client kubernetes.Interface
	var clientDynamic dynamic.Interface

	BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		SetDefaultEventuallyPollingInterval(10 * time.Second)
		client = newKubeClient(masterUrl)
		clientDynamic = newKubeClientDynamic(masterUrl)
	})

	It("Request 2 replicas", func() {
		klog.Info("Request 2 replicas")
		By("Creating new memcached for 2 replicas", func() {
			gvr := schema.GroupVersionResource{Group: "cache.example.com", Version: "v1alpha1", Resource: "memcacheds"}
			namespace := clientDynamic.Resource(gvr).Namespace(NAMESPACE)
			memcached, err := namespace.Get(MEMCACHE_DEPLOYMENT_NAME, metav1.GetOptions{})
			Expect(err).Should(BeNil())
			if v, ok := memcached.Object["spec"]; ok {
				spec := v.(map[string]interface{})
				spec["size"] = 2
			}
			Expect(namespace.Update(memcached, metav1.UpdateOptions{})).NotTo(BeNil())
			Expect(namespace.Get(MEMCACHE_DEPLOYMENT_NAME, metav1.GetOptions{})).NotTo(BeNil())
		})
		Eventually(func() error {
			d, err := client.AppsV1().Deployments(NAMESPACE).Get(MEMCACHE_DEPLOYMENT_NAME, metav1.GetOptions{})
			if err != nil {
				klog.V(5).Infof("Error: %s", err.Error())
				return err
			}
			if d.Status.AvailableReplicas != 2 {
				err = fmt.Errorf("Not yet 2 replicas available")
				klog.V(5).Infof("Error: %s", err.Error())
				return err
			}
			return nil
		}).Should(BeNil())
		klog.Info("Got 2 replicas")
	})

	It("Request 3 replicas", func() {
		klog.Info("Request 3 replicas")
		By("Creating new memcached for 2 replicas", func() {
			gvr := schema.GroupVersionResource{Group: "cache.example.com", Version: "v1alpha1", Resource: "memcacheds"}
			namespace := clientDynamic.Resource(gvr).Namespace(NAMESPACE)
			memcached, err := namespace.Get(MEMCACHE_DEPLOYMENT_NAME, metav1.GetOptions{})
			Expect(err).Should(BeNil())
			if v, ok := memcached.Object["spec"]; ok {
				spec := v.(map[string]interface{})
				spec["size"] = 3
			}
			Expect(namespace.Update(memcached, metav1.UpdateOptions{})).NotTo(BeNil())
			Expect(namespace.Get(MEMCACHE_DEPLOYMENT_NAME, metav1.GetOptions{})).NotTo(BeNil())
		})
		Eventually(func() error {
			d, err := client.AppsV1().Deployments(NAMESPACE).Get(MEMCACHE_DEPLOYMENT_NAME, metav1.GetOptions{})
			if err != nil {
				klog.V(5).Infof("Error: %s", err.Error())
				return err
			}
			if d.Status.AvailableReplicas != 3 {
				err = fmt.Errorf("Not yet 3 replicas available")
				klog.V(5).Infof("Error: %s", err.Error())
				return err
			}
			return nil
		}).Should(BeNil())
		klog.Info("Got 3 replicas")
	})

	It("Request 2 replicas", func() {
		klog.Info("Request 2 replicas")
		By("Creating new memcached for 2 replicas", func() {
			gvr := schema.GroupVersionResource{Group: "cache.example.com", Version: "v1alpha1", Resource: "memcacheds"}
			namespace := clientDynamic.Resource(gvr).Namespace(NAMESPACE)
			memcached, err := namespace.Get(MEMCACHE_DEPLOYMENT_NAME, metav1.GetOptions{})
			Expect(err).Should(BeNil())
			if v, ok := memcached.Object["spec"]; ok {
				spec := v.(map[string]interface{})
				spec["size"] = 2
			}
			Expect(namespace.Update(memcached, metav1.UpdateOptions{})).NotTo(BeNil())
			Expect(namespace.Get(MEMCACHE_DEPLOYMENT_NAME, metav1.GetOptions{})).NotTo(BeNil())
		})
		Eventually(func() error {
			d, err := client.AppsV1().Deployments(NAMESPACE).Get(MEMCACHE_DEPLOYMENT_NAME, metav1.GetOptions{})
			if err != nil {
				klog.V(5).Infof("Error: %s", err.Error())
				return err
			}
			if d.Status.AvailableReplicas != 2 {
				err = fmt.Errorf("Not yet 2 replicas available")
				klog.V(5).Infof("Error: %s", err.Error())
				return err
			}
			return nil
		}).Should(BeNil())
		klog.Info("Got 2 replicas")
	})

})

func newKubeClient(url string) kubernetes.Interface {
	klog.V(5).Infof("Create kubeclient for url %s\n", url)
	config, err := LoadConfig(url)
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	return clientset
}

func newKubeClientDynamic(url string) dynamic.Interface {
	klog.V(5).Infof("Create kubeclient dynamic for url %s\n", url)
	config, err := LoadConfig(url)
	if err != nil {
		panic(err)
	}

	clientset, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	return clientset
}

func LoadConfig(url string) (*rest.Config, error) {
	// If no in-cluster config, try the default location in the user's home directory.
	if usr, err := user.Current(); err == nil {
		klog.V(5).Infof("clientcmd.BuildConfigFromFlags for url %s using %s\n", url, filepath.Join(usr.HomeDir, ".kube", "config"))
		if c, err := clientcmd.BuildConfigFromFlags(url, filepath.Join(usr.HomeDir, ".kube", "config")); err == nil {
			return c, nil
		}
	}

	return nil, fmt.Errorf("could not create a valid kubeconfig")

}
