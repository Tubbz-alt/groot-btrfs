package integration_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"code.cloudfoundry.org/grootfs/groot"
	"code.cloudfoundry.org/grootfs/integration"
	"code.cloudfoundry.org/grootfs/store"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Clean", func() {
	var (
		baseImagePath             string
		anotherBaseImagePath      string
		yetAnotherBaseImagePath   string
		sourceImagePath           string
		anotherSourceImagePath    string
		yetAnotherSourceImagePath string
	)

	BeforeEach(func() {
		var err error
		sourceImagePath, err = ioutil.TempDir("", "")
		sess, err := gexec.Start(exec.Command("dd", "if=/dev/zero", fmt.Sprintf("of=%s", filepath.Join(sourceImagePath, "foo")), "count=2", "bs=1M"), GinkgoWriter, nil)
		Expect(err).NotTo(HaveOccurred())
		Eventually(sess).Should(gexec.Exit(0))
		baseImageFile := integration.CreateBaseImageTar(sourceImagePath)
		baseImagePath = baseImageFile.Name()

		anotherSourceImagePath, err = ioutil.TempDir("", "")
		sess, err = gexec.Start(exec.Command("dd", "if=/dev/zero", fmt.Sprintf("of=%s", filepath.Join(anotherSourceImagePath, "foo")), "count=2", "bs=1M"), GinkgoWriter, nil)
		Expect(err).NotTo(HaveOccurred())
		Eventually(sess).Should(gexec.Exit(0))
		anotherBaseImageFile := integration.CreateBaseImageTar(anotherSourceImagePath)
		anotherBaseImagePath = anotherBaseImageFile.Name()

		yetAnotherSourceImagePath, err = ioutil.TempDir("", "")
		yetAnotherBaseImageFile := integration.CreateBaseImageTar(yetAnotherSourceImagePath)
		yetAnotherBaseImagePath = yetAnotherBaseImageFile.Name()

		_, err = Runner.Create(groot.CreateSpec{
			ID:        "my-image-1",
			BaseImage: baseImagePath,
		})
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(Runner.Delete("my-image-1")).To(Succeed())
	})

	Context("when the store doesn't exist", func() {
		It("logs an error message and exits successfully", func() {
			logBuffer := gbytes.NewBuffer()
			_, err := Runner.WithStore("/invalid-store").WithStderr(logBuffer).
				Clean(0, []string{})
			Expect(err).ToNot(HaveOccurred())
			Expect(logBuffer).To(gbytes.Say(`"error":"no store found at /invalid-store"`))
		})
	})

	Context("when there are unused volumes", func() {
		BeforeEach(func() {
			_, err := Runner.Create(groot.CreateSpec{
				ID:        "my-image-2",
				BaseImage: anotherBaseImagePath,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(Runner.Delete("my-image-2")).To(Succeed())
		})

		It("removes them", func() {
			preContents, err := ioutil.ReadDir(filepath.Join(StorePath, store.VolumesDirName))
			Expect(err).NotTo(HaveOccurred())
			Expect(preContents).To(HaveLen(2))

			_, err = Runner.Clean(0, []string{})
			Expect(err).NotTo(HaveOccurred())

			afterContents, err := ioutil.ReadDir(filepath.Join(StorePath, store.VolumesDirName))
			Expect(err).NotTo(HaveOccurred())
			Expect(afterContents).To(HaveLen(1))

			afterDir := afterContents[0].Name()
			afterContentsSha := sha256.Sum256([]byte(baseImagePath))
			Expect(afterDir).To(MatchRegexp(`%s-\d+`, hex.EncodeToString(afterContentsSha[:32])))
		})

		Context("and ignored images flag is given", func() {
			var preContents []os.FileInfo

			JustBeforeEach(func() {
				var err error
				preContents, err = ioutil.ReadDir(filepath.Join(StorePath, store.VolumesDirName))
				Expect(err).NotTo(HaveOccurred())
			})

			It("doesn't delete their layers", func() {
				_, err := Runner.Clean(0, []string{anotherBaseImagePath})
				Expect(err).NotTo(HaveOccurred())

				afterContents, err := ioutil.ReadDir(filepath.Join(StorePath, store.VolumesDirName))
				Expect(err).NotTo(HaveOccurred())

				Expect(afterContents).To(Equal(preContents))
			})

			Context("when more than one image is to be ignored", func() {
				BeforeEach(func() {
					_, err := Runner.Create(groot.CreateSpec{
						ID:        "my-image-3",
						BaseImage: yetAnotherBaseImagePath,
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(Runner.Delete("my-image-3")).To(Succeed())
				})

				It("doesn't delete their layers", func() {
					_, err := Runner.Clean(0, []string{anotherBaseImagePath, yetAnotherBaseImagePath})
					Expect(err).NotTo(HaveOccurred())

					afterContents, err := ioutil.ReadDir(filepath.Join(StorePath, store.VolumesDirName))
					Expect(err).NotTo(HaveOccurred())

					Expect(afterContents).To(Equal(preContents))
				})
			})
		})

		Context("and a threshold is set", func() {
			var cleanupThresholdInBytes int64

			Context("and the total is less than the threshold", func() {
				BeforeEach(func() {
					cleanupThresholdInBytes = 50000000
				})

				It("does not remove the unused volumes", func() {
					preContents, err := ioutil.ReadDir(filepath.Join(StorePath, store.VolumesDirName))
					Expect(err).NotTo(HaveOccurred())

					_, err = Runner.Clean(cleanupThresholdInBytes, []string{})
					Expect(err).NotTo(HaveOccurred())

					afterContents, err := ioutil.ReadDir(filepath.Join(StorePath, store.VolumesDirName))
					Expect(err).NotTo(HaveOccurred())
					Expect(afterContents).To(HaveLen(len(preContents)))
				})

				It("reports that it was a no-op", func() {
					output, err := Runner.Clean(cleanupThresholdInBytes, []string{})
					Expect(err).NotTo(HaveOccurred())
					Expect(output).To(ContainSubstring("threshold not reached: skipping clean"))
				})
			})

			Context("and the total is more than the threshold", func() {
				BeforeEach(func() {
					cleanupThresholdInBytes = 70000
				})

				It("removes the unused volumes", func() {
					preContents, err := ioutil.ReadDir(filepath.Join(StorePath, store.VolumesDirName))
					Expect(err).NotTo(HaveOccurred())
					Expect(preContents).To(HaveLen(2))

					_, err = Runner.Clean(cleanupThresholdInBytes, []string{})
					Expect(err).NotTo(HaveOccurred())

					afterContents, err := ioutil.ReadDir(filepath.Join(StorePath, store.VolumesDirName))
					Expect(err).NotTo(HaveOccurred())
					Expect(afterContents).To(HaveLen(1))
				})
			})
		})
	})
})