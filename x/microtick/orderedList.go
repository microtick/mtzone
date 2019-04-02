package microtick

type ListItem struct {
    Id uint `json:"Id"`
    Value int `json:"Value"`
}

func NewListItem(id uint, value int) ListItem {
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
        if li.Value >= ol.Data[mid].Value {
            lo = mid
        } else {
            hi = mid
        }
    }
    if lo < len(ol.Data) && li.Value != ol.Data[lo].Value {
        return hi
    }
    return lo
}

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

func (ol *OrderedList) Delete(li ListItem) {
    len := len(ol.Data)
    if len > 0 {
        cur := ol.Data
        ol.Data = make([]ListItem, len-1, len-1)
        pos := 0
        for i := 0; i < len; i++ {
            if cur[i].Id != li.Id {
                ol.Data[pos] = cur[i]
                pos++
            }
        }
    }
}
