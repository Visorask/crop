package uploader

import (
	"github.com/l3uddz/crop/uploader/checker"
)

var (
	supportedCheckers = map[string]interface{}{
		"size": checker.Size{},
		"age":  checker.Age{},
	}
)

func (u *Uploader) Check() (bool, error) {
	// Perform the check
	return u.Checker.Check(&u.Config.Check, u.Log, u.LocalFiles, u.LocalFilesSize)
}