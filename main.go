package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
)

var blacklist = map[string]struct{}{
	"dev":                struct{}{},
	"etc/hostname":       struct{}{},
	"etc/hosts":          struct{}{},
	"etc/motd":           struct{}{},
	"etc/modules-load.d": struct{}{},
	"etc/mtab":           struct{}{},
}

type manifestEntry struct {
	Layers []string
}

func main() {
	image := os.Args[1]
	container := os.Args[2]

	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	fmt.Printf("saving %s\n", image)

	reader, err := cli.ImageSave(context.Background(), []string{image})
	if err != nil {
		panic(err)
	}

	saveDir, err := ioutil.TempDir("", "")
	if err != nil {
		panic(err)
	}

	if err := extractTar(reader, saveDir); err != nil {
		panic(err)
	}

	manifestBytes, err := ioutil.ReadFile(path.Join(saveDir, "manifest.json"))
	if err != nil {
		panic(err)
	}

	var manifest []manifestEntry
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		panic(err)
	}
	if len(manifest) != 1 {
		return
	}
	if len(manifest[0].Layers) != 1 {
		return
	}

	layerDir, err := ioutil.TempDir("", "")
	if err != nil {
		panic(err)
	}

	for i, layer := range manifest[0].Layers {
		filename := path.Join(saveDir, layer)

		layerFile, err := os.Open(filename)
		if err != nil {
			panic(err)
		}

		fmt.Printf("extracting layer [%d/%d]\n", i+1, len(manifest[0].Layers))
		if err := extractTar(layerFile, layerDir); err != nil {
			panic(err)
		}
	}

	if err := copy(cli, layerDir, "", container); err != nil {
		panic(err)
	}
}

func copy(cli *client.Client, dir, prefix, container string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for i, f := range files {
		name := path.Join(prefix, f.Name())
		fullname := path.Join(dir, f.Name())

		if _, ok := blacklist[name]; ok {
			continue
		}
		if name == "etc" {
			if err := copy(cli, path.Join(dir, name), name, container); err != nil {
				return err
			}
			continue
		}

		dstInfo := archive.CopyInfo{Path: name}

		srcInfo, err := archive.CopyInfoSourcePath(fullname, true)
		if err != nil {
			return err
		}

		srcArchive, err := archive.TarResource(srcInfo)
		if err != nil {
			return err
		}

		dstDir, preparedArchive, err := archive.PrepareArchiveCopy(srcArchive, srcInfo, dstInfo)
		if err != nil {
			return err
		}

		fmt.Printf("copying [%d/%d]: %s\n", i, len(files), name)
		if err := cli.CopyToContainer(context.Background(), container, dstDir,
			preparedArchive, types.CopyToContainerOptions{
				AllowOverwriteDirWithFile: true,
			}); err != nil {
			return err
		}

		preparedArchive.Close()
		srcArchive.Close()
	}

	return nil
}

func extractGzipTar(reader io.Reader, target string) error {
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	return extractTar(tar.NewReader(gzipReader), target)
}

func extractTar(reader io.Reader, target string) error {
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
