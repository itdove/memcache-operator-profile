module github.com/operator-framework/operator-sdk-samples/go/memcached-operator

go 1.13

require (
	github.com/google/pprof v0.0.0-20200507031123-427632fa3b1c // indirect
	github.com/ianlancetaylor/demangle v0.0.0-20200414190113-039b1ae3a340 // indirect
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	github.com/operator-framework/operator-sdk v0.17.0
	github.com/spf13/pflag v1.0.5
	golang.org/x/sys v0.0.0-20200511232937-7e40ca221e25 // indirect
	k8s.io/api v0.17.4
	k8s.io/apimachinery v0.17.4
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/klog v1.0.0
	sigs.k8s.io/controller-runtime v0.5.2
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	k8s.io/client-go => k8s.io/client-go v0.17.4 // Required by prometheus-operator
)
