package keeper

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
    
    mt "github.com/mjackson001/mtzone/x/microtick/types"
)

func NewListItem(id mt.MicrotickId, value sdk.Dec) ListItem {
    return ListItem {
        Id: id,
        Value: value,
    }
}

func NewOrderedList() OrderedList {
    return OrderedList {
        Data: make([]ListItem, 0, 0),
    }
}

func (ol *OrderedList) First() ListItem {
    return ol.Data[0]
}

func (ol *OrderedList) Last() ListItem {
    listLen := len(ol.Data)
    return ol.Data[listLen-1]
}

func (ol *OrderedList) Search(li ListItem) int {
    var lo, hi int
    lo = 0
    hi = len(ol.Data)
    for hi - lo > 1 {
        mid := (hi + lo) / 2
        if li.Value.GTE(ol.Data[mid].Value) {
            lo = mid
        } else {
            hi = mid
        }
    }
    if lo < len(ol.Data) && li.Value.GTE(ol.Data[lo].Value) {
        return hi
    }
    return lo
}

func (ol OrderedList) Contains(id mt.MicrotickId) bool {
    curlen := len(ol.Data)
    for i := 0; i < curlen; i++ {
        if ol.Data[i].Id == id {
            return true
        }
    }
    return false
}

// TODO: more efficient algorithms for insert / delete

func (ol *OrderedList) Insert(li ListItem) {
    pos := ol.Search(li)
    curlen := len(ol.Data)
    cur := ol.Data
    ol.Data = make([]ListItem, curlen+1, curlen+1)
    for i := 0; i < pos; i++ {
        ol.Data[i] = cur[i]
    }
    ol.Data[pos] = li
    for i := pos; i < curlen; i++ {
        ol.Data[i+1] = cur[i]
    }
}

func (ol *OrderedList) Delete(id mt.MicrotickId) {
    curlen := len(ol.Data)
    if curlen > 0 {
        cur := ol.Data
        ol.Data = make([]ListItem, 0)
        for i := 0; i < curlen; i++ {
            if cur[i].Id != id {
                ol.Data = append(ol.Data, cur[i])
            }
        }
    }
}
