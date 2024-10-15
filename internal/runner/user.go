package runner

import (
	"fmt"
	"log/slog"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
)

// RealUser returns a user struct containing details of the "real" user, which
// may differ from the current user when concierge is executed with `sudo`.
func RealUser() (*user.User, error) {
	realUser := os.Getenv("SUDO_USER")
	if len(realUser) == 0 {
		return user.Lookup("root")
	}

	return user.Lookup(realUser)
}

// ChownRecursively recursively changes ownership of a given filepath to the uid/gid of
// the specified user.
func ChownRecursively(path string, user *user.User) error {
	uid, err := strconv.Atoi(user.Uid)
	if err != nil {
		return fmt.Errorf("failed to convert user id string to int: %w", err)
	}
	gid, err := strconv.Atoi(user.Gid)
	if err != nil {
		return fmt.Errorf("failed to convert group id string to int: %w", err)
	}

	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		err = os.Chown(path, uid, gid)
		if err != nil {
			return err
		}

		return nil
	})

	slog.Debug("Filesystem ownership changed", "user", user.Username, "group", user.Gid, "path", path)
	return err
}
