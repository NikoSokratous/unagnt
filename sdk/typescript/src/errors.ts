/** APIError is thrown on HTTP errors. */
export class APIError extends Error {
  constructor(
    public readonly statusCode: number,
    message: string,
    public readonly body?: string
  ) {
    super(message);
    this.name = 'APIError';
  }
}

/** NotFoundError is thrown when a resource does not exist (404). */
export class NotFoundError extends APIError {
  constructor(message?: string) {
    super(404, message ?? 'Not found');
    this.name = 'NotFoundError';
  }
}

/** UnauthorizedError is thrown when auth fails (401). */
export class UnauthorizedError extends APIError {
  constructor() {
    super(401, 'Unauthorized');
    this.name = 'UnauthorizedError';
  }
}
