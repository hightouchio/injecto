package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/docker/docker/client"
	"github.com/hightouchio/injecto/tar"
)

var (
	blacklist = []string{
		"dev",
		"etc/hostname",
		"etc/hosts",
		"etc/motd",
		"etc/modules-load.d",
		"etc/mtab",
		"etc/resolv.conf",
		"media",
		"mnt",
		"sys",
		"tmp",
	}
)

type manifestEntry struct {
	Layers []string
}

func save(cli *client.Client, dir, image string) error {
	fmt.Printf("saving %s\n", image)

	reader, err := cli.ImageSave(context.Background(), []string{image})
	if err != nil {
		return err
	}

	saveDir, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}

	if err := tar.Extract(reader, saveDir, blacklist); err != nil {
		return err
	}

	manifestBytes, err := ioutil.ReadFile(path.Join(saveDir, "manifest.json"))
	if err != nil {
		return err
	}

	var manifest []manifestEntry
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return err
	}

	for i, layer := range manifest[0].Layers {
		filename := path.Join(saveDir, layer)

		layerFile, err := os.Open(filename)
		if err != nil {
			return err
		}

		fmt.Printf("extracting layer [%d/%d]\n", i+1, len(manifest[0].Layers))
		if err := tar.Extract(layerFile, dir, blacklist); err != nil {
			return err
		}
	}

	return nil
}
