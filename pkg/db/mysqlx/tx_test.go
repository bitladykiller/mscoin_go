package mysqlx

import (
	"context"
	"strings"
	"testing"
)

func TestTxManagerWithinTxRequiresInitializedDB(t *testing.T) {
	t.Parallel()

	manager := &txManager{}
	err := manager.WithinTx(context.Background(), func(exec ExtContext) error {
		return nil
	})
	if err == nil || !strings.Contains(err.Error(), "not initialized") {
		t.Fatalf("WithinTx() error = %v, want initialization error", err)
	}
}
