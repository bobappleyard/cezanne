package buildutil

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

//go:embed cz.h
var headerFile []byte

type Space struct {
	dir string
	err error
}

func New() *Space {
	dir, err := os.MkdirTemp(os.TempDir(), "cz-build")
	s := &Space{dir: dir, err: err}

	s.WriteFile("cz.h", headerFile)

	return s
}

func (s *Space) Path(name string) string {
	return filepath.Join(s.dir, name)
}

func (s *Space) Err() error {
	return s.err
}

func (s *Space) SetErr(err error) {
	if s.err != nil {
		return
	}

	s.err = err
}

func (s *Space) WriteFile(name string, data []byte) {
	if s.err != nil {
		return
	}

	s.err = os.WriteFile(s.Path(name), data, 0600)
}

func (s *Space) Run(name string, args ...string) {
	if s.err != nil {
		return
	}

	cmd := exec.Command(name, args...)
	cmd.Dir = s.dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		s.err = fmt.Errorf("%s\n%w", string(out), err)
	}
}

func (s *Space) Compile(file string) {
	s.Run("gcc", "-c", file)
}

func (s *Space) Link(exe string, files []string) {
	s.Run("gcc", append([]string{"-o", exe}, files...)...)
}

func (s *Space) Package(name string, files []string) {
	if s.err != nil {
		return
	}

	s.Run("rm", "-f", name)

	for _, f := range files {
		s.Run("ar", "csr", name, f)
	}
}
