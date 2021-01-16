// +build tools

package tools

//nolint
import (
	_ "github.com/vektra/mockery"
	_ "golang.org/x/tools/cmd/stringer"
	_ "k8s.io/code-generator"
	_ "sigs.k8s.io/kind"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
	_ "github.com/regen-network/cosmos-proto/protoc-gen-gocosmos"
	_ "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway"
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	_ "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger"
)
