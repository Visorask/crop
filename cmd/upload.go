package cmd

import (
	"github.com/dustin/go-humanize"
	"github.com/l3uddz/crop/config"
	"github.com/l3uddz/crop/uploader"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var uploadCmd = &cobra.Command{
	Use:   "upload [UPLOADER]",
	Short: "Perform uploader task",
	Long:  `This command can be used to trigger an uploader check / upload.`,

	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// init core
		initCore(true)

		// iterate uploader's
		for uploaderName, uploaderConfig := range config.Config.Uploader {
			// skip disabled uploader(s)
			if !uploaderConfig.Enabled {
				log.WithField("uploader", uploaderName).Trace("Skipping disabled uploader")
				continue
			}

			log := log.WithField("uploader", uploaderName)

			// create uploader
			upload, err := uploader.New(config.Config, &uploaderConfig, uploaderName)
			if err != nil {
				log.WithField("uploader", uploaderName).WithError(err).
					Error("Failed initializing uploader, skipping...")
				continue
			}

			if upload.ServiceAccountCount > 0 {
				upload.Log.WithField("found_files", upload.ServiceAccountCount).
					Info("Loaded service accounts")
			} else {
				// no service accounts were loaded
				// check to see if any of the copy or move remote(s) are banned
				banned, expiry := upload.RemotesBanned(upload.Config.Remotes.Copy)
				if banned && !expiry.IsZero() {
					// one of the copy remotes is banned, abort
					upload.Log.WithFields(logrus.Fields{
						"expires_time": expiry,
						"expires_in":   humanize.Time(expiry),
					}).Warn("Cannot proceed with upload as a copy remote is banned")
					continue
				}

				banned, expiry = upload.RemotesBanned([]string{upload.Config.Remotes.Move})
				if banned && !expiry.IsZero() {
					// the move remote is banned, abort
					upload.Log.WithFields(logrus.Fields{
						"expires_time": expiry,
						"expires_in":   humanize.Time(expiry),
					}).Warn("Cannot proceed with upload as the move remote is banned")
					continue
				}
			}

			log.Info("Uploader commencing...")

			// refresh details about files to upload
			if err := upload.RefreshLocalFiles(); err != nil {
				upload.Log.WithError(err).Error("Failed refreshing details of files to upload")
				continue
			}

			// check if upload criteria met
			if shouldUpload, err := upload.Check(); err != nil {
				upload.Log.WithError(err).Error("Failed checking if uploader check conditions met, skipping...")
				continue
			} else if !shouldUpload {
				upload.Log.Info("Upload conditions not met, skipping...")
				continue
			}

			// perform upload
			if err := performUpload(upload); err != nil {
				upload.Log.WithError(err).Error("Error occurred while running uploader, skipping...")
				continue
			}

			// clean local upload folder of empty directories
			upload.Log.Debug("Cleaning empty local directories...")
		}

	},
}

func init() {
	rootCmd.AddCommand(uploadCmd)
}

func performUpload(u *uploader.Uploader) error {
	u.Log.Info("Running...")

	/* Cleans */
	if u.Config.Hidden.Enabled {
		err := performClean(u)
		if err != nil {
			return errors.Wrap(err, "failed clearing remotes")
		}
	}

	/* Generate Additional Rclone Params */
	additionalRcloneParams := u.CheckRcloneParams()

	/* Copies */
	if len(u.Config.Remotes.Copy) > 0 {
		u.Log.Info("Running copies...")

		if err := u.Copy(additionalRcloneParams); err != nil {
			return errors.WithMessage(err, "failed performing all copies")
		}

		u.Log.Info("Finished copies!")
	}

	/* Move */
	if len(u.Config.Remotes.Move) > 0 {
		u.Log.Info("Running move...")

		if err := u.Move(false, additionalRcloneParams); err != nil {
			return errors.WithMessage(err, "failed performing move")
		}

		u.Log.Info("Finished move!")
	}

	/* Move Server Side */
	if len(u.Config.Remotes.MoveServerSide) > 0 {
		u.Log.Info("Running move server-sides...")

		if err := u.Move(true, nil); err != nil {
			return errors.WithMessage(err, "failed performing server-side moves")
		}

		u.Log.Info("Finished move server-sides!")
	}

	u.Log.Info("Finished!")
	return nil
}
