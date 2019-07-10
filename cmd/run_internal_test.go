package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/whalebrew/whalebrew/packages"
)

func TestShouldBind(t *testing.T) {
	wd, err := os.Getwd()
	assert.NoError(t, err)
	t.Run("with a file that exists", func(t *testing.T) {
		t.Run("when not skipping missing volumes", func(t *testing.T) {
			bind, err := shouldBind(filepath.Join(wd, "run.go"), &packages.Package{SkipMissingVolumes: false})
			assert.NoError(t, err)
			assert.True(t, bind)
		})
		t.Run("when skipping missing volumes", func(t *testing.T) {
			bind, err := shouldBind(filepath.Join(wd, "run.go"), &packages.Package{SkipMissingVolumes: true})
			assert.NoError(t, err)
			assert.True(t, bind)
		})
	})
	t.Run("with a file that does not exists", func(t *testing.T) {
		t.Run("when not skipping missing volumes", func(t *testing.T) {
			bind, err := shouldBind(filepath.Join(wd, "thisFileShouldNotExist.go"), &packages.Package{SkipMissingVolumes: false})
			assert.Error(t, err)
			assert.False(t, bind)
		})
		t.Run("when skipping missing volumes", func(t *testing.T) {
			bind, err := shouldBind(filepath.Join(wd, "thisFileShouldNotExist.go"), &packages.Package{SkipMissingVolumes: true})
			assert.NoError(t, err)
			assert.False(t, bind)
		})
		t.Run("when mounting missing volumes", func(t *testing.T) {
			bind, err := shouldBind(filepath.Join(wd, "thisFileShouldNotExist.go"), &packages.Package{MountMissingVolumes: true})
			assert.NoError(t, err)
			assert.True(t, bind)
		})
	})
}

func TestAppendVolumes(t *testing.T) {
	wd, err := os.Getwd()
	assert.NoError(t, err)
	volumes := []string{
		filepath.Join(wd, "thisFileShouldNotExist.go:/notExists"),
		filepath.Join(wd, "run.go:/exists"),
		"~/:/home/user",
	}
	t.Run("when non existing volumes should be skept", func(t *testing.T) {
		args, err := appendVolumes(nil, &packages.Package{SkipMissingVolumes: true, Volumes: volumes})
		assert.NoError(t, err)
		assert.NotNil(t, args)
		assert.Len(t, args, 4)
		t.Run("all existing volumes instructed to mount on command line", func(t *testing.T) {
			assert.Equal(t, []string{"-v", fmt.Sprintf("%s/run.go:/exists", wd), "-v"}, args[:3])
			if !strings.HasSuffix(args[3], "/:/home/user") {
				t.Errorf("home volume should be mounted")
			}
			if strings.HasPrefix(args[3], "~") {
				t.Errorf("~/ prefix should be replaced by current user home directory")
			}
		})
	})
	t.Run("when non existing volumes should be mounted", func(t *testing.T) {
		args, err := appendVolumes(nil, &packages.Package{MountMissingVolumes: true, Volumes: volumes})
		assert.NoError(t, err)
		assert.NotNil(t, args)
		assert.Len(t, args, 6)
		t.Run("all existing volumes instructed to mount on command line", func(t *testing.T) {
			assert.Equal(t, []string{"-v", fmt.Sprintf("%s/thisFileShouldNotExist.go:/notExists", wd), "-v", fmt.Sprintf("%s/run.go:/exists", wd), "-v"}, args[:5])
			if !strings.HasSuffix(args[5], "/:/home/user") {
				t.Errorf("home volume should be mounted")
			}
			if strings.HasPrefix(args[5], "~") {
				t.Errorf("~/ prefix should be replaced by current user home directory")
			}
		})
	})
	t.Run("when non existing volumes should not be skept", func(t *testing.T) {
		args, err := appendVolumes(nil, &packages.Package{SkipMissingVolumes: false, Volumes: volumes})
		assert.Error(t, err)
		assert.Nil(t, args)
	})
}

func TestParseRuntimeVolumes(t *testing.T) {
	pkg := &packages.Package{
		PathArguments: []string{"C", "exec-path", "X", "stream"},
	}
	parseRuntimeVolumes([]string{"-C/tmp", "--other", "arg", "--exec-path", "/some/path", "-C", "/other/path", "--exec-path=local", "--stream", "-"}, pkg)
	wd, err := os.Getwd()
	assert.NoError(t, err)
	assert.Equal(t, []string{"/tmp:/tmp", "/other/path:/other/path", "/some/path:/some/path", fmt.Sprintf("%s/local:%s/local", wd, wd)}, pkg.Volumes)
}
