package microtick

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
)

type ListItem struct {
    Id MicrotickId `json:"Id"`
    Value sdk.Dec `json:"Value"`
}

func NewListItem(id MicrotickId, value sdk.Dec) ListItem {
    return ListItem {
        Id: id,
        Value: value,
    }
}

type OrderedList struct {
    Data []ListItem
}

func NewOrderedList() OrderedList {
    return OrderedList {
        Data: make([]ListItem, 0, 0),
    }
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

func (ol *OrderedList) Delete(id MicrotickId) {
    len := len(ol.Data)
    if len > 0 {
        cur := ol.Data
        ol.Data = make([]ListItem, 0)
        for i := 0; i < len; i++ {
            if cur[i].Id != id {
                ol.Data = append(ol.Data, cur[i])
            }
        }
    }
}
