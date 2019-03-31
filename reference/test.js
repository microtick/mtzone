const Microtick = require('./microtick')

const state = {}
const mtm = new Microtick(state)

mtm.createAccount("0x12345")
//mtm.createAccount("0x12345")
mtm.createAccount("0x67890")

mtm.depositAccount('0x12345', 10000)
mtm.depositAccount('0x67890', 10000)

mtm.createMarket("ETHUSD")

const q1 = mtm.createQuote(0, "0x12345", "ETHUSD", 300, 100.5, 1, 20)
const q2 = mtm.createQuote(0, "0x12345", "ETHUSD", 300, 101, 0.5, 10)
mtm.createQuote(0, "0x12345", "ETHUSD", 300, 101, 1.1, 1000)
const q3 = mtm.createQuote(0, "0x67890", "ETHUSD", 900, 99, 3, 10)

//mtm.createTrade(1, '0x67890', "ETHUSD", 300, 0, 5)
mtm.limitTrade(1, '0x67890', "ETHUSD", 300, 0, .1, 2)

//mtm.updateBlock(100)
//mtm.settleTrades(400)

console.log(JSON.stringify(state, null, 2))
console.log("consensus=" + state.markets.ETHUSD.consensus)
