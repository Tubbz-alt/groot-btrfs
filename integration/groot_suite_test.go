package integration_test

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"code.cloudfoundry.org/grootfs/integration/runner"
	"code.cloudfoundry.org/grootfs/store"
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
	Driver     string
	Runner     runner.Runner
	StorePath  string
	StoreName  string

	GrootUser        *user.User
	GrootUID         uint32
	GrootGID         uint32
	RegistryUsername string
	RegistryPassword string
)

const btrfsMountPath = "/mnt/btrfs"
const xfsMountPath = "/mnt/xfs-%d"

func TestGroot(t *testing.T) {
	RegisterFailHandler(Fail)

	SynchronizedBeforeSuite(func() []byte {
		grootFSBin, err := gexec.Build("code.cloudfoundry.org/grootfs")
		Expect(err).NotTo(HaveOccurred())

		return []byte(grootFSBin)
	}, func(data []byte) {
		var err error
		GrootUser, err = user.Lookup("groot")
		Expect(err).NotTo(HaveOccurred())

		grootUID, err := strconv.ParseUint(GrootUser.Uid, 10, 32)
		Expect(err).NotTo(HaveOccurred())
		GrootUID = uint32(grootUID)

		grootGID, err := strconv.ParseUint(GrootUser.Gid, 10, 32)
		Expect(err).NotTo(HaveOccurred())
		GrootGID = uint32(grootGID)

		GrootFSBin = string(data)
		Driver = os.Getenv("VOLUME_DRIVER")

		fmt.Fprintf(os.Stderr, "============> RUNNING %s TESTS <=============", Driver)
	})

	SynchronizedAfterSuite(func() {
	}, func() {
		gexec.CleanupBuildArtifacts()
	})

	BeforeEach(func() {
		rand.Seed(time.Now().UnixNano() + int64(GinkgoParallelNode()*1000))

		if Driver == "overlay-xfs" {
			StorePath = fmt.Sprintf(xfsMountPath, GinkgoParallelNode())
		} else {
			Driver = "btrfs"
			StoreName = fmt.Sprintf("test-store-%d", GinkgoParallelNode())
			StorePath = path.Join(btrfsMountPath, StoreName)
			Expect(os.Mkdir(StorePath, 0755)).NotTo(HaveOccurred())
		}

		Expect(os.Chmod(StorePath, 0777)).To(Succeed())

		var err error
		DraxBin, err = gexec.Build("code.cloudfoundry.org/grootfs/store/filesystems/btrfs/drax")
		Expect(err).NotTo(HaveOccurred())
		testhelpers.SuidDrax(DraxBin)

		RegistryUsername = os.Getenv("REGISTRY_USERNAME")
		RegistryPassword = os.Getenv("REGISTRY_PASSWORD")

		Runner = runner.Runner{
			GrootFSBin: GrootFSBin,
			StorePath:  StorePath,
			DraxBin:    DraxBin,
			Driver:     Driver,
		}.WithLogLevel(lager.DEBUG).WithStderr(GinkgoWriter)
	})

	AfterEach(func() {
		if Driver == "overlay-xfs" {
			testhelpers.CleanUpOverlayMounts(StorePath, "images")
		} else {
			testhelpers.CleanUpBtrfsSubvolumes(btrfsMountPath, StoreName)
			Expect(os.RemoveAll(StorePath)).To(Succeed())
		}

		os.RemoveAll(filepath.Join(StorePath, store.MetaDirName))
		os.RemoveAll(filepath.Join(StorePath, store.ImageDirName))
		os.RemoveAll(filepath.Join(StorePath, store.VolumesDirName))
		os.RemoveAll(filepath.Join(StorePath, store.LocksDirName))
		os.RemoveAll(filepath.Join(StorePath, store.CacheDirName))
		os.RemoveAll(filepath.Join(StorePath, store.TempDirName))
	})

	RunSpecs(t, "Integration Suite")
}

func writeMegabytes(outputPath string, mb int) error {
	cmd := exec.Command("dd", "if=/dev/zero", fmt.Sprintf("of=%s", outputPath), "bs=1048576", fmt.Sprintf("count=%d", mb))
	sess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	if err != nil {
		return err
	}
	Eventually(sess).Should(gexec.Exit())
	if sess.ExitCode() > 0 {
		return errors.New(string(sess.Err.Contents()))
	}
	return nil
}