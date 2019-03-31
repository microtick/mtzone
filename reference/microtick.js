const Accounts = require('./accounts')
const Markets = require('./markets')
const Validator = require('./validator')
const OrderedList = require('./orderedlist')
require('./util')

const LEVERAGE = 10

class Microtick {
    
    constructor(state) {
        if (state.nextQuoteId === undefined) {
            //console.log("Initializing state")
            // uninitialized
            state.nextQuoteId = 1
            state.nextTradeId = 1
            state.accounts = {}
            state.markets = {}
            state.activeQuotes = {}
            state.activeTrades = {}
            state.tradeExpirations = []
            state.ca = 0
            state.config = {
                settleTime: 30,
                commissionQuote: 0.0005, // percentage of backing
                commissionUpdate: 0.00005, // percentage of backing
                commissionTrade: 0.05 // fixed fee
            }
        }
        this.state = state
        this.Accounts = new Accounts(this.state.accounts)
        this.Markets = new Markets(this.state.markets)
        this.valid = new Validator()
    }
    
    genesis(accounts) {
        accounts.map(acct => {
            console.log("Genesis: funding account " + acct.address + ": " + acct.balance)
            this.Accounts.createAccount(acct.address, acct.balance)
        })
    }
    
    createAccount(accountId, bal) {
        this.valid.matchAddress(accountId)
        
        this.Accounts.createAccount(accountId, bal)
    }
    
    getBalance(accountId) {
        this.valid.matchAddress(accountId)
        
        return this.Accounts.balance(accountId)
    }
    
    depositAccount(accountId, amount) {
        this.valid.matchAddress(accountId)
        this.valid.matchTokenAmount(amount)
        
        this.Accounts.deposit(accountId, amount)
        return {
            data: {
                amount: amount
            },
            tags: [{
                key: "acct." + accountId,
                value: "deposit"
            }]
        }
    }
    
    //withdrawAccount(accountId, amount) {
        //this.Accounts.withdraw(accountId, amount)
        //return {
            //data: {
                //amount: amount
            //},
            //tags: [{
                //key: "acct." + accountId,
                //value: "withdraw"
            //}]
        //}
    //}
    
    transfer(from, to, amount) {
        this.valid.matchAddress(from)
        this.valid.matchAddress(to)
        this.valid.matchTokenAmount(amount)
        
        this.Accounts.withdraw(from, amount)
        this.Accounts.deposit(to, amount)
    }
    
    createMarket(market) {
        this.valid.matchMarket(market)
        
        this.Markets.createMarket(market)
    }
    
    consensusSpot(market) {
        this.valid.matchMarket(market)
        
        return this.Markets.consensusSpot(market)
    }
    
    createQuote(timestamp, accountId, market, dur, spot, premium, backing) {
        this.valid.matchTimestamp(timestamp)
        this.valid.matchAddress(accountId)
        this.valid.matchMarket(market)
        this.valid.matchDuration(dur)
        this.valid.matchPrice(spot)
        this.valid.matchPrice(premium)
        this.valid.matchTokenAmount(backing)
        
        spot = Math.round10(spot, -6)
        premium = Math.round10(premium, -6)
        backing = Math.round10(backing, -6)
        this.Accounts.withdraw(accountId, backing)
        const quoteId = this.state.nextQuoteId++
        this.state.activeQuotes[quoteId] = {
            id: quoteId,
            provider: accountId,
            market: market,
            dur: dur,
            spot: spot,
            premium: premium,
            backing: backing,
            quantity: Math.round10(backing / (premium * LEVERAGE), -6),
            modified: timestamp,
            canModify: timestamp + this.state.config.settleTime
        }
        this.Markets.addQuote(quoteId, id => {
            return this.state.activeQuotes[id]
        })
        this.Accounts.addQuote(accountId, quoteId)
        this.Accounts.addQuoteBacking(accountId, backing)
        
        // commission
        const comm = Math.round10(backing * this.state.config.commissionQuote, -6)
        this.Accounts.withdraw(accountId, comm)
        this.state.ca = Math.round10(this.state.ca + comm, -6)
        this.state.activeQuotes[quoteId].commission = comm
        
        this.Accounts.incrementQuoteCount(accountId)
        return quoteId
    }
    
    getQuote(quoteid) {
        this.valid.matchId(quoteid)
        
        return this.state.activeQuotes[quoteid]
    }
    
    getTrade(tradeid) {
        this.valid.matchId(tradeid)
        
        return this.state.activeTrades[tradeid]
    }
    
    cancelQuote(timestamp, accountId, quoteId) {
        this.valid.matchTimestamp(timestamp)
        this.valid.matchAddress(accountId)
        this.valid.matchId(quoteId)
        
        if (!this.Accounts.exists(accountId)) {
            throw new Error("No such account")
        }
        const quote = this.state.activeQuotes[quoteId]
        if (quote === undefined) {
            throw new Error("No such quote")
        }
        if (quote.provider !== accountId) {
            throw new Error("Account cannot cancel quote")
        }
        if (quote.canModify > timestamp) {
            throw new Error("Quote can't be modified")
        }
        const backing = quote.backing
        this.Accounts.deposit(quote.provider, backing)
        this.Accounts.removeQuote(accountId, quoteId)
        this.Accounts.removeQuoteBacking(accountId, quote.backing)
        this.Markets.removeQuote(quoteId, id => {
            return this.state.activeQuotes[id]
        })
        delete this.state.activeQuotes[quoteId]
        return backing
    }
    
    depositQuote(timestamp, accountId, quoteId, backing) {
        this.valid.matchTimestamp(timestamp)
        this.valid.matchAddress(accountId)
        this.valid.matchId(quoteId)
        this.valid.matchTokenAmount(backing)
        
        backing = Math.round10(backing, -6)
        this.Accounts.withdraw(accountId, backing)
        if (!this.Accounts.exists(accountId)) {
            throw new Error("No such account")
        }
        const quote = this.state.activeQuotes[quoteId]
        if (quote === undefined) {
            throw new Error("No such quote")
        }
        this.Markets.removeQuote(quoteId, id => {
            return this.state.activeQuotes[id]
        })
        quote.backing = Math.round10(quote.backing + backing, -6)
        quote.quantity = Math.round10(quote.backing / (quote.premium * LEVERAGE), -6)
        quote.modified = timestamp
        quote.canModify = timestamp + this.state.config.settleTime
        this.Markets.addQuote(quoteId, id => {
            return this.state.activeQuotes[id]
        })
        this.Accounts.addQuoteBacking(accountId, backing)
        
        // commission
        const comm = Math.round10(backing * this.state.config.commissionQuote, -6)
        this.Accounts.withdraw(accountId, comm)
        this.state.ca = Math.round10(this.state.ca + comm, -6)
        
        return comm
    }
    
    updateQuote(timestamp, accountId, quoteId, spot, premium) {
        this.valid.matchTimestamp(timestamp)
        this.valid.matchAddress(accountId)
        this.valid.matchId(quoteId)
        this.valid.matchPrice(spot)
        this.valid.matchPrice(premium)
        
        spot = Math.round10(spot, -6)
        premium = Math.round10(premium, -6)
        if (!this.Accounts.exists(accountId)) {
            throw new Error("No such account")
        }
        const quote = this.state.activeQuotes[quoteId]
        if (quote === undefined) {
            throw new Error("No such quote")
        }
        if (quote.provider !== accountId) {
            throw new Error("Account cannot update quote")
        }
        if (quote.canModify > timestamp) {
            throw new Error("Quote can't be modified")
        }
        this.Markets.removeQuote(quoteId, id => {
            return this.state.activeQuotes[id]
        })
        if (spot > 0) {
            quote.spot = spot
            quote.modified = timestamp
            quote.canModify = timestamp + this.state.config.settleTime
        }
        if (premium > 0) {
            quote.premium = premium
            quote.quantity = Math.round10(quote.backing / (quote.premium * LEVERAGE), -6)
            quote.modified = timestamp
            quote.canModify = timestamp + this.state.config.settleTime
        }
        this.Markets.addQuote(quoteId, id => {
            return this.state.activeQuotes[id]
        })
        
        // commission
        const comm = Math.round10(quote.backing * this.state.config.commissionUpdate, -6)
        this.Accounts.withdraw(accountId, comm)
        this.state.ca = Math.round10(this.state.ca + comm, -6)
        return comm
    }
    
    callPrice(qid, strike) {
        this.valid.matchId(qid)
        this.valid.matchPrice(strike)
        
        const quote = this.state.activeQuotes[qid]
        var call = quote.premium - (strike - quote.spot) / 2
        if (call < 0) call = 0
        return call
    }
    
    putPrice(qid, strike) {
        this.valid.matchId(qid)
        this.valid.matchPrice(strike)
        
        const quote = this.state.activeQuotes[qid]
        var put = quote.premium - (quote.spot - strike) / 2
        if (put < 0) put = 0
        return put
    }
    
    createTrade(timestamp, accountId, market, dur, type, qty) {
        this.valid.matchTimestamp(timestamp)
        this.valid.matchAddress(accountId)
        this.valid.matchMarket(market)
        this.valid.matchDuration(dur)
        this.valid.matchType(type)
        this.valid.matchQuantity(qty)
        
        if (!this.Accounts.exists(accountId)) {
            throw new Error("No such account")
        }
        const tradeId = this.state.nextTradeId++
        const trade = {
            id: tradeId,
            long: accountId,
            type: type,
            market: market,
            dur: dur,
            start: timestamp,
            expiration: timestamp + dur,
            quantity: qty,
            strike: this.Markets.consensusSpot(market),
            counterparties: []
        }
        this.state.activeTrades[tradeId] = trade
        this.Accounts.addTrade(accountId, tradeId, 0)
        //console.log("Trade " + tradeId + " qty=" + qty + " strike=" + trade.strike + " expiration=" + trade.expiration)
        const Expirations = new OrderedList(this.state.tradeExpirations, (t1, t2) => {
            if (t1.expiration > t2.expiration) return 1
            if (t1.expiration < t2.expiration) return -1
            return 0
        })
        Expirations.insert({
            id: tradeId,
            expiration: trade.expiration
        })
        var total_premium = 0
        this.Markets.match(trade, id => {
            const quote = this.state.activeQuotes[id]
            if (quote.provider === accountId) {
                throw new Error("Attempted trade against same account")
            }
            return quote
        }, cpdata => {
            if (trade.type === 0) {
                // Call
                var premium = Math.round10(this.callPrice(cpdata.quoteId, trade.strike) * cpdata.quantity, -6)
            } else {
                // Put
                premium = Math.round10(this.putPrice(cpdata.quoteId, trade.strike) * cpdata.quantity, -6)
            }
            cpdata.premium = premium
            total_premium = Math.round10(total_premium + premium, -6)
            this.Accounts.withdraw(trade.long, premium)
            this.Accounts.deposit(cpdata.short, premium)
            this.Accounts.removeQuoteBacking(cpdata.short, cpdata.backing)
            this.Accounts.addTrade(cpdata.short, tradeId, 1)
            this.Accounts.addTradeBacking(cpdata.short, cpdata.backing)
            cpdata.startBalance = this.Accounts.balance(cpdata.short)
            trade.counterparties.push(cpdata)
        }, (acct, id, backing) => {
            delete this.state.activeQuotes[id]
            this.Accounts.removeQuote(acct, id)
        })
        
        // commission
        this.Accounts.withdraw(accountId, this.state.config.commissionTrade)
        this.state.ca = Math.round10(this.state.ca + this.state.config.commissionTrade, -6)
        
        this.Accounts.incrementTradeCount(accountId)
        
        trade.premium = total_premium
        trade.commission = this.state.config.commissionTrade
        trade.startBalance = this.Accounts.balance(trade.long)
        return trade
    }
    
    limitTrade(timestamp, accountId, market, dur, type, limit, maxcost) {
        this.valid.matchTimestamp(timestamp)
        this.valid.matchAddress(accountId)
        this.valid.matchMarket(market)
        this.valid.matchDuration(dur)
        this.valid.matchType(type)
        this.valid.matchPrice(limit)
        this.valid.matchTokenAmount(maxcost)
        
        console.log("limit trade: " + limit + " maxcost=" + maxcost)
        
        if (!this.Accounts.exists(accountId)) {
            throw new Error("No such account")
        }
        const tradeId = this.state.nextTradeId++
        const trade = {
            id: tradeId,
            long: accountId,
            type: type,
            market: market,
            dur: dur,
            start: timestamp,
            expiration: timestamp + dur,
            quantity: 0,
            strike: this.Markets.consensusSpot(market),
            counterparties: []
        }
        this.state.activeTrades[tradeId] = trade
        this.Accounts.addTrade(accountId, tradeId, 0)
        //console.log("Trade " + tradeId + " qty=" + qty + " strike=" + trade.strike + " expiration=" + trade.expiration)
        const Expirations = new OrderedList(this.state.tradeExpirations, (t1, t2) => {
            if (t1.expiration > t2.expiration) return 1
            if (t1.expiration < t2.expiration) return -1
            return 0
        })
        Expirations.insert({
            id: tradeId,
            expiration: trade.expiration
        })
        var total_cost = maxcost
        var total_premium = 0
        this.Markets.limit(trade, id => {
            return this.state.activeQuotes[id]
        }, id => {
            const quote = this.state.activeQuotes[id]
            if (quote.provider === accountId) {
                // no error, just not matched
                return 0
            }
            if (trade.type === 0) {
                // Call
                var price = Math.round10(this.callPrice(id, trade.strike), -6)
            } else {
                // Put
                price = Math.round10(this.putPrice(id, trade.strike), -6)
            }
            if (price <= limit) {
                if (Math.round10(price * quote.quantity, -6) <= total_cost) {
                    total_cost = Math.round10(total_cost - price * quote.quantity, -6)
                    return quote.quantity
                } else {
                    return Math.round10(total_cost / price, -6)
                }
            } else {
                return 0
            }
        }, cpdata => {
            if (trade.type === 0) {
                // Call
                var premium = Math.round10(this.callPrice(cpdata.quoteId, trade.strike) * cpdata.quantity, -6)
            } else {
                // Put
                premium = Math.round10(this.putPrice(cpdata.quoteId, trade.strike) * cpdata.quantity, -6)
            }
            cpdata.premium = premium
            total_premium = Math.round10(total_premium + premium, -6)
            this.Accounts.withdraw(trade.long, premium)
            this.Accounts.deposit(cpdata.short, premium)
            this.Accounts.removeQuoteBacking(cpdata.short, cpdata.backing)
            this.Accounts.addTrade(cpdata.short, tradeId, 1)
            this.Accounts.addTradeBacking(cpdata.short, cpdata.backing)
            cpdata.startBalance = this.Accounts.balance(cpdata.short)
            trade.counterparties.push(cpdata)
        }, (acct, id, backing) => {
            delete this.state.activeQuotes[id]
            this.Accounts.removeQuote(acct, id)
        })
        if (trade.quantity === 0) {
            throw new Error("No items matched limit")
        }
        
        // commission
        this.Accounts.withdraw(accountId, this.state.config.commissionTrade)
        this.state.ca = Math.round10(this.state.ca + this.state.config.commissionTrade, -6)
        
        this.Accounts.incrementTradeCount(accountId)
        
        trade.premium = total_premium
        trade.commission = this.state.config.commissionTrade
        trade.startBalance = this.Accounts.balance(trade.long)
        return trade
    }
    
    settleTrades(timestamp) {
        this.valid.matchTimestamp(timestamp)
        
        const ret = {
            settled: [],
            final: {}
        }
        while (this.state.tradeExpirations.length > 0 &&
                this.state.tradeExpirations[0].expiration <= timestamp) {
            const tid = this.state.tradeExpirations[0].id
            //console.log("settling: " + tid)
            const trade = this.state.activeTrades[tid]
            ret.settled.push(trade)
            const settle = this.Markets.consensusSpot(trade.market)
            ret.final[trade.market] = settle
            //console.log("Trade " + tid + " strike=" + trade.strike + " settle=" + settle)
            var totalSettle = 0
            trade.counterparties.map(cp => {
                if (trade.type === 0) { // Call
                    var profit = Math.round10((settle - trade.strike) * cp.quantity, -6)
                } else { // Put
                    profit = Math.round10((trade.strike - settle) * cp.quantity, -6)
                }
                if (profit < 0) profit = 0
                if (profit > cp.backing) profit = cp.backing
                const remain = Math.round10(cp.backing - profit, -6)
                totalSettle = Math.round10(totalSettle + profit, -6)
                if (profit > 0) {
                    this.Accounts.deposit(trade.long, profit)
                }
                if (remain > 0) {
                    this.Accounts.deposit(cp.short, remain)
                }
                this.Accounts.removeTradeBacking(cp.short, cp.backing)
                this.Accounts.removeTrade(cp.short, tid, 1)
                cp.settle = remain
                cp.endBalance = this.Accounts.balance(cp.short)
            })
            trade.settle = totalSettle
            trade.endBalance = this.Accounts.balance(trade.long)
            this.state.tradeExpirations.splice(0, 1)
            this.Accounts.removeTrade(trade.long, tid, 0)
            delete this.state.activeTrades[tid]
        }
        return ret
    }
}

module.exports = Microtick
