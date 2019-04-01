package microtick

import (
    "fmt"
)

type ListItem interface {
    ListItemId() uint
}

type CompareFunc func(x, y ListItem) int

type OrderedList struct {
    Compare CompareFunc `json:"-"`
    Data []ListItem
}

func NewOrderedList(compareFunc CompareFunc) OrderedList {
    return OrderedList {
        Compare: compareFunc,
        Data: make([]ListItem, 0, 0),
    }
}

func (ol *OrderedList) Search(li ListItem) int {
    var lo, hi int
    lo = 0
    hi = len(ol.Data)
    for hi - lo > 1 {
        mid := (hi + lo) / 2
        if ol.Compare(li, ol.Data[mid]) >= 0 {
            lo = mid
        } else {
            hi = mid
        }
    }
    if lo < len(ol.Data) && ol.Compare(li, ol.Data[lo]) != 0 {
        return hi
    }
    return lo
}

func (ol *OrderedList) Insert(li ListItem) {
    pos := ol.Search(li)
    curlen := len(ol.Data)
    fmt.Printf("Before size=%d\n", curlen)
    cur := ol.Data
    fmt.Printf("Resizing list: %d\n", curlen+1)
    ol.Data = make([]ListItem, curlen+1, curlen+1)
    for i := 0; i < pos; i++ {
        fmt.Printf("Copying element %d to %d\n", i, i)
        ol.Data[i] = cur[i]
    }
    fmt.Printf("Assigning %d\n", pos)
    ol.Data[pos] = li
    for i := pos; i < curlen; i++ {
        fmt.Printf("Copying element %d to %d\n", i, i+1)
        ol.Data[i+1] = cur[i]
    }
    fmt.Printf("After size=%d\n", len(ol.Data))
    fmt.Printf("%v\n", ol.Data)
}

func (ol *OrderedList) Delete(li ListItem) {
    len := len(ol.Data)
    if len > 0 {
        cur := ol.Data
        ol.Data = make([]ListItem, len-1, len-1)
        pos := 0
        for i := 0; i < len; i++ {
            if cur[i].ListItemId() != li.ListItemId() {
                ol.Data[pos] = cur[i]
                pos++
            }
        }
    }
}
