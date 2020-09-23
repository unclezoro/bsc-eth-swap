package swap

import (
	"testing"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/stretchr/testify/require"
)

func TestSwap_InsertAndConfirmSwapRequest(t *testing.T) {
	config := GetTestConfig()
	_, err := PrepareDB(config)
	require.NoError(t, err)
}
