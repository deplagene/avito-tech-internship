// DBTX is an interface for a database transaction.
type DBTX interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}