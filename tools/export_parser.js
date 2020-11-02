/*
 * Parses a state export and produces a table like this:
 *
 * Account balances
 * micro1rxzy06d4cg5rj3nghgyx7x9g862wwn4563mfr6                          : 9684.574654dai     0.478686tick
 * micro1y2j5tjymxgwkq5srppgmm75a62qhkkvd52sxdg                          : 0dai               0tick
 * micro1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38r5gqr (bonded_tokens_pool)     : 0dai               1000000tick
 * micro1tygms3xhhs3yv487phx3dw4a95jn7t7lnrgekh (not_bonded_tokens_pool) : 0dai               0tick
 * micro10gn0ns7wppvv84yy3c9zumxy2z0mnp2k6qenmv (microtick)              : 237.189603dai      0tick
 * micro13j682thdnzhsljge4rraz5gh66syxsdvv0k0rt                          : 176.631834dai      0.2625tick
 * micro1jv65s3grqf6v6jl3dp4t6c9t9rk99cd848emst (distribution)           : 1.604001dai        39.626067tick
 * micro1m0ej6gc3fqnrpun9wwmpec80yye6um4nkay63z                          : 989899.999908dai   0tick
 * micro1m3h30wlvsf8llruxtpukdvsy0km2kum86f6l75 (mint)                   : 0dai               0tick
 * micro17xpfvakm2amg962yls6f84z3kell8c5lzp78jf (fee_collector)          : 0dai               0tick
 * Total: 1000000dai 1000040.367253tick
 *
 * Outstanding rewards:
 * microvaloper1y2j5tjymxgwkq5srppgmm75a62qhkkvdhsksx0: 1.57192098dai 38.83354566tick
 * Total: 1.57192098dai 38.83354566tick
 *
 * Jailed tokens:
 * Total: 0
 *
 * Community pool: 0.03208002dai
 *
 * INVOCATION:
 * node export_parser.js <filename of state export>
 */

const fs = require('fs')
const bignum = require('bignumber.js')

const data = JSON.parse(fs.readFileSync(process.argv[2]).toString())

var totalDai = new bignum(0)
var totalTick = new bignum(0)

console.log("Account balances")
const moduleAccounts = {}
data.app_state.auth.accounts.map(auth => {
  if (auth["@type"] === "/cosmos.auth.v1beta1.ModuleAccount") {
    moduleAccounts[auth.base_account.address] = auth.name
  }
})
data.app_state.bank.balances.map(acct => {
  const amts = acct.coins.reduce((acc, coin) => {
    if (acc[coin.denom] === undefined) {
      acc[coin.denom] = new bignum(0)
    }
    const amt = new bignum(coin.amount).div(1000000)
    acc[coin.denom] = acc[coin.denom].plus(amt)
    return acc
  }, {
    udai: new bignum(0),
    utick: new bignum(0)
  })
  var name = ""
  if (moduleAccounts[acct.address] !== undefined) {
    name = " (" + moduleAccounts[acct.address] + ")"
  }
  const daiAmt = amts["udai"] + "dai"
  const tickAmt = amts["utick"] + "tick"
  console.log(acct.address + name.padEnd(26) + ": " + daiAmt.padEnd(18) + " " + tickAmt)
  totalDai = totalDai.plus(amts["udai"])
  totalTick = totalTick.plus(amts["utick"])
})
console.log("Total: " + totalDai + "dai" + " " + totalTick + "tick")

console.log()
console.log("Outstanding rewards:")
var rewards = {
  dai: new bignum(0),
  tick: new bignum(0)
}
data.app_state.distribution.outstanding_rewards.map(val => {
  var amount = new bignum(0)
  if (val.outstanding_rewards !== null) {
    var damount = new bignum(0)
    var tamount = new bignum(0)
    amount = val.outstanding_rewards.map(coin => {
      if (coin.denom === "udai") {
        damount = new bignum(coin.amount).div(1000000)
        rewards.dai = rewards.dai.plus(damount)
      }
      if (coin.denom === "utick") {
        tamount = new bignum(coin.amount).div(1000000)
        rewards.tick = rewards.tick.plus(tamount)
      }
    })
    console.log(val.validator_address + ": " + damount + "dai" + " " + tamount + "tick")
  }
})

console.log("Total: " + rewards.dai + "dai" + " " + rewards.tick + "tick")

console.log()
console.log("Jailed tokens:")
var jailed = new bignum(0)
data.app_state.staking.validators.map(val => {
  if (val.jailed) {
    console.log(val.operator_address + ": " + val.tokens + "utick")
    jailed = jailed.plus(new bignum(val.tokens))
  }
})
console.log("Total: " + jailed)

console.log()
var community = new bignum(0)
if (data.app_state.distribution.fee_pool.community_pool !== null) {
  community = data.app_state.distribution.fee_pool.community_pool.reduce((acc, coin) => {
    if (coin.denom === "udai") {
      const amt = new bignum(coin.amount).div(1000000)
      return acc.plus(amt)
    }
    return acc
  }, new bignum(0))
}
console.log("Community pool: " + community + "dai")
