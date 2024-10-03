package batcherror

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBatchError(t *testing.T) {
	testBatch := []bool{false, true, true, true, true, false, false, false} // should_fail = false / true
	processBatch := func(fail bool) error {
		if fail {
			return errors.New("failure")
		}
		return nil
	}

	var joinedErrs error
	for idx, batchItem := range testBatch {
		if processErr := processBatch(batchItem); processErr != nil {
			joinedErrs = errors.Join(joinedErrs, New(processErr, idx))
			joinedErrs = errors.Join(joinedErrs, errors.New("some other failure"))
		}
	}

	failedBatchItemIdx := map[int]struct{}{}
	for _, err := range UnwrapJoinedErrors(joinedErrs) {
		batchErr := new(BatchError)
		if errors.As(err, &batchErr) {
			failedBatchItemIdx[batchErr.Idx()] = struct{}{}
		}
	}

	for idx := range testBatch {
		_, failed := failedBatchItemIdx[idx]
		require.Equal(t, testBatch[idx], failed, "item [%v] in testBatch should be %v", idx, failed)
	}

	m := MapIndexedErrors(joinedErrs)
	require.Len(t, m, 4)
	for idx, err := range m {
		require.True(t, testBatch[idx])
		require.Equal(t, err.Error(), AtIdx(joinedErrs, idx).Error())
	}

	msg := Short(joinedErrs, 3)
	require.Equal(t,
		"failure at [1]\nsome other failure\nfailure at [2] \n5 other errors",
		msg.Error(),
	)

	msg = Short(joinedErrs, 10)
	require.Equal(t,
		"failure at [1]\nsome other failure\nfailure at [2]\nsome other failure\nfailure at [3]\nsome other failure\nfailure at [4]\nsome other failure",
		msg.Error(),
	)

	msg = Short(errors.New("foo"), 1)
	require.Equal(t, "foo", msg.Error())

	msg = Short(nil, 1)
	require.Nil(t, msg)

	msg = Short(errors.Join(nil, New(ErrBatchFailed, 1)), 2)
	require.Equal(t, "batch failed: batch failed at [1]", msg.Error())
}
