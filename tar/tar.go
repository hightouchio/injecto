package tar

import (
	"archive/tar"
	"io"
	"os"
	"path"
	"strings"
)

func Extract(reader io.Reader, target string, blacklist []string) error {
	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		filename := header.Name
		filename = path.Join(target, header.Name)

		skip := false
		for _, prefix := range blacklist {
			if strings.HasPrefix(header.Name, prefix) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err = os.MkdirAll(filename, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if _, err := os.Stat(filename); err == nil {
				if err := os.Remove(filename); err != nil {
					return err
				}
			}
			writer, err := os.Create(filename)
			if err != nil {
				return err
			}
			io.Copy(writer, tarReader)
			if err = os.Chmod(filename, header.FileInfo().Mode()); err != nil {
				return err
			}
			writer.Close()
		case tar.TypeLink:
			if _, err := os.Stat(filename); err == nil {
				if err := os.Remove(filename); err != nil {
					return err
				}
			}
			if err := os.Link(header.Linkname, filename); err != nil {
				return err
			}
		case tar.TypeSymlink:
			if _, err := os.Stat(filename); err == nil {
				if err := os.Remove(filename); err != nil {
					return err
				}
			}
			if err := os.Symlink(header.Linkname, filename); err != nil {
				return err
			}
		}
	}

	return nil
}
