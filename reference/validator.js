class Validator {
    
    constructor() {
        this.hex = /(0x)?[0-9a-f]+/i
        this.market = /[0-9a-z\.]+/i
    }
    
    matchTimestamp(ts) {
        if (typeof ts !== 'number') {
            throw new Error("Invalid timestamp (1): " + ts)
        }
        /*
        const now = new Date().getTime() / 1000
        if (ts < now - 300 || ts > now + 300) {
            throw new Error("Invalid timestamp (2): " + ts)
        }
        */
    }
    
    matchAddress(addr) {
        if (typeof addr !== 'string') {
            throw new Error("Invalid address (1): " + addr)
        }
        if (!addr.match(this.hex)) {
            throw new Error("Invalid address (2): " + addr)
        }
        if (addr.slice(0,2) === '0x') addr = addr.slice(2)
        if (addr.length !== 40) {
            throw new Error("Invalid address (3): " + addr)
        }
    }
    
    matchTokenAmount(num) {
        if (typeof num !== 'number') {
            throw new Error("Invalid token amount (1): " + num)
        }
        if (num < 0) {
            throw new Error("Invalid token amount (2): " + num)
        }
    }
    
    matchMarket(mkt) {
        if (typeof mkt !== 'string') {
            throw new Error("Invalid market: (1)" + mkt)
        }
        if (!mkt.match(this.market)) {
            throw new Error("Invalid market: (2)" + mkt)
        }
    }
    
    matchDuration(dur) {
        if (typeof dur !== 'number') {
            throw new Error("Invalid duration (1): " + dur)
        }
        if (dur !== 300 && 
            dur !== 900 &&
            dur !== 3600 &&
            dur !== 14400 &&
            dur !== 43200) {
            throw new Error("Invalid duration (2): " + dur)
        }
    }
    
    matchPrice(pr) {
        if (typeof pr !== 'number') {
            throw new Error("Invalid price (1): " + pr)
        }
        if (pr < 0) {
            throw new Error("Invalid price (2): " + pr)
        }
    }
    
    matchId(id) {
        if (typeof id !== 'number') {
            throw new Error("Invalid id (1): " + id)
        }
        if (!Number.isInteger(id)) {
            throw new Error("Invalid id (2): " + id)
        }
        if (id < 0) {
            throw new Error("Invalid id (3): " + id)
        }
    }
    
    matchType(type) {
        if (typeof type !== 'number') {
            throw new Error("Invalid trade type (1): " + type)
        }
        if (type !== 0 && type !== 1) {
            throw new Error("Invalid trade type (2): " + type)
        }
    }
    
    matchQuantity(qty) {
        if (typeof qty !== 'number') {
            throw new Error("Invalid quantity (1): " + qty)
        }
        if (qty < 0) {
            throw new Error("Invalid quantity (2): " + qty)
        }   
    }
    
}

module.exports = Validator