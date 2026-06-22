module github.com/squat/schemasonnet

go 1.25.0

require (
	github.com/gobuffalo/here v0.6.7
	github.com/google/jsonschema-go v0.4.2
	github.com/jsonnet-libs/docsonnet v0.0.6
	github.com/markbates/pkger v0.15.1
	github.com/spf13/cobra v1.10.2
)

require (
	github.com/campoy/embedmd v1.0.0 // indirect
	github.com/google/go-jsonnet v0.21.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)

tool github.com/campoy/embedmd

replace github.com/jsonnet-libs/docsonnet => github.com/squat/docsonnet v0.0.0-20260323023749-7f0aee931ffa
