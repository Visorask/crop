package uploader

import (
	"fmt"
	"github.com/l3uddz/crop/cache"
	"github.com/l3uddz/crop/pathutils"
	"github.com/l3uddz/crop/rclone"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

func (u *Uploader) Copy() error {
	// iterate all remotes and run copy
	for _, remotePath := range u.Config.Remotes.Copy {
		// set logic variables
		attempts := 1

		// set log
		rLog := u.Log.WithFields(logrus.Fields{
			"copy_remote":     remotePath,
			"copy_local_path": u.Config.LocalFolder,
			"attempts":        attempts,
		})

		// copy to remote
		for {
			// get service account file
			var serviceAccount *pathutils.Path
			var err error

			if u.ServiceAccountCount > 0 {
				serviceAccount, err = u.getAvailableServiceAccount()
				if err != nil {
					return errors.WithMessagef(err,
						"aborting further copy attempts of %q due to serviceAccount exhaustion",
						u.Config.LocalFolder)
				}

				// reset log
				rLog = u.Log.WithFields(logrus.Fields{
					"copy_remote":     remotePath,
					"copy_local_path": u.Config.LocalFolder,
					"attempts":        attempts,
					"service_account": serviceAccount.RealPath,
				})
			}

			// copy
			rLog.Info("Copying...")
			success, exitCode, err := rclone.Copy(u.Config, u.Config.LocalFolder, remotePath, serviceAccount)

			// check result
			if err != nil {
				rLog.WithError(err).Errorf("Failed unexpectedly...")
				return errors.WithMessagef(err, "copy failed unexpectedly with exit code: %v", exitCode)
			} else if success {
				// successful exit code
				break
			}

			// are we using service accounts?
			if u.ServiceAccountCount == 0 {
				return fmt.Errorf("copy failed with exit code: %v", exitCode)
			}

			// is this an exit code we can retry?
			switch exitCode {
			case rclone.EXIT_FATAL_ERROR:
				// ban this service account
				if err := cache.Set(serviceAccount.RealPath, time.Now().UTC().Add(25*time.Hour)); err != nil {
					rLog.WithError(err).Error("Failed banning service account, cannot proceed...")
					return fmt.Errorf("failed banning service account: %v", serviceAccount.RealPath)
				}

				// attempt copy again
				rLog.Warnf("Copy failed with retryable exit code %v, attempting again...", exitCode)
				attempts++
				continue
			default:
				return fmt.Errorf("failed and cannot proceed with exit code: %v", exitCode)
			}
		}
	}

	return nil
}