const OrderedList = require('./orderedlist')

const durs = [300, 900, 3600, 14400, 43200]
const LEVERAGE = 10

class Markets {
    
    constructor(state) {
        this.state = state
    }
    
    createMarket(market) {
        if (this.state[market] === undefined) {
            this.state[market] = {
                sumSpots: 0,
                sumWeight: 0,
                sumBacking: 0,
                consensus: 0,
                orderbooks: {}
            }
            for (var i=0; i<durs.length; i++) {
                this.state[market].orderbooks[durs[i]] = {
                    calls: [],
                    puts: [],
                    sumWeight: 0
                }
            }
        }
    }
    
    consensusSpot(market) {
        if (this.state[market] === undefined) {
            throw new Error("No such market: " + market)
        }
        return this.state[market].consensus
    }
    
    factorOut(quote) {
        const market = this.state[quote.market]
        market.sumSpots = Math.round10(market.sumSpots - quote.spot * quote.quantity, -6)
        market.sumWeight = Math.round10(market.sumWeight - quote.quantity, -6)
        market.sumBacking = Math.round10(market.sumBacking - quote.backing, -6)
        market.orderbooks[quote.dur].sumWeight = Math.round10(market.orderbooks[quote.dur].sumWeight - quote.quantity, -6)
        if (market.sumWeight > 0) market.consensus =  Math.round10(market.sumSpots / market.sumWeight, -6)
    }
    
    factorIn(quote) {
        const market = this.state[quote.market]
        const previousConsensus = market.consensus
        market.sumSpots = Math.round10(market.sumSpots + quote.spot * quote.quantity, -6)
        market.sumWeight = Math.round10(market.sumWeight + quote.quantity, -6)
        market.sumBacking = Math.round10(market.sumBacking + quote.backing, -6)
        market.orderbooks[quote.dur].sumWeight = Math.round10(market.orderbooks[quote.dur].sumWeight + quote.quantity, -6)
        market.consensus = Math.round10(market.sumSpots / market.sumWeight, -6)
        if (previousConsensus > 0) {
            const delta = Math.abs(Math.round10(market.consensus - previousConsensus, -6))
            if (quote.premium < delta / 2) {
                // Artificially adjust quote premium up to match delta / 2
                quote.premium = delta / 2
            }
        }
    }
    
    addQuote(id, fetch) {
        const quote = fetch(id)
        if (!durs.includes(quote.dur)) {
            throw new Error("Invalid duration: " + quote.dur)
        }
        if (this.state[quote.market] === undefined) {
            throw new Error("No such market: " + quote.market)
        }
        const CallList = new OrderedList(this.state[quote.market].orderbooks[quote.dur].calls, (id1, id2) => {
            const q1 = fetch(id1)
            const q2 = fetch(id2)
            const v1 = q1.premium + q1.spot / 2
            const v2 = q2.premium + q2.spot / 2
            if (v1 > v2) return 1
            if (v1 < v2) return -1
            return 0
        })
        CallList.insert(id)
        const PutList = new OrderedList(this.state[quote.market].orderbooks[quote.dur].puts, (id1, id2) => {
            const q1 = fetch(id1)
            const q2 = fetch(id2)
            const v1 = q1.premium - q1.spot / 2
            const v2 = q2.premium - q2.spot / 2
            if (v1 > v2) return 1
            if (v1 < v2) return -1
            return 0
        })
        PutList.insert(id)
        this.factorIn(quote)
    }
    
    removeQuote(id, fetch) {
        const quote = fetch(id)
        this.factorOut(quote)
        const calls = this.state[quote.market].orderbooks[quote.dur].calls
        calls.splice(calls.indexOf(id), 1)
        const puts = this.state[quote.market].orderbooks[quote.dur].puts
        puts.splice(puts.indexOf(id), 1)
    }
    
    match(trade, fetchQuote, addCounterparty, deleteQuote) {
        if (this.state[trade.market] === undefined) {
            throw new Error("No such market: " + trade.market)
        }
        if (!durs.includes(trade.dur)) {
            throw new Error("Invalid duration: " + trade.dur)
        }
        const market = this.state[trade.market]
        const ob = market.orderbooks[trade.dur]
        var quantityToMatch = trade.quantity
        while (quantityToMatch > 0) {
            if (trade.type === 0) { // Call
                var list = ob.calls
            } else { // Put
                list = ob.puts
            }
            if (list.length > 0) {
                const qid = list[0]
                const quote = fetchQuote(qid)
                const params = {
                    spot: quote.spot,
                    premium: quote.premium,
                    quantity: quote.quantity
                }
                this.factorOut(quote)
                if (quote.quantity >= quantityToMatch) { 
                    // Complete fill
                    const backing = Math.round10(quote.backing * (quantityToMatch / quote.quantity), -6)
                    quote.backing = Math.round10(quote.backing - backing, -6)
                    quote.quantity = Math.round10(quote.backing / (quote.premium * LEVERAGE), -6)
                    addCounterparty({
                        quoteId: list[0],
                        qparams: params,
                        short: quote.provider,
                        quantity: quantityToMatch,
                        backing: backing,
                        final: quote.quantity === 0 ? true : false
                    })
                    quantityToMatch = 0
                } else { 
                    // Partial fill
                    quantityToMatch = Math.round10(quantityToMatch - quote.quantity, -6)
                    addCounterparty({
                        quoteId: list[0],
                        qparams: params,
                        short: quote.provider,
                        quantity: quote.quantity,
                        backing: quote.backing,
                        final: true
                    })
                    quote.quantity = 0
                }
                if (quote.quantity === 0) {
                    // Delete
                    ob.calls.splice(ob.calls.indexOf(qid), 1)
                    ob.puts.splice(ob.puts.indexOf(qid), 1)
                    deleteQuote(quote.provider, qid, quote.backing)
                } else {
                    this.factorIn(quote)
                }
            } else {
                throw new Error("Not enough quantity for trade")
            }
        }
    }
    
    limit(trade, fetchQuote, computeQuantity, addCounterparty, deleteQuote) {
        if (this.state[trade.market] === undefined) {
            throw new Error("No such market: " + trade.market)
        }
        if (!durs.includes(trade.dur)) {
            throw new Error("Invalid duration: " + trade.dur)
        }
        const market = this.state[trade.market]
        const ob = market.orderbooks[trade.dur]
        if (trade.type === 0) { // Call
            var list = ob.calls
        } else { // Put
            list = ob.puts
        }
        const toDelete = []
        for (var i=0; i<list.length; i++) {
            const qid = list[i]
            const qty = computeQuantity(qid)
            if (qty > 0) {
                trade.quantity = Math.round10(trade.quantity + qty, -6)
                const quote = fetchQuote(qid)
                const params = {
                    spot: quote.spot,
                    premium: quote.premium,
                    quantity: quote.quantity
                }
                this.factorOut(quote)
                
                // Complete fill
                quote.quantity = Math.round10(quote.quantity - qty, -6)
                const backing = Math.round10(quote.backing - quote.quantity * (quote.premium * LEVERAGE), -6)
                quote.backing = Math.round10(quote.backing - backing, -6)
                addCounterparty({
                    quoteId: qid,
                    qparams: params,
                    short: quote.provider,
                    quantity: qty,
                    backing: backing,
                    final: quote.quantity > 0 ? false : true
                })
                
                if (quote.quantity === 0) {
                    toDelete.push(qid)
                    deleteQuote(quote.provider, qid, backing)
                } else {
                    this.factorIn(quote)
                }
            }
        }
        for (i=0; i<toDelete.length; i++) {
            const qid = toDelete[i]
            // Delete
            ob.calls.splice(ob.calls.indexOf(qid), 1)
            ob.puts.splice(ob.puts.indexOf(qid), 1)
        }
    }
    
}

module.exports = Markets
