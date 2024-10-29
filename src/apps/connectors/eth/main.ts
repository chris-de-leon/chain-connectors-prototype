import { RedisBlockStore } from "../../../libs/providers/block-store/redis.provider.ts"
import { EthereumBlockConnector } from "../../../libs/chains/eth/block.connector.ts"
import { EthereumBlockManager } from "../../../libs/chains/eth/block.manager.ts"
import { EthereumBlockStream } from "../../../libs/chains/eth/block.stream.ts"
import { common } from "../../../libs/common/mod.ts"
import { ethers } from "ethers"
import { Redis } from "ioredis"
import { z } from "zod"

const envvars = z
  .object({
    CHAIN_RPC_URL: z.string().url(),
    CHAIN_WSS_URL: z.string().url(),
    REDIS_URL: z.string().url(),
    CHAIN_ID: z.string().min(1),
  })
  .parse(Deno.env.toObject())

const wssProvider = new ethers.WebSocketProvider(envvars.CHAIN_WSS_URL)
const rpcProvider = new ethers.JsonRpcProvider(envvars.CHAIN_RPC_URL)
const redisClient = new Redis(envvars.REDIS_URL)
const controller = new AbortController()

const blockManager = new EthereumBlockManager()

const blockStream = new EthereumBlockStream({
  providers: {
    wss: wssProvider,
  },
})

const blockStore = new RedisBlockStore({
  providers: {
    redis: redisClient,
    blockManager,
  },
  opts: {
    chainId: envvars.CHAIN_ID,
  },
})

const service = new EthereumBlockConnector({
  providers: {
    rpc: rpcProvider,
    blockStream,
    blockStore,
  },
})

controller.signal.onabort = async () => {
  await common.logIfErrorAsync(async () => {
    await redisClient.quit()
    console.log("Successfully closed redis client")
  })

  common.logIfError(() => {
    wssProvider.websocket.close()
    console.log("Successfully closed websocket")
  })
}

Deno.addSignalListener("SIGTERM", () => controller.abort("SIGTERM"))
Deno.addSignalListener("SIGHUP", () => controller.abort("SIGHUP"))
Deno.addSignalListener("SIGINT", () => controller.abort("SIGINT"))

console.log("Process ID:", Deno.pid)
try {
  await service.run(controller)
} catch (err) {
  if (controller.signal.aborted) {
    console.log(`Stopping service: ${err}`)
  } else {
    throw err
  }
}
