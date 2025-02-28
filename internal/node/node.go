package node

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"slices"
	"strings"
)

const (
	NODEJS_DIST_PATH          = "https://nodejs.org/dist"
	NODEJS_VERSION_INDEX_PATH = "https://nodejs.org/dist/index.json"
)

type Container struct {
	Version string
	NpmPath string
	WorkDir string
}

var (
	DefaultVersion = "v22.12.0"
	VersionRegexp  *regexp.Regexp
)

func init() {
	VersionRegexp = regexp.MustCompile(`^v\d{1,3}\.\d{1,3}\.\d{1,3}$`)

	ltsVersions := Versions()

	if len(ltsVersions) > 0 {
		DefaultVersion = ltsVersions[0]

		MustInstall(DefaultVersion)
	}
}

func Versions() []string {
	return getNodejsLTSVersions()
}

func MustInstall(version string) (string, error) {
	return installNodejs(version)
}

func New(version string) (Container, error) {
	if version == "" {
		version = DefaultVersion
	}

	var (
		e error
		c Container
	)

	c.Version = version
	if c.NpmPath, e = MustInstall(version); e == nil {
		c.WorkDir, e = os.MkdirTemp("", fmt.Sprintf("nodelayer-*"))
	}

	return c, e
}

func (c *Container) InstallPackages(packages []string) error {
	if e := os.Chdir(c.WorkDir); e != nil {
		return e
	}

	if e := os.Mkdir("nodejs", 0700); e != nil {
		return e
	}

	if e := os.Chdir(path.Join(c.WorkDir, "nodejs")); e != nil {
		return e
	}

	cmd := exec.Command(c.NpmPath, "init", "-y")
	prependPathEnv(cmd, path.Dir(c.NpmPath))

	if out, e := cmd.CombinedOutput(); e != nil {
		if len(out) > 0 {
			return errors.New(string(out))
		}

		return e
	}

	args := append([]string{"install", "--save"}, packages...)

	cmd = exec.Command(c.NpmPath, args...)
	prependPathEnv(cmd, path.Dir(c.NpmPath))

	if out, e := cmd.CombinedOutput(); e != nil {
		if len(out) > 0 {
			return errors.New(string(out))
		}

		return e
	}

	return nil
}

func (c *Container) CreateArchive() (string, error) {
	if e := os.Chdir(c.WorkDir); e != nil {
		return "", e
	}

	cmd := exec.Command("zip", "-r", "layer.zip", "nodejs")
	prependPathEnv(cmd, path.Dir(c.NpmPath))

	if out, e := cmd.CombinedOutput(); e != nil {
		if len(out) > 0 {
			return "", errors.New(string(out))
		}

		return "", e
	}

	layerPath := path.Join(c.WorkDir, "layer.zip")

	return layerPath, nil
}

func prependPathEnv(cmd *exec.Cmd, binPath string) {
	ups := false
	cmd.Env = slices.Clone(os.Environ())

	for i, env := range cmd.Env {
		if strings.HasPrefix(env, "PATH=") {
			ups = true
			cmd.Env[i] = strings.Replace(env, "PATH=", fmt.Sprintf("PATH=%s:", binPath), 1)
			break
		}
	}

	if ups == false {
		cmd.Env = append(cmd.Env, fmt.Sprintf("PATH=%s:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin", binPath))
	}
}
