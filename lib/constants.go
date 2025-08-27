package lib

type txKey string

const (
	ContextKeyGormTx txKey = "ezutil.gormTx"

	MsgTransactionError = "error processing transaction"
)
