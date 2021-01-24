package client

import ( 
  "github.com/cosmos/cosmos-sdk/client"
 	govrest "github.com/cosmos/cosmos-sdk/x/gov/client/rest"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	
 	"gitlab.com/microtick/mtzone/x/microtick/client/cli"
)

// Microtick does not support proposals over REST, but SDK does not handle
// nil for the handler value
func DummyRESTHandler(clientCtx client.Context) govrest.ProposalRESTHandler {
  return govrest.ProposalRESTHandler {
    SubRoute: "n/a",
    Handler: nil,
  }
}

var DenomChangeProposalHandler = govclient.NewProposalHandler(cli.NewCmdSubmitDenomChangeProposal, DummyRESTHandler)
var AddMarketsProposalHandler = govclient.NewProposalHandler(cli.NewCmdSubmitAddMarketsProposal, DummyRESTHandler)
