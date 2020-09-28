package types

func NewGenesisState(
    params MicrotickParams, 
    accounts []GenesisAccount, 
    markets []GenesisMarket, 
    durations []GenesisDuration) GenesisMicrotick {
        
    return GenesisMicrotick {
        Params: params,
        Accounts: accounts,
        Markets: markets,
        Durations: durations,
    }
}

func DefaultGenesisState() GenesisMicrotick {
    return NewGenesisState(DefaultParams(), nil, nil, nil)
}

func ValidateGenesis(gs GenesisMicrotick) error {
  if err := gs.Params.ValidateBasic(); err != nil {
    return err
  }
  return nil
}
