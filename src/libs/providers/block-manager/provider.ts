export type BlockManager<T> = {
  getMaxBlockNumber(blocks: T[]): number | undefined
}
