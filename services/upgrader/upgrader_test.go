package upgrader

import (
	"archive/tar"
	"compress/gzip"
	"io/ioutil"
	"os"
	"testing"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/stretchr/testify/require"
)

func TestCanIUpgradeToE0(t *testing.T) {
	t.Run("when you are already in an enterprise build", func(t *testing.T) {
		buildEnterprise := model.BuildEnterpriseReady
		model.BuildEnterpriseReady = "true"
		defer func() {
			model.BuildEnterpriseReady = buildEnterprise
		}()
		require.Error(t, CanIUpgradeToE0())
	})

	t.Run("when you are not in an enterprise build", func(t *testing.T) {
		buildEnterprise := model.BuildEnterpriseReady
		model.BuildEnterpriseReady = "false"
		defer func() {
			model.BuildEnterpriseReady = buildEnterprise
		}()
		require.NoError(t, CanIUpgradeToE0())
	})
}

func TestGetCurrentVersionTgzUrl(t *testing.T) {
	currentVersion := model.CurrentVersion
	model.CurrentVersion = "5.22.0"
	defer func() {
		model.CurrentVersion = currentVersion
	}()
	require.Equal(t, "https://releases.mattermost.com/5.22.0/mattermost-5.22.0-linux-amd64.tar.gz", getCurrentVersionTgzUrl())
}

func TestExtractBinary(t *testing.T) {
	t.Run("extract from empty file", func(t *testing.T) {
		tmpMockTarGz, err := ioutil.TempFile("", "mock_tgz")
		require.Nil(t, err)
		defer os.Remove(tmpMockTarGz.Name())
		tmpMockTarGz.Close()

		tmpMockExecutable, err := ioutil.TempFile("", "mock_exe")
		require.Nil(t, err)
		defer os.Remove(tmpMockExecutable.Name())
		tmpMockExecutable.Close()

		extractBinary(tmpMockExecutable.Name(), tmpMockTarGz.Name())
	})

	t.Run("extract from empty tar.gz file", func(t *testing.T) {
		tmpMockTarGz, err := ioutil.TempFile("", "mock_tgz")
		require.Nil(t, err)
		defer os.Remove(tmpMockTarGz.Name())
		gz := gzip.NewWriter(tmpMockTarGz)
		tw := tar.NewWriter(gz)
		tw.Close()
		gz.Close()
		tmpMockTarGz.Close()

		tmpMockExecutable, err := ioutil.TempFile("", "mock_exe")
		require.Nil(t, err)
		defer os.Remove(tmpMockExecutable.Name())
		tmpMockExecutable.Close()

		require.Error(t, extractBinary(tmpMockExecutable.Name(), tmpMockTarGz.Name()))
	})

	t.Run("extract from tar.gz without mattermost/bin/mattermost file", func(t *testing.T) {
		tmpMockTarGz, err := ioutil.TempFile("", "mock_tgz")
		require.Nil(t, err)
		defer os.Remove(tmpMockTarGz.Name())
		gz := gzip.NewWriter(tmpMockTarGz)
		tw := tar.NewWriter(gz)

		tw.WriteHeader(&tar.Header{
			Typeflag: tar.TypeReg,
			Name:     "test-filename",
			Size:     4,
		})
		tw.Write([]byte("test"))

		gz.Close()
		tmpMockTarGz.Close()

		tmpMockExecutable, err := ioutil.TempFile("", "mock_exe")
		require.Nil(t, err)
		defer os.Remove(tmpMockExecutable.Name())
		tmpMockExecutable.Close()

		require.Error(t, extractBinary(tmpMockExecutable.Name(), tmpMockTarGz.Name()))
	})

	t.Run("extract from tar.gz with mattermost/bin/mattermost file", func(t *testing.T) {
		tmpMockTarGz, err := ioutil.TempFile("", "mock_tgz")
		require.Nil(t, err)
		defer os.Remove(tmpMockTarGz.Name())
		gz := gzip.NewWriter(tmpMockTarGz)
		tw := tar.NewWriter(gz)

		tw.WriteHeader(&tar.Header{
			Typeflag: tar.TypeReg,
			Name:     "mattermost/bin/mattermost",
			Size:     4,
		})
		tw.Write([]byte("test"))

		gz.Close()
		tmpMockTarGz.Close()

		tmpMockExecutable, err := ioutil.TempFile("", "mock_exe")
		require.Nil(t, err)
		defer os.Remove(tmpMockExecutable.Name())
		tmpMockExecutable.Close()

		require.NoError(t, extractBinary(tmpMockExecutable.Name(), tmpMockTarGz.Name()))
		tmpMockExecutableAfter, err := os.Open(tmpMockExecutable.Name())
		require.NoError(t, err)
		defer tmpMockExecutableAfter.Close()
		bytes, err := ioutil.ReadAll(tmpMockExecutableAfter)
		require.NoError(t, err)
		require.Equal(t, []byte("test"), bytes)
	})
}
