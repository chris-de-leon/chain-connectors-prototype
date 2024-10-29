export type BlockStream = {
  waitForNextBlock(): Promise<void>
  unsubscribe(): Promise<void>
  subscribe(): Promise<void>
}
