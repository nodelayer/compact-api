package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
)

func getNodejsLTSVersions() []string {
	v := []string{}
	r, e := http.Get(NODEJS_VERSION_INDEX_PATH)

	if e != nil {
		return v
	}

	defer r.Body.Close()

	var l []struct {
		LTS     any    `json:"lts"`
		Version string `json:"version"`
	}

	if e = json.NewDecoder(r.Body).Decode(&l); e != nil {
		return v
	}

	for _, i := range l {
		if _, ok := i.LTS.(string); ok {
			v = append(v, i.Version)
		}
	}

	return v
}

func downloadNodejsArchive(version string) (string, error) {
	if VersionRegexp.MatchString(version) == false {
		return "", errors.New("version does not satisfy regex")
	}

	o, arch := runtime.GOOS, runtime.GOARCH

	if arch == "amd64" {
		arch = "x64"
	}

	dir, e := os.MkdirTemp("", fmt.Sprintf("nodejs-%s-*", version))
	if e != nil {
		return "", e
	}

	localGzPath := fmt.Sprintf("%s/archive.tar.gz", dir)

	f, e := os.OpenFile(localGzPath, os.O_RDWR|os.O_CREATE, 0777)
	if e != nil {
		return "", e
	}

	defer f.Close()

	remoteGzPath := fmt.Sprintf("%[1]s/node-%[1]s-%[2]s-%[3]s.tar.gz", version, o, arch)

	r, e := http.Get(fmt.Sprintf("%s/%s", NODEJS_DIST_PATH, remoteGzPath))
	if e != nil {
		return "", e
	}

	defer r.Body.Close()

	if _, e = io.Copy(f, r.Body); e != nil {
		return "", e
	}

	return localGzPath, nil
}

func detectNodejs(version string) (string, error) {
	files, e := os.ReadDir(os.TempDir())
	if e != nil {
		return "", e
	}

	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "x64"
	}

	for _, f := range files {
		if f.IsDir() && strings.HasPrefix(f.Name(), fmt.Sprintf("nodejs-%s", version)) {
			potentialNpmPath := path.Join(os.TempDir(), f.Name(), fmt.Sprintf("node-%s-%s-%s", version, runtime.GOOS, arch), "bin", "npm")

			if _f, e := os.Open(potentialNpmPath); e == nil && f != nil {
				_f.Close()
				return potentialNpmPath, nil
			}
		}
	}

	return "", errors.New("nodejs version not detected on host")
}

func installNodejs(version string) (string, error) {
	if npmPath, e := detectNodejs(version); e == nil && npmPath != "" {
		return npmPath, nil
	}

	gzPath, e := downloadNodejsArchive(version)
	if e != nil {
		return "", e
	}

	if e = os.Chdir(path.Dir(gzPath)); e != nil {
		return "", e
	}

	if e = exec.Command("tar", "-xzf", gzPath).Run(); e != nil {
		return "", e
	}

	files, e := os.ReadDir(".")
	if e != nil {
		return "", e
	}

	var fptr *fs.DirEntry

	for _, f := range files {
		if f.IsDir() && strings.HasPrefix(f.Name(), "node-v") {
			fptr = &f
			break
		}
	}

	if fptr == nil {
		return "", errors.New("tar extraction failed")
	}

	npmPath := fmt.Sprintf("%s/%s/bin/npm", path.Dir(gzPath), (*fptr).Name())

	return npmPath, nil
}
