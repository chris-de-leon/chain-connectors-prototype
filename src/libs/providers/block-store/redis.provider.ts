import { BlockManager } from "../../providers/block-manager/provider.ts"
import { common } from "../../common/mod.ts"
import { BlockStore } from "./provider.ts"
import { Redis, Result } from "ioredis"

const SAVE_BLOCKS_SCRIPT = `
  local hash_key = KEYS[1]
  local stream_key = KEYS[2]

  redis.call("HSET", hash_key, "cursor", ARGV[1])
  for i = 2, #ARGV do
    redis.call("XADD", stream_key, "*", "data", ARGV[i])
  end
`

// https://github.com/redis/ioredis/blob/ec42c82ceab1957db00c5175dfe37348f1856a93/examples/typescript/scripts.ts#L12
declare module "ioredis" {
  interface RedisCommander<Context> {
    saveBlocks(
      hashKey: string,
      streamKey: string,
      ...argv: string[]
    ): Result<string, Context>
  }
}

export class RedisBlockStore<T> implements BlockStore<T> {
  private readonly streamKey: string
  private readonly hashKey: string

  constructor(
    private readonly config: {
      providers: {
        blockManager: BlockManager<T>
        redis: Redis
      }
      opts: {
        chainId: string
      }
    },
  ) {
    this.streamKey = common.makeBlockStreamKey(this.config.opts.chainId)
    this.hashKey = common.makeBlockCursorKey(this.config.opts.chainId)
    this.config.providers.redis.defineCommand("saveBlocks", {
      lua: SAVE_BLOCKS_SCRIPT,
      numberOfKeys: 2,
    })
  }

  async getCursor() {
    return await this.config.providers.redis
      .hget(this.hashKey, "cursor")
      .then((c) => (typeof c === "string" ? parseInt(c, 10) : c))
  }

  async save(blocks: T[]) {
    const num = this.config.providers.blockManager.getMaxBlockNumber(blocks)
    if (num == null) {
      throw new Error("unexpectedly received undefined checkpoint")
    }

    const data = blocks.map((b) => JSON.stringify(b))
    await this.config.providers.redis.saveBlocks(
      this.hashKey,
      this.streamKey,
      num.toString(),
      ...data,
    )

    return num
  }
}
