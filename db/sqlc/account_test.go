package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/JCardoso9/simplebank/util"
	"github.com/stretchr/testify/require"
)

// Make tests independent of each other, this function doesnt start
// with caps so won0t be run
func createRandomAccount(t *testing.T) Account {
	arg := CreateAccountParams{
		Owner:    util.RandomOwner(),
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
	account, err := testQueries.CreateAccount(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, account)

	require.Equal(t, arg.Owner, account.Owner)
	require.Equal(t, arg.Balance, account.Balance)
	require.Equal(t, arg.Currency, account.Currency)

	require.NotZero(t, account.ID)
	require.NotZero(t, account.CreatedAt)
	return account
}

func TestCreateAccount(t *testing.T) {
	createRandomAccount(t)
}

func TestGetAccount(t *testing.T) {
	acc1 := createRandomAccount(t)
	acc2, err := testQueries.GetAccount(context.Background(), acc1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, acc2)
	require.Equal(t, acc1.Owner, acc2.Owner)
	require.Equal(t, acc1.Balance, acc2.Balance)
	require.Equal(t, acc1.Currency, acc2.Currency)
	require.WithinDuration(t, acc1.CreatedAt.Time, acc2.CreatedAt.Time, time.Second)
}

func TestUpdateAccount(t *testing.T) {
	acc1 := createRandomAccount(t)
	args := UpdateAccountParams{
		ID:      acc1.ID,
		Balance: util.RandomMoney(),
	}

	acc2, err := testQueries.UpdateAccount(context.Background(), args)
	require.NoError(t, err)
	require.NotEmpty(t, acc1)
	require.Equal(t, args.Balance, acc2.Balance)
	require.Equal(t, acc1.Currency, acc2.Currency)
	require.Equal(t, acc1.Owner, acc2.Owner)
}

func TestDeleteAccount(t *testing.T) {
	acc1 := createRandomAccount(t)

	err := testQueries.DeleteAccount(context.Background(), acc1.ID)
	require.NoError(t, err)

	acc2, err := testQueries.GetAccount(context.Background(), acc1.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, acc2)
}

func TestListAccounts(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomAccount(t)
	}

	// Skip the first 5 and return the next 5
	args := ListAccountsParams{
		Limit:  5,
		Offset: 5,
	}

	accs, err := testQueries.ListAccounts(context.Background(), args)
	require.NoError(t, err)

	require.Len(t, accs, 5)

	for _, account := range accs {
		require.NotEmpty(t, account)
	}
}
