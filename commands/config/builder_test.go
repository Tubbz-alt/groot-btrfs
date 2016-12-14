package config_test

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"code.cloudfoundry.org/grootfs/commands/config"
	yaml "gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Builder", func() {
	var (
		configDir                string
		configFilePath           string
		builder                  *config.Builder
		configStorePath          string
		configDraxBin            string
		configMetronEndpoint     string
		configUIDMappings        []string
		configGIDMappings        []string
		configDiskLimitSizeBytes *int64
	)

	BeforeEach(func() {
		configStorePath = "/hello"
		configDraxBin = "/config/drax"
		configMetronEndpoint = "config_endpoint:1111"
		configUIDMappings = []string{"config-uid-mapping"}
		configGIDMappings = []string{"config-gid-mapping"}
		diskLimit := int64(1000)
		configDiskLimitSizeBytes = &diskLimit
	})

	JustBeforeEach(func() {
		var err error
		configDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		cfg := config.Config{
			InsecureRegistries: []string{"http://example.org"},
			IgnoreBaseImages:   []string{"docker:///busybox"},
			BaseStorePath:      configStorePath,
			DraxBin:            configDraxBin,
			MetronEndpoint:     configMetronEndpoint,
			UIDMappings:        configUIDMappings,
			GIDMappings:        configGIDMappings,
			DiskLimitSizeBytes: configDiskLimitSizeBytes,
		}

		configYaml, err := yaml.Marshal(cfg)
		Expect(err).NotTo(HaveOccurred())
		configFilePath = path.Join(configDir, "config.yaml")

		Expect(ioutil.WriteFile(configFilePath, configYaml, 0755)).To(Succeed())
		builder, err = config.NewBuilder(configFilePath)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(configDir)).To(Succeed())
	})

	Describe("Build", func() {
		It("returns the values read from the config yaml", func() {
			config, err := builder.Build()
			Expect(err).NotTo(HaveOccurred())
			Expect(config.InsecureRegistries).To(Equal([]string{"http://example.org"}))
			Expect(config.IgnoreBaseImages).To(Equal([]string{"docker:///busybox"}))
			Expect(config.BaseStorePath).To(Equal("/hello"))
			Expect(config.UserBasedStorePath).To(Equal(filepath.Join("/hello", CurrentUserID)))
		})

		Context("when disk limit property is invalid", func() {
			BeforeEach(func() {
				diskLimit := int64(-1)
				configDiskLimitSizeBytes = &diskLimit
			})

			It("returns an error", func() {
				_, err := builder.Build()
				Expect(err).To(MatchError("invalid argument: disk limit cannot be negative"))
			})
		})
	})

	Describe("WithInsecureRegistries", func() {
		It("overrides the config's InsecureRegistries entry", func() {
			builder = builder.WithInsecureRegistries([]string{"1", "2"})
			config, err := builder.Build()
			Expect(err).NotTo(HaveOccurred())
			Expect(config.InsecureRegistries).To(Equal([]string{"1", "2"}))
		})

		Context("when empty", func() {
			It("doesn't override the config's InsecureRegistries entry", func() {
				builder = builder.WithInsecureRegistries([]string{})
				config, err := builder.Build()
				Expect(err).NotTo(HaveOccurred())
				Expect(config.InsecureRegistries).To(Equal([]string{"http://example.org"}))
			})
		})

		Context("when nil", func() {
			It("doesn't override the config's InsecureRegistries entry", func() {
				builder = builder.WithInsecureRegistries(nil)
				config, err := builder.Build()
				Expect(err).NotTo(HaveOccurred())
				Expect(config.InsecureRegistries).To(Equal([]string{"http://example.org"}))
			})
		})
	})

	Describe("WithIgnoreBaseImages", func() {
		It("overrides the config's IgnoreBaseImages entry", func() {
			builder = builder.WithIgnoreBaseImages([]string{"1", "2"})
			config, err := builder.Build()
			Expect(err).NotTo(HaveOccurred())
			Expect(config.IgnoreBaseImages).To(Equal([]string{"1", "2"}))
		})

		Context("when empty", func() {
			It("doesn't override the config's IgnoreBaseImages entry", func() {
				builder = builder.WithIgnoreBaseImages([]string{})
				config, err := builder.Build()
				Expect(err).NotTo(HaveOccurred())
				Expect(config.IgnoreBaseImages).To(Equal([]string{"docker:///busybox"}))
			})
		})

		Context("when nil", func() {
			It("doesn't override the config's IgnoreBaseImages entry", func() {
				builder = builder.WithIgnoreBaseImages(nil)
				config, err := builder.Build()
				Expect(err).NotTo(HaveOccurred())
				Expect(config.IgnoreBaseImages).To(Equal([]string{"docker:///busybox"}))
			})
		})
	})

	Describe("WithStorePath", func() {
		Context("when provided store path and default store path are different", func() {
			It("overrides the config's store path entry with the provided store path with user ID postfix", func() {
				builder = builder.WithStorePath("/mnt/grootfs/data", "/var/lib/grootfs/data")
				config, err := builder.Build()
				Expect(err).NotTo(HaveOccurred())
				Expect(config.UserBasedStorePath).To(Equal(filepath.Join("/mnt/grootfs/data", CurrentUserID)))
				Expect(config.BaseStorePath).To(Equal("/mnt/grootfs/data"))
			})
		})

		Context("when provided store path and default store path are the same", func() {
			It("uses the config's store path with user ID postfix", func() {
				builder = builder.WithStorePath("/var/lib/grootfs/data", "/var/lib/grootfs/data")
				config, err := builder.Build()
				Expect(err).NotTo(HaveOccurred())
				Expect(config.UserBasedStorePath).To(Equal(filepath.Join("/hello", CurrentUserID)))
				Expect(config.BaseStorePath).To(Equal("/hello"))
			})

		})

		Context("when config doesn't have a store path property", func() {
			BeforeEach(func() {
				configStorePath = ""
			})

			It("uses the provided store path with user ID postfix", func() {
				builder = builder.WithStorePath("/var/lib/grootfs/data", "/var/lib/grootfs/data")
				config, err := builder.Build()
				Expect(err).NotTo(HaveOccurred())
				Expect(config.UserBasedStorePath).To(Equal(filepath.Join("/var/lib/grootfs/data", CurrentUserID)))
				Expect(config.BaseStorePath).To(Equal("/var/lib/grootfs/data"))
			})
		})
	})

	Describe("WithDraxBin", func() {
		Context("when provided drax bin and default drax bin are different", func() {
			It("overrides the config's drax bin entry", func() {
				builder = builder.WithDraxBin("/my/drax", "/default/drax")
				config, err := builder.Build()
				Expect(err).NotTo(HaveOccurred())
				Expect(config.DraxBin).To(Equal("/my/drax"))
			})
		})

		Context("when provided drax bin and default drax bin are the same", func() {
			It("uses the config's drax bin", func() {
				builder = builder.WithDraxBin("/default/drax", "/default/drax")
				config, err := builder.Build()
				Expect(err).NotTo(HaveOccurred())
				Expect(config.DraxBin).To(Equal("/config/drax"))
			})

		})

		Context("when config doesn't have a drax bin property", func() {
			BeforeEach(func() {
				configDraxBin = ""
			})

			It("uses the provided drax bin", func() {
				builder = builder.WithDraxBin("/default/drax", "/default/drax")
				config, err := builder.Build()
				Expect(err).NotTo(HaveOccurred())
				Expect(config.DraxBin).To(Equal("/default/drax"))
			})
		})
	})

	Describe("WithMetronEndpoint", func() {
		It("overrides the config's metron endpoint entry", func() {
			builder = builder.WithMetronEndpoint("127.0.0.1:5555")
			config, err := builder.Build()
			Expect(err).NotTo(HaveOccurred())
			Expect(config.MetronEndpoint).To(Equal("127.0.0.1:5555"))
		})

		Context("when empty", func() {
			It("doesn't override the config's metron endpoint entry", func() {
				builder = builder.WithMetronEndpoint("")
				config, err := builder.Build()
				Expect(err).NotTo(HaveOccurred())
				Expect(config.MetronEndpoint).To(Equal("config_endpoint:1111"))
			})
		})
	})

	Describe("WithUIDMappings", func() {
		It("overrides the config's UIDMappings entry", func() {
			builder = builder.WithUIDMappings([]string{"1", "2"})
			config, err := builder.Build()
			Expect(err).NotTo(HaveOccurred())
			Expect(config.UIDMappings).To(Equal([]string{"1", "2"}))
		})

		Context("when empty", func() {
			It("doesn't override the config's UIDMappings entry", func() {
				builder = builder.WithUIDMappings([]string{})
				config, err := builder.Build()
				Expect(err).NotTo(HaveOccurred())
				Expect(config.UIDMappings).To(Equal([]string{"config-uid-mapping"}))
			})
		})

		Context("when nil", func() {
			It("doesn't override the config's UIDMappings entry", func() {
				builder = builder.WithUIDMappings(nil)
				config, err := builder.Build()
				Expect(err).NotTo(HaveOccurred())
				Expect(config.UIDMappings).To(Equal([]string{"config-uid-mapping"}))
			})
		})
	})

	Describe("WithGIDMappings", func() {
		It("overrides the config's GIDMappings entry", func() {
			builder = builder.WithGIDMappings([]string{"1", "2"})
			config, err := builder.Build()
			Expect(err).NotTo(HaveOccurred())
			Expect(config.GIDMappings).To(Equal([]string{"1", "2"}))
		})

		Context("when empty", func() {
			It("doesn't override the config's GIDMappings entry", func() {
				builder = builder.WithGIDMappings([]string{})
				config, err := builder.Build()
				Expect(err).NotTo(HaveOccurred())
				Expect(config.GIDMappings).To(Equal([]string{"config-gid-mapping"}))
			})
		})

		Context("when nil", func() {
			It("doesn't override the config's GIDMappings entry", func() {
				builder = builder.WithGIDMappings(nil)
				config, err := builder.Build()
				Expect(err).NotTo(HaveOccurred())
				Expect(config.GIDMappings).To(Equal([]string{"config-gid-mapping"}))
			})
		})
	})

	Describe("WithDiskLimitSizeBytes", func() {
		It("overrides the config's DiskLimitSizeBytes entry", func() {
			diskLimit := int64(3000)
			builder = builder.WithDiskLimitSizeBytes(&diskLimit)
			config, err := builder.Build()
			Expect(err).NotTo(HaveOccurred())
			Expect(*config.DiskLimitSizeBytes).To(Equal(diskLimit))
		})

		Context("when negative", func() {
			It("returns an error", func() {
				diskLimit := int64(-300)
				builder = builder.WithDiskLimitSizeBytes(&diskLimit)
				_, err := builder.Build()
				Expect(err).To(MatchError("invalid argument: disk limit cannot be negative"))
			})
		})

		Context("when nil", func() {
			It("doesn't override the config's DiskLimitSizeBytes entry", func() {
				builder = builder.WithDiskLimitSizeBytes(nil)
				config, err := builder.Build()
				Expect(err).NotTo(HaveOccurred())
				Expect(*config.DiskLimitSizeBytes).To(Equal(*configDiskLimitSizeBytes))
			})
		})

		Context("when neither the config file or flag values are specified", func() {
			BeforeEach(func() {
				configDiskLimitSizeBytes = nil
			})

			It("defaults the limit to 0", func() {
				builder = builder.WithDiskLimitSizeBytes(nil)
				config, err := builder.Build()
				Expect(err).NotTo(HaveOccurred())
				Expect(*config.DiskLimitSizeBytes).To(Equal(int64(0)))
			})
		})
	})
})
