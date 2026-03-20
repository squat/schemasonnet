package schema

import (
	"embed"

	"github.com/gobuffalo/here"
	"github.com/markbates/pkger"

	"github.com/squat/schemasonnet/embedpkging"
)

// fs holds the Jsonnet library files that docsonnet loads at runtime via pkger.Open.
// In order to avoid using the deprecated pkger project, we use the upstream embed
// package and provide a compatibility layer to translate between fs.FS and pkger.Pkger.
//
//go:embed load.libsonnet doc-util
var fs embed.FS

func init() {
	info := here.Info{
		ImportPath: "github.com/squat/schemasonnet",
		Name:       "schemasonnet",
		Module: here.Module{
			Path: "github.com/squat/schemasonnet",
		},
	}
	if err := pkger.Apply(embedpkging.New(fs, info), nil); err != nil {
		panic(err.Error())
	}
}
