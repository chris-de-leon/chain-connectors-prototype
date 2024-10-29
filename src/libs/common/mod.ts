export const common = {
  makeBlockStreamKey: (id: string) => `${id}:block-stream`,
  makeBlockCursorKey: (id: string) => `${id}:block-cursor`,
  logIfErrorAsync: async (cb: () => Promise<unknown>) => {
    await cb().catch(console.error)
  },
  logIfError: (cb: () => unknown) => {
    try {
      cb()
    } catch (err) {
      console.error(err)
    }
  },
}
