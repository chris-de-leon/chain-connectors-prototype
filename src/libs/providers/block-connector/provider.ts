export type BlockConnector = {
  run(controller: AbortController): Promise<void>
}
