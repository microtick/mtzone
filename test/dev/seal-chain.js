(async () => {
  
  const cp = require('child_process')
  const fs = require('fs')
  
  const HOME=process.argv[2]
  const CONFIG=process.argv[3]
  const config = JSON.parse(fs.readFileSync(CONFIG))
  
  const CHAINHOME=HOME + "/" + config.chain_id
  
  console.log()
  console.log("Sealing chain: " + config.chain_id)
  
  const chainexec = cmd => {
    console.log("  $ " + config.executable + " --home " + CHAINHOME + " " + cmd)
    const bufs = cp.spawnSync(config.executable, [
      "--home " + CHAINHOME,
      cmd
    ],{
      shell: true
    })
    const ret = {
      stdout: bufs.stdout.toString(),
      stderr: bufs.stderr.toString()
    }
    return ret
  }
  
  await chainexec("gentx validator 100000000000stake --keyring-backend test --chain-id " + config.chain_id)
  await chainexec("collect-gentxs")
  
})().catch(e => {
  
  console.error(e)
  
})