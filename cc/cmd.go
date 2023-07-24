package cc

import (
	"fmt"
	"os/exec"
)

func Compile(target, source, include string) error {
	buf, err := exec.Command("gcc", "-c", "-I"+include, "-Wall", "-Werror", "-o", target, source).CombinedOutput()
	if err != nil {
		return fmt.Errorf("compile failed: %w", err)
	}
	if len(buf) != 0 {
		return fmt.Errorf("compile returned:\n%s", string(buf))
	}
	return nil
}

func Link(target string, sources []string) error {
	buf, err := exec.Command("gcc", append([]string{"-o", target}, sources...)...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("compile failed: %w\n%s", err, string(buf))
	}
	if len(buf) != 0 {
		return fmt.Errorf("compile returned:\n%s", string(buf))
	}
	return nil
}
