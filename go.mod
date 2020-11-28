module gitlab.com/microtick/mtzone

go 1.14

require (
	github.com/cosmos/cosmos-sdk v0.40.0-rc3
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.4.3
	github.com/google/gofuzz v1.1.1-0.20200604201612-c04b05f3adfa // indirect
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.15.2
	github.com/pkg/errors v0.9.1
	github.com/rakyll/statik v0.1.7
	github.com/regen-network/cosmos-proto v0.3.0
	github.com/spf13/cast v1.3.1
	github.com/spf13/cobra v1.1.1
	github.com/tendermint/tendermint v0.34.0-rc6
	github.com/tendermint/tm-db v0.6.2
	github.com/vektra/mockery v1.1.2
	golang.org/x/tools v0.0.0-20200616133436-c1934b75d054
	google.golang.org/genproto v0.0.0-20201014134559-03b6142f0dc9
	google.golang.org/grpc v1.33.0
	k8s.io/code-generator v0.19.4
	sigs.k8s.io/kind v0.9.0
)

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.2-alpha.regen.4
