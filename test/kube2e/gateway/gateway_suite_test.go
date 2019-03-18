package gateway_test

import (
	"github.com/avast/retry-go"
	"github.com/solo-io/gloo/test/kube2e"
	"github.com/solo-io/go-utils/testutils/clusterlock"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/solo-io/go-utils/testutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils/helper"
	skhelpers "github.com/solo-io/solo-kit/test/helpers"
)

func TestGateway(t *testing.T) {
	if testutils.AreTestsDisabled() {
		return
	}
	skhelpers.RegisterCommonFailHandlers()
	skhelpers.SetupLog()
	RunSpecs(t, "Gateway Suite")
}

var testHelper *helper.SoloTestHelper
var failed bool
var locker *clusterlock.TestClusterLocker

var _ = BeforeSuite(func() {
	cwd, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())

	testHelper, err = helper.NewSoloTestHelper(func(defaults helper.TestConfig) helper.TestConfig {
		defaults.RootDir = filepath.Join(cwd, "../../..")
		defaults.HelmChartName = "gloo"
		return defaults
	})
	Expect(err).NotTo(HaveOccurred())

	locker, err = clusterlock.NewTestClusterLocker(kube2e.MustKubeClient(), "")
	Expect(err).NotTo(HaveOccurred())
	Expect(locker.AcquireLock(retry.Attempts(8), retry.Delay(2 * time.Second))).NotTo(HaveOccurred())

	// Install Gloo
	err = testHelper.InstallGloo(helper.GATEWAY, 5*time.Minute)
	Expect(err).NotTo(HaveOccurred())
	failed = false
})

var _ = AfterEach(func() {
	failed = failed || CurrentGinkgoTestDescription().Failed
})

var _ = AfterSuite(func() {
	defer locker.ReleaseLock()
	// stop cleaning up after failures
	err := testHelper.UninstallGloo()
	Expect(err).NotTo(HaveOccurred())

	Eventually(func() error {
		return testutils.Kubectl("get", "namespace", testHelper.InstallNamespace)
	}, "60s", "1s").Should(HaveOccurred())
})
