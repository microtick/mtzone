package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/distribution module sentinel errors
var (
    ModuleName = "microtick"
    ErrMissingParam            = sdkerrors.Register(ModuleName, 1, "missing parameter")
    ErrInvalidAddress          = sdkerrors.Register(ModuleName, 2, "invalid address")
    ErrInvalidMarket           = sdkerrors.Register(ModuleName, 3, "invalid market")
    ErrInvalidDuration         = sdkerrors.Register(ModuleName, 4, "invalid duration")
    ErrInsufficientFunds       = sdkerrors.Register(ModuleName, 5, "insufficient funds")
    ErrTradeMatch              = sdkerrors.Register(ModuleName, 6, "trade matching failed")
    ErrInvalidQuote            = sdkerrors.Register(ModuleName, 7, "invalid quote id")
    ErrInvalidTrade            = sdkerrors.Register(ModuleName, 8, "invalid trade id")
    ErrInvalidRequest          = sdkerrors.Register(ModuleName, 9, "invalid request")
    ErrQuoteNotStale           = sdkerrors.Register(ModuleName, 10, "quote is not stale")
    ErrQuoteFrozen             = sdkerrors.Register(ModuleName, 11, "quote is frozen") 
    ErrQuoteBacking            = sdkerrors.Register(ModuleName, 12, "quote refund failed")
    ErrQuoteParams             = sdkerrors.Register(ModuleName, 13, "quote params out of range")
    ErrNotOwner                = sdkerrors.Register(ModuleName, 14, "not owner")
    ErrTradeSettlement         = sdkerrors.Register(ModuleName, 15, "trade settlement")
    
    ErrGeneral                 = sdkerrors.Register(ModuleName, 999, "general")
)