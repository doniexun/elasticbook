package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
)

// BookmarksFilePath want to guess which is the local bookmarks DB from the
// Chrome installation.
// This one is from my OSX, brew-installed, Chrome.
// "/Users/edoardo/Library/Application Support/Google/Chrome/Default/Bookmarks"
func BookmarksFilePath() string {
	user, err := user.Current()
	if err != nil {
		fmt.Fprintf(os.Stderr, "OS usupported? %s\n", err.Error())
	}

	return filepath.Join(
		user.HomeDir, "Library", "Application Support",
		"Google", "Chrome", "Default", "Bookmarks")
}

// BookmarksFile opens and return the local Chrome bookmarks file
func BookmarksFile() []byte {
	b, err := ioutil.ReadFile(BookmarksFilePath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to load file (%s)", err.Error())
		os.Exit(1)
	}
	return b
}
