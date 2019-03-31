class OrderedList {
    
    constructor(state, compareFunc) {
        this.state = state
        this.compare = compareFunc
    }
    
    search(data) {
        var lo = 0
        var hi = this.state.length
        while (hi - lo > 1) {
            const mid = Math.round((hi + lo) / 2)
            if (this.compare(data, this.state[mid]) >= 0) {
                lo = mid
            } else {
                hi = mid
            }
        }
        if (lo < this.state.length && this.compare(data, this.state[lo]) >= 0)
            return hi
        return lo
    }
    
    insert(data) {
        const pos = this.search(data)
        this.state.splice(pos, 0, data)
    }
    
    delete(data) {
        this.state = this.state.reduce((x, d) => {
            if (d !== data) {
                x.push(d)
                return x
            }
        }, [])
    }
    
}

module.exports = OrderedList
