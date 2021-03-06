package rclone

import (
	"github.com/go-cmd/cmd"
	"github.com/l3uddz/crop/pathutils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

/* Public */

func Sync(from string, to string, serviceAccountFile *pathutils.Path,
	additionalRcloneParams []string) (bool, int, error) {
	// set variables
	rLog := log.WithFields(logrus.Fields{
		"action": CmdSync,
		"from":   from,
		"to":     to,
	})
	result := false

	// generate required rclone parameters
	params := []string{
		CmdSync,
		from,
		to,
	}

	if baseParams, err := getBaseParams(); err != nil {
		return false, 1, errors.WithMessagef(err, "failed generating baseParams to %q: %q -> %q",
			CmdSync, from, to)
	} else {
		params = append(params, baseParams...)
	}

	extraParams := additionalRcloneParams

	if additionalParams, err := getAdditionalParams(CmdSync, extraParams); err != nil {
		return false, 1, errors.WithMessagef(err, "failed generating additionalParams to %q: %q -> %q",
			CmdSync, from, to)
	} else {
		params = append(params, additionalParams...)
	}

	if serviceAccountFile != nil {
		saParams := getServiceAccountParams(serviceAccountFile)
		params = append(params, saParams...)
	}

	rLog.Debugf("Generated params: %v", params)

	// setup cmd
	cmdOptions := cmd.Options{
		Buffered:  false,
		Streaming: true,
	}
	rcloneCmd := cmd.NewCmdOptions(cmdOptions, cfg.Rclone.Path, params...)

	// live stream logs
	doneChan := make(chan struct{})
	go func() {
		defer close(doneChan)

		for rcloneCmd.Stdout != nil || rcloneCmd.Stderr != nil {
			select {
			case line, open := <-rcloneCmd.Stdout:
				if !open {
					rcloneCmd.Stdout = nil
					continue
				}
				log.Info(line)
			case line, open := <-rcloneCmd.Stderr:
				if !open {
					rcloneCmd.Stderr = nil
					continue
				}
				log.Info(line)
			}
		}
	}()

	// run command
	rLog.Debug("Starting...")

	status := <-rcloneCmd.Start()
	<-doneChan

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
