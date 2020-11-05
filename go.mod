module github.com/magic-pod/bitrise-step-magic-pod

go 1.13

require (
	github.com/Magic-Pod/magic-pod-api-client v0.0.0-20201105033644-ffea7384265d
	github.com/bitrise-io/go-steputils v0.0.0-20200227150459-94490ca44ddb // indirect
	github.com/bitrise-io/go-utils v0.0.0-20200224122728-e212188d99b4
	github.com/bitrise-tools/go-steputils v0.0.0-20200227150459-94490ca44ddb
	github.com/nwaples/rardecode v1.1.0 // indirect
	github.com/pierrec/lz4 v2.5.2+incompatible // indirect
	github.com/ulikunitz/xz v0.5.7 // indirect
	github.com/urfave/cli v1.22.2
	gopkg.in/resty.v1 v1.12.0 // indirect
)

replace github.com/go-resty/resty => gopkg.in/resty.v1 v1.11.0
