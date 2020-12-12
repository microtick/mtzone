package msg 

import(
  "fmt"
  "io/ioutil"
  "encoding/json"
  
  gov "github.com/cosmos/cosmos-sdk/x/gov/types"
  mt "gitlab.com/microtick/mtzone/x/microtick/types"
)

const (
  ProposalTypeDenomChange string = "MicrotickDenomChange"
  ProposalTypeAddMarkets string = "MicrotickAddMarkets"
)

var _ gov.Content = &DenomChangeProposal{}
var _ gov.Content = &AddMarketsProposal{}

func init() {
  gov.RegisterProposalType(ProposalTypeDenomChange)
  gov.RegisterProposalTypeCodec(&DenomChangeProposal{}, "microtick/DenomChangeProposal")
  gov.RegisterProposalType(ProposalTypeAddMarkets)
  gov.RegisterProposalTypeCodec(&AddMarketsProposal{}, "microtick/AddMarketsProposal")
}


// Denom Change

func NewDenomChangeProposal(title, description, extDenom string, extPerInt int64) gov.Content {
  return &DenomChangeProposal{
    Title: title, 
    Description: description, 
    ExtDenom: extDenom,
    ExtPerInt: extPerInt,
  }
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

// New Markets

func NewAddMarketsProposal(fileName string) (gov.Content, error) {
	contents, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	
	var proposal AddMarketsProposal
	err = json.Unmarshal(contents, &proposal)
	if err != nil {
		return nil, err
	}
	  
	return &proposal, nil
}

func (amp *AddMarketsProposal) ProposalRoute() string  { return mt.ModuleName }
func (amp *AddMarketsProposal) ProposalType() string   { return ProposalTypeAddMarkets }
func (amp *AddMarketsProposal) ValidateBasic() error {
  return gov.ValidateAbstract(amp)
}

func (nm *MarketMetadata) String() string {
  return fmt.Sprintf(`{Name: %s Description: %s}`, nm.Name, nm.Description)
}

func (amp *AddMarketsProposal) String() string {
  var res string = ""
  for i := range amp.Markets {
    res += amp.Markets[i].Name + " "
  }
  return fmt.Sprintf(`Add Markets Proposal:
  Title: %s
  Description: %s
  Markets: %s
`, amp.Title, amp.Description, res)
}
