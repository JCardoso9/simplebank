package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)

	acc1 := createRandomAccount(t)
	acc2 := createRandomAccount(t)
	fmt.Println(">> before:", acc1.Balance, acc2.Balance)

	// Writing to DB is usually problematic with concurrency so we test that
	n := 5
	amount := int64(10)

	// Cant check for error inside go routines because it does not guarantee the test will stop
	// since it is running in another go routine
	errs := make(chan error)
	results := make(chan TransferTxResult)

	for i := 0; i < n; i++ {
		go func() {
			result, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: acc1.ID,
				ToAccountID:   acc2.ID,
				Amount:        amount,
			})

			errs <- err
			results <- result
		}()
	}

	// check results
	existed := make(map[int]bool)

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
		result := <-results

		// check transfer
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, acc1.ID, transfer.FromAccountID)
		require.Equal(t, acc2.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		// Notice that since store has *Queries it has access to the CRUD methods as well
		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		// check entries
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, acc1.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, acc2.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		// check accounts
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, acc1.ID, fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, acc2.ID, toAccount.ID)

		// check balances
		fmt.Println(">> tx:", fromAccount.Balance, toAccount.Balance)

		diff1 := acc1.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - acc2.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0) // 1 * amount, 2 * amount, 3 * amount, ..., n * amount

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	updatedAccount1, err := store.GetAccount(context.Background(), acc1.ID)
	require.NoError(t, err)

	updatedAccount2, err := store.GetAccount(context.Background(), acc2.ID)
	require.NoError(t, err)

	fmt.Println(">> after:", updatedAccount1.Balance, updatedAccount2.Balance)

	require.Equal(t, acc1.Balance-int64(n)*amount, updatedAccount1.Balance)
	require.Equal(t, acc2.Balance+int64(n)*amount, updatedAccount2.Balance)

}

// Deadlock can still occur when you have two concurrent transactions, one A->B and one B->A
func TestTransferTxDeadlock(t *testing.T) {
	store := NewStore(testDB)

	acc1 := createRandomAccount(t)
	acc2 := createRandomAccount(t)
	fmt.Println(">> before:", acc1.Balance, acc2.Balance)

	// Writing to DB is usually problematic with concurrency so we test that
	n := 10 // 5 to send to A and 5 send to B
	amount := int64(10)

	// Cant check for error inside go routines because it does not guarantee the test will stop
	// since it is running in another go routine
	errs := make(chan error)
	// results := make(chan TransferTxResult)  Result has already been checked

	for i := 0; i < n; i++ {
		fromAccountID := acc1.ID
		toAccountID := acc2.ID

		if i%2 == 0 {
			fromAccountID = acc2.ID
			toAccountID = acc1.ID
		}

		go func() {
			_, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: fromAccountID,
				ToAccountID:   toAccountID,
				Amount:        amount,
			})

			errs <- err
		}()
	}

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

	}

	updatedAccount1, err := store.GetAccount(context.Background(), acc1.ID)
	require.NoError(t, err)

	updatedAccount2, err := store.GetAccount(context.Background(), acc2.ID)
	require.NoError(t, err)

	fmt.Println(">> after:", updatedAccount1.Balance, updatedAccount2.Balance)

	require.Equal(t, acc1.Balance, updatedAccount1.Balance)
	require.Equal(t, acc2.Balance, updatedAccount2.Balance)

}
