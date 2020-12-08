package types

func NewGenesisState(
    extTokenType string,
    extPerInt uint32,
    params MicrotickParams, 
    accounts []GenesisAccount, 
    markets []GenesisMarket, 
    durations []GenesisDuration) GenesisMicrotick {
        
    return GenesisMicrotick {
        ExtDenom: extTokenType,
        ExtPerInt: extPerInt,
        Params: params,
        Accounts: accounts,
        Markets: markets,
        Durations: durations,
    }
}

func DefaultGenesisState() GenesisMicrotick {
    return NewGenesisState("udai", 1000000, DefaultParams(), nil, nil, nil)
}

func ValidateGenesis(gs GenesisMicrotick) error {
  if err := gs.Params.ValidateBasic(); err != nil {
    return err
  }
  return nil
}
