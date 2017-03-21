package bookmarks

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/mitchellh/go-homedir"

	"github.com/sosedoff/pgweb/pkg/command"
	"github.com/sosedoff/pgweb/pkg/shared"
)

type Bookmark struct {
	Url      string         `json:"url"`      // Postgres connection URL
	Host     string         `json:"host"`     // Server hostname
	Port     int            `json:"port"`     // Server port
	User     string         `json:"user"`     // Database user
	Password string         `json:"password"` // User password
	Database string         `json:"database"` // Database name
	Ssl      string         `json:"ssl"`      // Connection SSL mode
	Ssh      shared.SSHInfo `json:"ssh"`      // SSH tunnel config
}

func (b Bookmark) SSHInfoIsEmpty() bool {
	return b.Ssh.User == "" && b.Ssh.Host == "" && b.Ssh.Port == ""
}

func (b Bookmark) ConvertToOptions() command.Options {
	return command.Options{
		Url:    b.Url,
		Host:   b.Host,
		Port:   b.Port,
		User:   b.User,
		Pass:   b.Password,
		DbName: b.Database,
		Ssl:    b.Ssl,
	}
}

func readServerConfig(path string) (Bookmark, error) {
	bookmark := Bookmark{}

	buff, err := ioutil.ReadFile(path)
	if err != nil {
		return bookmark, err
	}

	_, err = toml.Decode(string(buff), &bookmark)

	if bookmark.Port == 0 {
		bookmark.Port = 5432
	}

	return bookmark, err
}

func fileBasename(path string) string {
	filename := filepath.Base(path)
	return strings.Replace(filename, filepath.Ext(path), "", 1)
}

func Path(overrideDir string) string {
	if overrideDir == "" {
		path, _ := homedir.Dir()
		return fmt.Sprintf("%s/.pgweb/bookmarks", path)
	}

	return overrideDir
}

func ReadAll(path string) (map[string]Bookmark, error) {
	results := map[string]Bookmark{}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return results, err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".toml" {
			continue
		}

		fullPath := path + "/" + file.Name()
		key := fileBasename(file.Name())
		config, err := readServerConfig(fullPath)

		if err != nil {
			fmt.Printf("%s parse error: %s\n", fullPath, err)
			continue
		}

		results[key] = config
	}

	return results, nil
}

func GetBookmark(bookmarkPath string, bookmarkName string) (Bookmark, error) {
	bookmarks, err := ReadAll(bookmarkPath)
	if err != nil {
		return Bookmark{}, err
	}

	bookmark, ok := bookmarks[bookmarkName]
	if !ok {
		return Bookmark{}, fmt.Errorf("couldn't find a bookmark with name %s", bookmarkName)
	}

	return bookmark, nil
}
