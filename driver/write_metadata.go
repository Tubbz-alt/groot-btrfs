package driver

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/groot"
	"code.cloudfoundry.org/grootfs/base_image_puller"
	"code.cloudfoundry.org/grootfs/store"
	"code.cloudfoundry.org/grootfs/store/filesystems"
	"code.cloudfoundry.org/lager"

	errorspkg "github.com/pkg/errors"
)

func (d *Driver) bundleMetaFilePath(bundleID string) string {
	return filepath.Join(d.conf.StorePath, store.MetaDirName, fmt.Sprintf("bundle-%s-metadata.json", bundleID))
}

// WriteMetadata writes a metadata file for a specific bundle.
func (d *Driver) WriteMetadata(logger lager.Logger, bundleID string, metadata groot.ImageMetadata) error {
	metaFile, err := os.Create(d.bundleMetaFilePath(bundleID))
	if err != nil {
		return errorspkg.Wrap(err, "creating metadata file")
	}

	if err = json.NewEncoder(metaFile).Encode(metadata); err != nil {
		return errorspkg.Wrap(err, "writing metadata file")
	}

	return nil
}

// WriteVolumeMeta writes a metadata file for the volume specified with "id"
func (d *Driver) WriteVolumeMeta(logger lager.Logger, id string, metadata base_image_puller.VolumeMeta) error {
	logger = logger.Session("btrfs-writing-volume-metadata", lager.Data{"volumeID": id})
	logger.Debug("starting")
	defer logger.Debug("ending")

	return filesystems.WriteVolumeMeta(logger, d.conf.StorePath, id, metadata)
}