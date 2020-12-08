package msg 

import(
  "fmt"
  gov "github.com/cosmos/cosmos-sdk/x/gov/types"
  mt "gitlab.com/microtick/mtzone/x/microtick/types"
)

const (
  ProposalTypeDenomChange string = "MicrotickDenomChange"
)

func NewDenomChangeProposal(title, description, extDenom string, extPerInt int64) gov.Content {
  return &DenomChangeProposal{
    Title: title, 
    Description: description, 
    ExtDenom: extDenom,
    ExtPerInt: extPerInt,
  }
}

var _ gov.Content = &DenomChangeProposal{}

func init() {
  gov.RegisterProposalType(ProposalTypeDenomChange)
  gov.RegisterProposalTypeCodec(&DenomChangeProposal{}, "microtick/DenomChangeProposal")
}

func (dcp *DenomChangeProposal) ProposalRoute() string  { return mt.ModuleName }
func (dcp *DenomChangeProposal) ProposalType() string   { return ProposalTypeDenomChange }
func (dcp *DenomChangeProposal) ValidateBasic() error {
  return gov.ValidateAbstract(dcp)
}

func (dcp *DenomChangeProposal) String() string {
  return fmt.Sprintf(`Denom Change Proposal:
  Title: %s
  Description: %s
  Ext Denom: %s
  Ratio: %d
`, dcp.Title, dcp.Description, dcp.ExtDenom, dcp.ExtPerInt)
}
