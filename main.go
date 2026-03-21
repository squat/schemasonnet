package main

import (
	"fmt"
	"os"

	"github.com/jsonnet-libs/docsonnet/pkg/docsonnet"
	"github.com/spf13/cobra"

	"github.com/squat/schemasonnet/schema"
	"github.com/squat/schemasonnet/version"
)

func main() {
	cmd := &cobra.Command{
		Use:          "schemasonnet [file]",
		Short:        "Convert Docsonnet type annotations to JSON Schema",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		Version:      version.Version,
	}
	jpath := cmd.Flags().StringSliceP("jpath", "J", []string{"vendor"}, "Specify an additional library search dir (right-most wins)")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		pkg, err := docsonnet.Load(args[0], docsonnet.Opts{JPath: *jpath})
		if err != nil {
			return err
		}
		schema := schema.Convert(pkg)
		s, err := schema.MarshalJSON()
		if err != nil {
			return err
		}
		if _, err := os.Stdout.Write(s); err != nil {
			return err
		}
		return nil
	}

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
