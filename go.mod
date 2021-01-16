module gitlab.com/microtick/mtzone

go 1.15

require (
	github.com/btcsuite/btcd v0.21.0-beta // indirect
	github.com/btcsuite/btcutil v1.0.2 // indirect
	github.com/cosmos/cosmos-sdk v0.40.0
	github.com/ethereum/go-ethereum v1.9.25 // indirect
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.4.3
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/pkg/errors v0.9.1
	github.com/rakyll/statik v0.1.7 // indirect
	github.com/regen-network/cosmos-proto v0.3.0
	github.com/spf13/cast v1.3.1
	github.com/spf13/cobra v1.1.1
	github.com/stretchr/testify v1.6.1 // indirect
	github.com/tendermint/tendermint v0.34.1
	github.com/tendermint/tm-db v0.6.3
	github.com/vektra/mockery v1.1.2
	golang.org/x/tools v0.0.0-20200616133436-c1934b75d054
	google.golang.org/genproto v0.0.0-20201214200347-8c77b98c765d
	google.golang.org/grpc v1.33.2
	k8s.io/code-generator v0.20.2
	sigs.k8s.io/kind v0.9.0
)

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.2-alpha.regen.4
