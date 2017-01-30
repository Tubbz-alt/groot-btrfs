package store_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"syscall"

	"code.cloudfoundry.org/grootfs/store"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Configurer", func() {
	var (
		storePath      string
		originalTmpDir string
		currentUID     int
		currentGID     int
		logger         lager.Logger
	)

	BeforeEach(func() {
		originalTmpDir = os.TempDir()
		tempDir, err := ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		currentUID = os.Getuid()
		currentGID = os.Getgid()
		storePath = path.Join(tempDir, "store")

		logger = lagertest.NewTestLogger("store-configurer")
	})

	AfterEach(func() {
		Expect(os.RemoveAll(path.Dir(storePath))).To(Succeed())
		Expect(os.Setenv("TMPDIR", originalTmpDir)).To(Succeed())
	})

	Describe("ConfigureStore", func() {
		It("creates the store directory", func() {
			Expect(store.ConfigureStore(logger, storePath, currentUID, currentGID, "random-id")).To(Succeed())

			Expect(storePath).To(BeADirectory())
		})

		It("creates the correct internal structure", func() {
			Expect(store.ConfigureStore(logger, storePath, currentUID, currentGID, "random-id")).To(Succeed())

			Expect(filepath.Join(storePath, "images")).To(BeADirectory())
			Expect(filepath.Join(storePath, "cache")).To(BeADirectory())
			Expect(filepath.Join(storePath, "volumes")).To(BeADirectory())
			Expect(filepath.Join(storePath, "tmp")).To(BeADirectory())
			Expect(filepath.Join(storePath, "locks")).To(BeADirectory())
			Expect(filepath.Join(storePath, "meta")).To(BeADirectory())
			Expect(filepath.Join(storePath, "meta", "dependencies")).To(BeADirectory())
		})

		It("creates tmp files into TMPDIR inside storePath", func() {
			Expect(store.ConfigureStore(logger, storePath, currentUID, currentGID, "random-id")).To(Succeed())
			file, _ := ioutil.TempFile("", "")
			Expect(filepath.Join(storePath, store.TEMP_DIR_NAME, filepath.Base(file.Name()))).To(BeAnExistingFile())

		})

		It("chmods the storePath to 700", func() {
			Expect(store.ConfigureStore(logger, storePath, currentUID, currentGID, "random-id")).To(Succeed())

			stat, err := os.Stat(storePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(stat.Mode().Perm()).To(Equal(os.FileMode(0700)))
		})

		It("chowns the storePath to the owner UID/GID", func() {
			Expect(store.ConfigureStore(logger, storePath, currentUID, currentGID, "random-id")).To(Succeed())

			stat, err := os.Stat(storePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(stat.Sys().(*syscall.Stat_t).Uid).To(Equal(uint32(currentUID)))
			Expect(stat.Sys().(*syscall.Stat_t).Gid).To(Equal(uint32(currentGID)))
		})

		It("doesn't fail on race conditions", func() {
			for i := 0; i < 50; i++ {
				storePath, err := ioutil.TempDir("", "")
				Expect(err).NotTo(HaveOccurred())
				start1 := make(chan bool)
				start2 := make(chan bool)

				go func() {
					defer GinkgoRecover()
					<-start1
					Expect(store.ConfigureStore(logger, storePath, currentUID, currentGID, "random-id")).To(Succeed())
					close(start1)
				}()

				go func() {
					defer GinkgoRecover()
					<-start2
					Expect(store.ConfigureStore(logger, storePath, currentUID, currentGID, "random-id")).To(Succeed())
					close(start2)
				}()

				start1 <- true
				start2 <- true

				Eventually(start1).Should(BeClosed())
				Eventually(start2).Should(BeClosed())
			}
		})

		Context("it will change the owner of the created folders to the provided userID", func() {
			It("returns an error", func() {
				Expect(store.ConfigureStore(logger, storePath, 0, 1, "random-id")).To(
					MatchError(ContainSubstring("changing store owner to 0:1 for path")),
				)
			})
		})

		Context("when it can't change the store path ownership", func() {
			BeforeEach(func() {
				Expect(os.Mkdir(storePath, 0777)).To(Succeed())
				Expect(os.Chmod(storePath, 0777)).To(Succeed())
				chown := exec.Command("sudo", "chown", "5000:5000", storePath)
				sess, err := gexec.Start(chown, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(gexec.Exit(0))
			})

			It("returns an error", func() {
				err := store.ConfigureStore(logger, storePath, currentUID, currentGID, "random-id")
				Expect(err).To(MatchError(ContainSubstring("changing store owner to 1000:1000 for path")))
			})
		})

		Context("when the base directory does not exist", func() {
			It("returns an error", func() {
				Expect(store.ConfigureStore(logger, "/not/exist", currentUID, currentGID, "random-id")).To(
					MatchError(ContainSubstring("making directory")),
				)
			})
		})

		Context("when the store already exists", func() {
			It("succeeds", func() {
				Expect(os.Mkdir(storePath, 0700)).To(Succeed())
				Expect(store.ConfigureStore(logger, storePath, currentUID, currentGID, "random-id")).To(Succeed())
			})

			Context("and it's a regular file", func() {
				It("returns an error", func() {
					Expect(ioutil.WriteFile(storePath, []byte("hello"), 0600)).To(Succeed())

					Expect(store.ConfigureStore(logger, storePath, currentUID, currentGID, "random-id")).To(
						MatchError(ContainSubstring("is not a directory")),
					)
				})
			})
		})

		Context("when any internal directory already exists", func() {
			It("succeeds", func() {
				Expect(os.MkdirAll(filepath.Join(storePath, "volumes"), 0700)).To(Succeed())
				Expect(store.ConfigureStore(logger, storePath, currentUID, currentGID, "random-id")).To(Succeed())
			})

			Context("and it's a regular file", func() {
				It("returns an error", func() {
					Expect(os.Mkdir(storePath, 0700)).To(Succeed())
					Expect(ioutil.WriteFile(filepath.Join(storePath, "volumes"), []byte("hello"), 0600)).To(Succeed())

					Expect(store.ConfigureStore(logger, storePath, currentUID, currentGID, "random-id")).To(
						MatchError(ContainSubstring("is not a directory")),
					)
				})
			})
		})
	})
})
