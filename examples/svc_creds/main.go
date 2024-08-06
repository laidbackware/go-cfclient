package main

import (
	"context"
	"fmt"
	"os"

	"github.com/laidbackware/go-cfclient/v3/client"
	"github.com/laidbackware/go-cfclient/v3/config"
	"github.com/laidbackware/go-cfclient/v3/resource"
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

	bindings, err := cf.ServiceCredentialBindings.ListAll(ctx, nil)
	if err != nil {
		return err
	}
	for _, b := range bindings {
		fmt.Printf("GUID=%s, App=%s\n", b.GUID, b.Relationships.App.Data.GUID)
		details, err := cf.ServiceCredentialBindings.GetDetails(ctx, b.GUID)
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", details.Credentials)
		params, err := cf.ServiceCredentialBindings.GetParameters(ctx, b.GUID)
		if resource.IsServiceFetchBindingParametersNotSupportedError(err) {
			fmt.Println(err.(resource.CloudFoundryError).Detail)
		} else if err != nil {
			return err
		} else {
			fmt.Printf("%v\n", params)
		}
	}

	return nil
}
