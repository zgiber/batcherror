package batcherror

import (
	"errors"
	"fmt"
	"strings"
)

var ErrBatchFailed = errors.New("batch failed")

type multiErr interface {
	Unwrap() []error
}

// BatchError is a special error type that can be used to allow individual
// failures in batch operations without failing the full batch.
// BatchError references the index in the batch where the error occurred,
// enabling the caller to decide whether to continue or not.
// The common pattern for usage:
//
// collecting the errors while processing the batch:
//
// var batchErr error
//
//	for idx, batchItem := range batchItems {
//		if err := process(batchItem); err != nil {
//			batchErr = errors.Join(batchErr, batcherror.New(err, idx))
//			continue
//		}
//	}
//
// Example for error handling:
//
//	for idx, batchItem := range batchItems {
//			if err := batcherror.AtIdx(idx); err != nil {
//				// handle error...
//			}
//		}
type BatchError struct {
	idx int
	err error
}

func New(err error, idx int) *BatchError {
	return &BatchError{
		idx: idx,
		err: err,
	}
}

func (b *BatchError) Error() string {
	return fmt.Sprintf("%s at [%v]", b.err.Error(), b.idx)
}

func (b *BatchError) Unwrap() error {
	return b.err
}

func (b *BatchError) Idx() int {
	return b.idx
}

// AtIdx returns an error if the provided err is a joinError type and
// any of the joined errors is a BatchError that is at the specified idx.
func AtIdx(err error, idx int) error {
	var match error
	collect := func(e error) {
		be := new(BatchError)
		if errors.As(e, &be) {
			if be.Idx() == idx {
				match = errors.Join(match, be)
			}
		}
	}
	traverseJoinedErrors(err, collect)
	return match
}

// MapIndexedErrors returns a map where the batch errors are mapped
// to the index of the batch item they occurred at.
func MapIndexedErrors(err error) map[int]error {
	m := map[int]error{}
	collect := func(e error) {
		be := new(BatchError)
		if errors.As(e, &be) {
			m[be.Idx()] = be
		}
	}
	traverseJoinedErrors(err, collect)
	return m
}

// UnwrapJoinedErrors returns the slice of errors that is the result of using errors.Join
// If err does not implement MultiErr then it is returned as the single item in the slice.
func UnwrapJoinedErrors(err error) []error {
	errs := []error{}
	collect := func(e error) {
		errs = append(errs, e)
	}

	traverseJoinedErrors(err, collect)
	return errs
}

// traverseJoinedErrors traverses the tree of wrapped errors (DFS) and collect them
// using the provided function.
func traverseJoinedErrors(err error, collect func(error)) {
	e, ok := err.(multiErr)
	if !ok {
		collect(err)
		return
	}

	errs := e.Unwrap()
	if len(errs) == 1 {
		collect(err)
		return
	}

	for _, e := range errs {
		traverseJoinedErrors(e, collect)
	}
}

// Short generic version of the error for logging purposes.
// Prints the details of the first N errors and the total number of errors.
func Short(err error, maxMessages int) error {
	if err == nil {
		return nil
	}
	errs := UnwrapJoinedErrors(err)
	var msg string
	if len(errs) > maxMessages {
		msg = errors.Join(errs[:maxMessages]...).Error()
		msg = strings.Join([]string{msg, fmt.Sprintf("\n%v other errors", len(errs)-maxMessages)}, " ")
	} else {
		msg = errors.Join(errs...).Error()
	}

	if errors.Is(err, ErrBatchFailed) {
		msg = strings.Join([]string{"batch failed:", msg}, " ")
	}

	return errors.New(msg)
}
