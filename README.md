# batcherror

Keep track of individual errors within a batched process.

## Usage

```go
package main

import (
	"errors"
	"fmt"

	"github.com/zgiber/batcherror"
)

func main() {
	items := []any{1, 2, 3, "foo", 4, "bar", 5, 6}
	if err := processItems(items); err != nil {
		fmt.Println(err)
		// not an integer at [3]
		// not an integer at [5]

		fmt.Println(batcherror.At(err, 1) == nil)
		// true

		fmt.Println(batcherror.At(err, 3))
		// not an integer at [3]

		for k, v := range batcherror.Map(err) {
			fmt.Println(k, v)
			// 3 not an integer at [3]
			// 5 not an integer at [5]
		}
	}
}

func processItems(items []any) error {
	var batchErr error
	for idx, item := range items {
		if _, ok := item.(int); !ok {
			batchErr = errors.Join(batchErr, batcherror.New(fmt.Errorf("not an integer"), idx))
		}
	}
	return batchErr
}
```
