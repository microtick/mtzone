module gitlab.com/microtick/mtzone

go 1.12

require (
	github.com/cosmos/cosmos-sdk v0.34.4-0.20200519133235-d7677e087117
	github.com/gorilla/mux v1.7.4
	github.com/onsi/ginkgo v1.8.0 // indirect
	github.com/onsi/gomega v1.5.0 // indirect
	github.com/otiai10/copy v1.1.1
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.0
	github.com/tendermint/go-amino v0.15.1
	github.com/tendermint/tendermint v0.33.4
	github.com/tendermint/tm-db v0.5.1
)

replace github.com/cosmos/cosmos-sdk => github.com/microtick/cosmos-sdk v0.34.4-0.20210306204855-c554b4494f47
