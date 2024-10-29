import { BlockManager } from "../../providers/block-manager/provider.ts"
import { maxBy } from "@std/collections/max-by"
import { Block } from "ethers"

export class EthereumBlockManager implements BlockManager<Block> {
  getMaxBlockNumber(blocks: Block[]) {
    return maxBy(blocks, (b) => b.number)?.number
  }
}
