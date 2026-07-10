// The client tests need to be inside the lib package so this test file
// was created for examples that use the lib package from an external package
// like a normal user would.
package lib_test

import (
	"fmt"

	"quiet_hn/lib"
)

func ExampleClient() {
	var client lib.Client
	ids, err := client.TopItems()
	if err != nil {
		panic(err)
	}
	for i := 0; i < 5; i++ {
		item, err := client.GetItem(ids[i])
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s (by %s)\n", item.Title, item.By)
	}
}
