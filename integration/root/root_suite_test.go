package root_test

import (
	"fmt"
	"math/rand"
	"os"
	"path"
	"time"

	"code.cloudfoundry.org/grootfs/integration"
	"code.cloudfoundry.org/grootfs/integration/runner"
	"code.cloudfoundry.org/grootfs/testhelpers"
	"code.cloudfoundry.org/lager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

var (
	GrootFSBin string
	DraxBin    string
	Runner     runner.Runner
	Driver     string

	GrootUID uint32
	GrootGID uint32

	storeName string
	StorePath string
	MountPath string
)

const btrfsMountPath = "/mnt/btrfs"
const xfsMountPath = "/mnt/xfs"

func TestRoot(t *testing.T) {
	RegisterFailHandler(Fail)
	rand.Seed(time.Now().Unix())

	SynchronizedBeforeSuite(func() []byte {
		grootFSBin, err := gexec.Build("code.cloudfoundry.org/grootfs")
		Expect(err).NotTo(HaveOccurred())

		fixPermission(path.Dir(grootFSBin))

		return []byte(grootFSBin)
	}, func(data []byte) {
		GrootUID = integration.FindUID("groot")
		GrootGID = integration.FindGID("groot")
		GrootFSBin = string(data)
		Driver = os.Getenv("VOLUME_DRIVER")
		if Driver == "overlay-xfs" {
			MountPath = xfsMountPath
		} else {
			Driver = "btrfs"
			MountPath = btrfsMountPath
		}

		fmt.Fprintf(os.Stderr, "============> RUNNING %s TESTS <=============", Driver)
	})

	SynchronizedAfterSuite(func() {
	}, func() {
		gexec.CleanupBuildArtifacts()
	})

	BeforeEach(func() {
		if os.Getuid() != 0 {
			Skip("This suite is only running as root")
		}

		storeName = fmt.Sprintf("test-store-%d", GinkgoParallelNode())
		StorePath = path.Join(MountPath, storeName)
		Expect(os.Mkdir(StorePath, 0700)).NotTo(HaveOccurred())

		Expect(os.Chown(StorePath, int(GrootUID), int(GrootGID))).To(Succeed())

		var err error
		DraxBin, err = gexec.Build("code.cloudfoundry.org/grootfs/store/filesystems/btrfs/drax")
		Expect(err).NotTo(HaveOccurred())
		testhelpers.SuidDrax(DraxBin)

		r := runner.Runner{
			GrootFSBin: GrootFSBin,
			StorePath:  StorePath,
			DraxBin:    DraxBin,
			Driver:     Driver,
		}
		Runner = r.WithLogLevel(lager.DEBUG).WithStderr(GinkgoWriter)
	})

	AfterEach(func() {
		testhelpers.CleanUpBtrfsSubvolumes(btrfsMountPath, storeName)
		testhelpers.CleanUpOverlayMounts(xfsMountPath, storeName)
		Expect(os.RemoveAll(StorePath)).To(Succeed())
	})

	RunSpecs(t, "GrootFS Integration Suite - Running as root")
}

func fixPermission(dirPath string) {
	fi, err := os.Stat(dirPath)
	Expect(err).NotTo(HaveOccurred())
	if !fi.IsDir() {
		return
	}

	// does other have the execute permission?
	if mode := fi.Mode(); mode&01 == 0 {
		Expect(os.Chmod(dirPath, 0755)).To(Succeed())
	}

	if dirPath == "/" {
		return
	}
	fixPermission(path.Dir(dirPath))
}
