package dataloadgen_test

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/vikstrous/dataloadgen"
)

func ExampleLoader() {
	ctx := context.Background()
	loader := dataloadgen.NewLoader(func(keys []string) (ret []int, errs []error) {
		for _, key := range keys {
			num, err := strconv.ParseInt(key, 10, 32)
			ret = append(ret, int(num))
			errs = append(errs, err)
		}
		return
	},
		dataloadgen.WithBatchCapacity(1),
		dataloadgen.WithWait(16*time.Millisecond),
	)
	one, err := loader.Load(ctx, "1")
	if err != nil {
		panic(err)
	}
	fmt.Println(one)
	// Output: 1
}

func TestEdgeCases(t *testing.T) {
	ctx := context.Background()
	var fetches [][]int
	var mu sync.Mutex
	dl := dataloadgen.NewLoader(func(keys []int) ([]string, []error) {
		mu.Lock()
		fetches = append(fetches, keys)
		mu.Unlock()

		results := make([]string, len(keys))
		errors := make([]error, len(keys))

		for i, key := range keys {
			if key%2 == 0 {
				errors[i] = fmt.Errorf("not found")
			} else {
				results[i] = fmt.Sprint(key)
			}
		}
		return results, errors
	},
		dataloadgen.WithBatchCapacity(5),
		dataloadgen.WithWait(1*time.Millisecond),
	)

	t.Run("load function called only once when cached", func(t *testing.T) {
		for i := 0; i < 2; i++ {
			_, err := dl.Load(ctx, 0)
			if err == nil {
				t.Fatal("expected error")
			}
			if len(fetches) != 1 {
				t.Fatal("wrong number of fetches", fetches)
			}
			if len(fetches[0]) != 1 {
				t.Fatal("wrong number of keys in fetch request")
			}
		}
		for i := 0; i < 2; i++ {
			r, err := dl.Load(ctx, 1)
			if err != nil {
				t.Fatal(err)
			}
			if len(fetches) != 2 {
				t.Fatal("wrong number of fetches", fetches)
			}
			if len(fetches[1]) != 1 {
				t.Fatal("wrong number of keys in fetch request")
			}
			if r != "1" {
				t.Fatal("wrong data fetched", r)
			}
		}
	})
}
