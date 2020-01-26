package oauth2bearer

// TokenSourceParams holds the parameters needed to build a token source
type TokenSourceParams struct {
	RefreshMargin float64
	NumRetries    int
	RetrySleep    float64
}

// refreshMarginSeconds controls how many seconds before the expected expiry
// of a bearer token we decide to refresh it
const refreshMarginSeconds = 10

// how many times to try refreshing a token before panicing
const refreshRetries = 10

// how long to sleep, in milliseconds, between retries
const retrySleepMS = 100

// NewDefaultTokenSourceParams returns a params struct populated with the
// defaults
func NewDefaultTokenSourceParams() TokenSourceParams {
	return TokenSourceParams{RefreshMargin: refreshMarginSeconds, NumRetries: refreshRetries, RetrySleep: retrySleepMS}
}

// eof
