package retry

// Limit represents the maximum number of times an action can be retried before failing
type Limit int

const RetryOnce = Limit(0)
const RetryForever = Limit(-1)
