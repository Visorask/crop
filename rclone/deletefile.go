package rclone

import (
	"github.com/go-cmd/cmd"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

/* Public */

func DeleteFile(remoteFilePath string) (bool, int, error) {
	// set variables
	rLog := log.WithFields(logrus.Fields{
		"action":      CmdDeleteFile,
		"remote_path": remoteFilePath,
	})
	result := false

	// generate required rclone parameters
	params := []string{
		CmdDeleteFile,
		remoteFilePath,
	}

	if baseParams, err := getBaseParams(); err != nil {
		return false, 1, errors.WithMessagef(err, "failed generating baseParams to %q: %q", CmdDeleteFile,
			remoteFilePath)
	} else {
		params = append(params, baseParams...)
	}

	rLog.Debugf("Generated params: %v", params)

	// remove file
	rcloneCmd := cmd.NewCmd(cfg.Rclone.Path, params...)
	status := <-rcloneCmd.Start()

	// check status
	switch status.Exit {
	case ExitSuccess:
		result = true
	default:
		break
	}

	rLog.WithField("exit_code", status.Exit).Debug("Finished")
	return result, status.Exit, status.Error
}
