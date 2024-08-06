package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/laidbackware/go-cfclient/v3/client"
	"github.com/laidbackware/go-cfclient/v3/config"
)

func main() {
	err := execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Done!")
}

func execute() error {
	ctx := context.Background()
	conf, err := config.NewFromCFHome(config.SkipTLSValidation())
	if err != nil {
		return err
	}
	cf, err := client.New(conf)
	if err != nil {
		return err
	}

	droplets, err := cf.Droplets.ListAll(ctx, nil)
	if err != nil {
		return err
	}
	if len(droplets) < 1 {
		return errors.New("error listing droplets, expected at least one droplet")
	}
	droplet := droplets[0]

	reader, err := cf.Droplets.Download(ctx, droplet.GUID)
	if err != nil {
		return err
	}
	defer func() { _ = reader.Close() }()

	dropletFile, err := os.CreateTemp("", "droplet-*.zip")
	if err != nil {
		return err
	}
	defer func() { _ = dropletFile.Close() }()

	fmt.Printf("Writing droplet %s to %s\n", droplet.GUID, dropletFile.Name())
	_, err = io.Copy(dropletFile, reader)
	return err
}
