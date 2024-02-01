package main

import (
	"strings"
	"testing"

	"gotest.tools/assert"
)

func TestConfigYaml(t *testing.T) {
	var c Config
	configText := `
cli: ["prog", "arg1", "arg2", "arg3"]
mount:
  - src: src1
    dst: dst1
    perm: perm1
  - src: src2
    dst: dst2
    perm: perm2
`

	err := c.parseYAML(strings.NewReader(configText))
	if err != nil {
		t.Error(err)
	}

	assert.DeepEqual(t, c.TargetCli, []string{"prog", "arg1", "arg2", "arg3"})

	mounts := []Mount{
		{"src1", "dst1", "perm1"},
		{"src2", "dst2", "perm2"},
	}
	assert.DeepEqual(t, c.MountPoints, mounts)
}
