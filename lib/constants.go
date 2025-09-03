package lib

type txKey string

const (
	ContextKeyGormTx txKey = "go-crud.gormTx"

	MsgTransactionError = "error processing transaction"
)
