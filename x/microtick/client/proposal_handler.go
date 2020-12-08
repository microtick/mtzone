package client

import ( 
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
 	"gitlab.com/microtick/mtzone/x/microtick/client/cli"
)

var DenomChangeProposalHandler = govclient.NewProposalHandler(cli.NewCmdSubmitDenomChangeProposal, nil)
