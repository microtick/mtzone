package types

type GenesisAccount struct {
    Account MicrotickAccount `json:"account"`
    NumQuotes uint32 `json:"numQuotes"`
    NumTrades uint32 `json:"numTrades"`
}

type GenesisMarket struct {
    Name MicrotickMarket `json:"name"`
    Description string `json:"description"`
}

type GenesisDuration struct {
    Name MicrotickDurationName `json:"name"`
    Seconds MicrotickDuration `json:"seconds"`
}

type GenesisState struct {
    Params Params `json:"params"`
    Accounts []GenesisAccount `json:"accounts"`
    Markets []GenesisMarket `json:"markets"`
    Durations []GenesisDuration `json:"durations"`
}

func NewGenesisState(params Params, accounts []GenesisAccount, markets []GenesisMarket, durations []GenesisDuration) GenesisState {
    return GenesisState {
        Params: params,
        Accounts: accounts,
        Markets: markets,
        Durations: durations,
    }
}

func DefaultGenesisState() GenesisState {
    return NewGenesisState(DefaultParams(), nil, nil, nil)
}

func ValidateGenesis(gs GenesisState) error {
  if err := gs.Params.ValidateBasic(); err != nil {
    return err
  }
  return nil
}
