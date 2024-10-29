import { BlockStream } from "../../providers/block-stream/provider.ts"
import { WebSocketProvider } from "ethers"
import { EventEmitter } from "node:events"
import { abortable } from "@std/async"
import { z } from "zod"

export class EthereumBlockStream implements BlockStream {
  private subscriptionId: string | undefined
  private isSubscribed = false

  private readonly evEmitter = new EventEmitter()
  private readonly unsub = new AbortController()
  private readonly sub = new AbortController()

  constructor(
    private readonly config: {
      providers: {
        wss: WebSocketProvider
      }
    },
  ) {
    this.config.providers.wss.websocket.onmessage = (event) => {
      const data = JSON.parse(event.data)

      const zUnsubscribeData = z.object({
        jsonrpc: z.string(),
        result: z.boolean(),
        id: z.number(),
      })

      const zSubscribeData = z.object({
        jsonrpc: z.string(),
        result: z.string(),
        id: z.literal(1),
      })

      const zChainIdData = z.object({
        jsonrpc: z.string(),
        result: z.string(),
        id: z.literal(2),
      })

      const zBlockData = z.object({
        jsonrpc: z.string(),
        method: z.string(),
        params: z.object({
          subscription: z.string(),
          result: z.object({ number: z.string() }),
        }),
      })

      if (this.subscriptionId == null) {
        const result = zSubscribeData.safeParse(data)
        if (result.success) {
          this.subscriptionId = result.data.result
          this.sub.abort()
          return
        } else {
          throw new Error(
            `Failed to set subscription ID: ${JSON.stringify({ data })}`,
          )
        }
      }

      const unsubscribeResult = zUnsubscribeData.safeParse(data)
      if (unsubscribeResult.success) {
        this.unsub.abort()
        return
      }

      const chainIdResult = zChainIdData.safeParse(data)
      if (chainIdResult.success) {
        console.log("Chain ID:", parseInt(chainIdResult.data.result, 16))
        return
      }

      const blockResult = zBlockData.safeParse(data)
      if (!blockResult.success) {
        console.warn(`Ignoring event: ${JSON.stringify({ data })}`)
      } else {
        this.evEmitter.emit("block")
      }
    }

    this.config.providers.wss.websocket.onerror = (err) => {
      this.evEmitter.emit("error", err)
    }
  }

  async subscribe() {
    if (this.isSubscribed) {
      return
    } else {
      this.isSubscribed = true
    }

    try {
      await abortable(
        // NOTE: this call will block if it is awaited, so we need to make it abortable
        this.config.providers.wss.send("eth_subscribe", ["newHeads"]),
        this.sub.signal,
      )
    } catch (err) {
      if (!this.sub.signal.aborted) {
        throw err
      } else {
        console.log("Subscription ID:", this.subscriptionId)
      }
    }
  }

  async unsubscribe() {
    const subId = this.subscriptionId
    if (subId == null) {
      return
    }

    try {
      await abortable(
        // NOTE: this call will block if it is awaited, so we need to make it abortable
        this.config.providers.wss.send("eth_unsubscribe", [subId]),
        this.unsub.signal,
      )
    } catch (err) {
      if (!this.unsub.signal.aborted) {
        throw err
      } else {
        console.log("Successfully closed subscription")
      }
    }
  }

  async waitForNextBlock() {
    return await new Promise<void>((res, rej) => {
      const blockListener = () => {
        this.evEmitter.off("block", blockListener)
        this.evEmitter.off("error", errorListener)
        res(undefined)
      }

      const errorListener = (err: unknown) => {
        this.evEmitter.off("block", blockListener)
        this.evEmitter.off("error", errorListener)
        rej(err)
      }

      this.evEmitter.on("block", blockListener)
      this.evEmitter.on("error", errorListener)
    })
  }
}
