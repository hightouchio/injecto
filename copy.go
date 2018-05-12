package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"path"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
)

var (
	copySubdirPaths = map[string]struct{}{
		"etc": struct{}{},
	}
)

func copy(cli *client.Client, dir, container string) error {
	totalCopyCount, err := getTotalCopyCount(dir, "")
	if err != nil {
		return err
	}
	_, err = copyRec(cli, dir, container, "", 0, totalCopyCount)
	return err
}

func copyRec(cli *client.Client, dir, container, prefix string, count, totalCopyCount int) (int, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return 0, err
	}

	for _, f := range files {
		name := path.Join(prefix, f.Name())
		fullname := path.Join(dir, f.Name())

		if _, ok := copySubdirPaths[name]; ok {
			partialCount, err := copyRec(cli, path.Join(dir, name), container, name, count, totalCopyCount)
			if err != nil {
				return 0, err
			}
			count += partialCount
			continue
		}

		dstInfo := archive.CopyInfo{Path: name}

		srcInfo, err := archive.CopyInfoSourcePath(fullname, true)
		if err != nil {
			return 0, err
		}

		srcArchive, err := archive.TarResource(srcInfo)
		if err != nil {
			return 0, err
		}

		dstDir, preparedArchive, err := archive.PrepareArchiveCopy(srcArchive, srcInfo, dstInfo)
		if err != nil {
			return 0, err
		}

		fmt.Printf("copying [%d/%d]: %s\n", count, totalCopyCount, name)
		if err := cli.CopyToContainer(context.Background(), container, dstDir,
			preparedArchive, types.CopyToContainerOptions{
				AllowOverwriteDirWithFile: true,
			}); err != nil {
			return 0, err
		}

		count++

		preparedArchive.Close()
		srcArchive.Close()
	}

	return count, nil
}

func getTotalCopyCount(dir, prefix string) (int, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return 0, err
	}

	totalCount := 0
	for _, f := range files {
		name := path.Join(prefix, f.Name())

		if _, ok := copySubdirPaths[name]; ok {
			partialCount, err := getTotalCopyCount(path.Join(dir, name), name)
			if err != nil {
				return 0, err
			}
			totalCount += partialCount
			continue
		}

		totalCount++
	}

	return totalCount, nil
}
