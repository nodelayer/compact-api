package node

import (
	"os"
	"path"
	"testing"
)

func TestVersions(t *testing.T) {
	ltsVersions := Versions()
	if len(ltsVersions) < 20 {
		t.Errorf("expected lts versions to be at least 20 but got %d", len(ltsVersions))
	}
}

func TestMustInstall(t *testing.T) {
	if _, e := MustInstall("00.00.00"); e == nil {
		t.Errorf("expected MustInstall(00.00.00) to fail")
	}

	npmPath, e := MustInstall("v22.12.0")
	if e != nil {
		t.Errorf("expected MustInstall(v22.12.0) to complete successfully but got error: %s", e.Error())
		t.FailNow()
	}

	f, e := os.Open(npmPath)
	if e != nil {
		t.Errorf("expected %s to exist but got error: %s", npmPath, e.Error())
		t.FailNow()
	}

	defer f.Close()

	if s, e := f.Stat(); e == nil {
		if m := s.Mode(); (m & 0x0080) != 0x0080 {
			t.Errorf("expected %s to be executable", npmPath)
		}
	}
}

func TestNew(t *testing.T) {
	c, e := New("v22.12.0")
	if e != nil {
		t.Errorf("expected new container struct but got error: %s", e.Error())
		t.FailNow()
	}

	f, e := os.Open(c.WorkDir)
	if e != nil {
		t.Errorf("expected %s to exist but got error: %s", c.WorkDir, e.Error())
		t.FailNow()
	}

	defer f.Close()

	if s, e := f.Stat(); e == nil {
		if s.IsDir() != true {
			t.Errorf("expected %s to be a directory", c.WorkDir)
		}
	}
}

func TestInstallPackages(t *testing.T) {
	pkgs := []string{"is-odd"}
	c, _ := New("v22.12.0")

	e := c.InstallPackages(pkgs)
	if e != nil {
		t.Errorf("expected InstallPackages(is-odd) to completed successfully but got error: %s", e.Error())
		t.FailNow()
	}

	pkgPath := path.Join(c.WorkDir, "nodejs", "node_modules", "is-odd")

	f, e := os.Open(pkgPath)
	if e != nil {
		t.Errorf("expected %s to exist but got error: %s", pkgPath, e.Error())
		t.FailNow()
	}

	defer f.Close()

	if s, e := f.Stat(); e == nil {
		if s.IsDir() != true {
			t.Errorf("expected %s to be a directory", pkgPath)
		}
	}
}

func TestCreateArchive(t *testing.T) {
	pkgs := []string{"is-odd"}
	c, _ := New("v22.12.0")

	c.InstallPackages(pkgs)

	layerPath, e := c.CreateArchive()
	if e != nil {
		t.Errorf("expected CreateArchive() to completed successfully but got error: %s", e.Error())
		t.FailNow()
	}

	f, e := os.Open(layerPath)
	if e != nil {
		t.Errorf("expected %s to exist but got error: %s", layerPath, e.Error())
		t.FailNow()
	}

	defer f.Close()
}
