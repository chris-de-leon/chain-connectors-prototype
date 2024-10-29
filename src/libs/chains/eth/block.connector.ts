import { BlockConnector } from "../../providers/block-connector/provider.ts"
import { BlockStore } from "../../providers/block-store/provider.ts"
import { EthereumBlockStream } from "./block.stream.ts"
import { common } from "../../common/mod.ts"
import { JsonRpcProvider } from "ethers"
import { abortable } from "@std/async"

export class EthereumBlockConnector implements BlockConnector {
  constructor(
    private readonly config: {
      providers: {
        blockStream: EthereumBlockStream
        blockStore: BlockStore<unknown>
        rpc: JsonRpcProvider
      }
    },
  ) {}

  public async run(controller: AbortController) {
    await this.config.providers.blockStream.subscribe()

    const onabort = controller.signal.onabort
    controller.signal.onabort = async (ev) => {
      await common.logIfErrorAsync(() =>
        this.config.providers.blockStream.unsubscribe()
      )
      if (onabort != null) {
        await onabort.call(controller.signal, ev)
      }
    }

    return await abortable(
      (async () => {
        const cursor = await this.initCursor()
        for await (const b of this.waitForNextBlock(cursor)) {
          await this.config.providers.blockStore
            .save([b])
            .then((c) => console.log("Processed block:", c))
        }
      })(),
      controller.signal,
    )
  }

  private async initCursor() {
    const { providers } = this.config

    const cursor = await providers.blockStore.getCursor()
    if (cursor != null) {
      return cursor
    }

    const block = await providers.rpc.getBlock("latest")
    if (block == null) {
      throw new Error("failed to retrieve latest block")
    } else {
      return await providers.blockStore.save([block])
    }
  }

  private async *waitForNextBlock(startNum: number) {
    for (let cursor = startNum;; cursor++) {
      const nextNum = cursor + 1
      if (nextNum > (await this.config.providers.rpc.getBlockNumber())) {
        await this.config.providers.blockStream.waitForNextBlock()
      }

      const block = await this.config.providers.rpc.getBlock(nextNum)
      if (block == null) {
        throw new Error(`Failed to fetch the next block (number = ${nextNum})`)
      } else {
        yield block
      }
    }
  }
}
