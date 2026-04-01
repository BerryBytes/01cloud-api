package helper

import (
	"archive/tar"
	"compress/gzip"
	b64 "encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"

	"github.com/docker/docker/pkg/archive"
	log "github.com/sirupsen/logrus"
)

func CopyDirectory(scrDir, dest string) error {
	entries, err := os.ReadDir(scrDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(scrDir, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		fileInfo, err := os.Stat(sourcePath)
		if err != nil {
			return err
		}

		stat, ok := fileInfo.Sys().(*syscall.Stat_t)
		if !ok {
			return fmt.Errorf("failed to get raw syscall.Stat_t data for '%s'", sourcePath)
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err := CreateIfNotExists(destPath, 0755); err != nil {
				return err
			}
			if err := CopyDirectory(sourcePath, destPath); err != nil {
				return err
			}
		case os.ModeSymlink:
			if err := CopySymLink(sourcePath, destPath); err != nil {
				return err
			}
		default:
			if err := Copy(sourcePath, destPath); err != nil {
				return err
			}
		}

		if err := os.Lchown(destPath, int(stat.Uid), int(stat.Gid)); err != nil {
			return err
		}
		isSymlink := fileInfo.Mode()&os.ModeSymlink != 0
		if !isSymlink {
			if err := os.Chmod(destPath, fileInfo.Mode()); err != nil {
				return err
			}
		}
	}
	return nil
}

func CreateTarFile(path string, tarFile string) error {
	log.Debug("Creating tar file :: ", path)
	tarStream, err := archive.Tar(path, archive.Gzip)
	if err != nil {
		log.Errorf("Error when creating tar stream: %s", err)
		return err
	}
	out, err := os.Create(tarFile)
	if err != nil {
		log.Error("Error create tar")
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, tarStream)
	if err != nil {
		log.Errorf("Error reading tar stream: %s", err)
		return err
	}
	return nil
}

func Copy(srcFile, dstFile string) error {
	// Create the destination directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dstFile), 0755); err != nil {
		return err
	}

	out, err := os.Create(dstFile)
	if err != nil {
		return err
	}

	defer out.Close()

	in, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer in.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return nil
}

func Exists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}

	return true
}
func ExtractTarGz(gzipStream io.Reader, path string) {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		log.Error("ExtractTarGz: NewReader failed")
		return
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0775); err != nil {
			log.Error(err)
			return
		}
	}

	tarReader := tar.NewReader(uncompressedStream)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Errorf("ExtractTarGz: Next() failed: %s", err.Error())
			return
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(fmt.Sprintf("%s/%s", path, header.Name), 0755); err != nil {
				log.Errorf("ExtractTarGz: Mkdir() failed: %s", err.Error())
			}
		case tar.TypeReg:
			outFile, err := os.Create(fmt.Sprintf("%s/%s", path, header.Name))
			if err != nil {
				log.Errorf("ExtractTarGz: Create() failed: %s", err.Error())
				return
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				log.Errorf("ExtractTarGz: Copy() failed: %s", err.Error())
				return
			}
			outFile.Close()

		default:
			log.Errorf(
				"ExtractTarGz: uknown type: %v in %s",
				header.Typeflag,
				header.Name)
		}
	}
}

func CreateIfNotExists(dir string, perm os.FileMode) error {
	if Exists(dir) {
		return nil
	}

	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("failed to create directory: '%s', error: '%s'", dir, err.Error())
	}

	return nil
}

func CopySymLink(source, dest string) error {
	link, err := os.Readlink(source)
	if err != nil {
		return err
	}
	return os.Symlink(link, dest)
}

func ReadFileInB64String(path string) string {
	cred, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return b64.StdEncoding.EncodeToString(cred)
}

func ReadFileInString(path string) string {
	cred, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(cred)
}
