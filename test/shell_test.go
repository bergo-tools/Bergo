package test

import (
	"bergo/utils"
	"testing"
)

func TestShellTask(t *testing.T) {
	s := utils.Shell{
		IsTask: false,
	}
	output, err := s.Run("ls -l")
	if err != nil {
		t.Errorf("Run command failed: %v", err)
	}
	if output == "" {
		t.Errorf("Run command failed: empty output")
	}

}
