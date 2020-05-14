package memcache_operator_profile_test

import (
	"flag"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/klog"
)

var masterUrl string

func init() {
	klog.SetOutput(GinkgoWriter)
	klog.InitFlags(nil)

	flag.StringVar(&masterUrl, "master-url", "", "The cluster master url")

}

func TestMemcacheOperatorProfile(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MemcacheOperatorProfile Suite")
}
