export type BlockStore<T> = {
  getCursor(): Promise<number | null>
  save(blocks: T[]): Promise<number>
}
