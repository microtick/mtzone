class Accounts {
    
    constructor(state) {
        this.state = state
    }
    
    createAccount(id, bal) {
        if (this.state[id] === undefined) {
            this.state[id] = {
                id: id,
                balance: bal,
                numQuotes: 0,
                numTrades: 0,
                quoteBacking: 0,
                tradeBacking: 0,
                activeQuotes: [],
                activeTrades: {
                    long: [],
                    short: []
                }
            }
            return true
        } else {
            return false
        }
    }
    
    exists(id) {
        return this.state[id] !== undefined
    }
    
    balance(id) {
        if (this.state[id] === undefined) {
            throw new Error("No such account")
        }
        return this.state[id].balance
    }
    
    withdraw(acct, amount) {
        if (this.state[acct] === undefined) {
            throw new Error("No such account")
        }
        if (amount > this.state[acct].balance) {
            throw new Error("Insufficient funds")
        }
        this.state[acct].balance = Math.round10(this.state[acct].balance - amount, -6)
    }
    
    deposit(acct, amount) {
        if (this.state[acct] === undefined) {
            throw new Error("No such account")
        }
        this.state[acct].balance = Math.round10(this.state[acct].balance + amount, -6)
    }
    
    incrementQuoteCount(acct) {
        if (this.state[acct] === undefined) {
            throw new Error("No such account")
        }
        this.state[acct].numQuotes++
    }
    
    addQuote(acct, id, side) {
        if (this.state[acct] === undefined) {
            throw new Error("No such account")
        }
        var ary = this.state[acct].activeQuotes
        ary.push(id)
    }
    
    addQuoteBacking(acct, backing) {
        this.state[acct].quoteBacking = Math.round10(this.state[acct].quoteBacking + backing, -6)
    }
    
    removeQuote(acct, id, side) {
        if (this.state[acct] === undefined) {
            throw new Error("No such account")
        }
        var ary = this.state[acct].activeQuotes
        ary.splice(ary.indexOf(id), 1)
    }
    
    removeQuoteBacking(acct, backing) {
        this.state[acct].quoteBacking = Math.round10(this.state[acct].quoteBacking - backing, -6)
    }
    
    incrementTradeCount(acct) {
        if (this.state[acct] === undefined) {
            throw new Error("No such account")
        }
        this.state[acct].numTrades++
    }
    
    addTrade(acct, id, side) {
        if (this.state[acct] === undefined) {
            throw new Error("No such account")
        }
        if (side === 0) {
          var ary = this.state[acct].activeTrades['long']
        } else {
          ary = this.state[acct].activeTrades['short']
        }
        ary.push(id)
    }
    
    addTradeBacking(acct, backing) {
        this.state[acct].tradeBacking = Math.round10(this.state[acct].tradeBacking + backing, -6)
    }
    
    removeTradeBacking(acct, backing) {
        this.state[acct].tradeBacking = Math.round10(this.state[acct].tradeBacking - backing, -6)
    }
    
    removeTrade(acct, id, side) {
        if (this.state[acct] === undefined) {
            throw new Error("No such account")
        }
        if (side === 0) {
          var ary = this.state[acct].activeTrades['long']
        } else {
          ary = this.state[acct].activeTrades['short']
        }
        ary.splice(ary.indexOf(id), 1)
    }
    
}

module.exports = Accounts
