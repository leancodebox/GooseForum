export interface GooseClientErrorOptions {
  status?: number
  code?: number
  messageCode?: string
  params?: Record<string, unknown>
  cause?: unknown
}

export class GooseClientError extends Error {
  readonly status?: number
  readonly code?: number
  readonly messageCode?: string
  readonly params?: Record<string, unknown>

  constructor(message: string, options: GooseClientErrorOptions = {}) {
    super(message, { cause: options.cause })
    this.name = 'GooseClientError'
    this.status = options.status
    this.code = options.code
    this.messageCode = options.messageCode
    this.params = options.params
  }
}

export class GooseProtocolError extends GooseClientError {
  constructor(message: string, options: GooseClientErrorOptions = {}) {
    super(message, options)
    this.name = 'GooseProtocolError'
  }
}
