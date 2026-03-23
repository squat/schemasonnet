package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

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
	extCode := cmd.Flags().StringArray("ext-code", nil, "Set code value of extVar (Format: key=<code>)")
	extCodeFile := cmd.Flags().StringArray("ext-code-file", nil, "Set code value of extVar from file (Format: key=filename)")
	extStr := cmd.Flags().StringArrayP("ext-str", "V", nil, "Set string value of extVar (Format: key=value)")
	extStrFile := cmd.Flags().StringArray("ext-str-file", nil, "Set string value of extVar from file (Format: key=filename)")

	tlaCode := cmd.Flags().StringArray("tla-code", nil, "Set code value of top level function (Format: key=<code>)")
	tlaCodeFile := cmd.Flags().StringArray("tla-code-file", nil, "Set code value of top level function from file (Format: key=filename)")
	tlaStr := cmd.Flags().StringArrayP("tla-str", "A", nil, "Set string value of top level function (Format: key=value)")
	tlaStrFile := cmd.Flags().StringArray("tla-str-file", nil, "Set string value of top level function from file (Format: key=filename)")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		extCodes, err := parseCodeFlags("ext", extCode, extStr, extCodeFile, extStrFile)
		if err != nil {
			return err
		}
		tlaCodes, err := parseCodeFlags("tla", tlaCode, tlaStr, tlaCodeFile, tlaStrFile)
		if err != nil {
			return err
		}
		pkg, err := docsonnet.Load(args[0], &docsonnet.Opts{JPath: *jpath, ExtCode: extCodes, TLACode: tlaCodes})
		if err != nil {
			return err
		}
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

func parseCodeFlags(kind string, code, str, codeFile, strFile *[]string) (map[string]string, error) {
	m := make(map[string]string)
	for _, s := range *code {
		split := strings.SplitN(s, "=", 2)
		if len(split) != 2 {
			return nil, fmt.Errorf(kind+"-code argument has wrong format: `%s`. Expected `key=<code>`", s)
		}
		m[split[0]] = split[1]
	}

	for _, s := range *str {
		split := strings.SplitN(s, "=", 2)
		if len(split) != 2 {
			return nil, fmt.Errorf(kind+"-str argument has wrong format: `%s`. Expected `key=<value>`", s)
		}
		// Properly quote the string; note that fmt.Sprintf("%q",...) could
		// produce \U escapes which are not valid Jsonnet.
		js, err := json.Marshal(split[1])
		if err != nil {
			return nil, fmt.Errorf("impossible: failed to convert string to JSON: %w", err)
		}
		m[split[0]] = string(js)
	}

	for _, x := range []struct {
		arg        *[]string
		kind2, imp string
	}{
		{arg: codeFile, kind2: "code", imp: "import"},
		{arg: strFile, kind2: "str", imp: "importstr"},
	} {
		for _, s := range *x.arg {
			split := strings.SplitN(s, "=", 2)
			if len(split) != 2 {
				return nil, fmt.Errorf("%s-%s-file argument has wrong format: `%s`. Expected `key=filename`", kind, x.kind2, s)
			}
			m[split[0]] = fmt.Sprintf(`%s @"%s"`, x.imp, strings.ReplaceAll(split[1], `"`, `""`))
		}
	}
	return m, nil
}
