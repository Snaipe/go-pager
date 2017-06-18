/* go-pager
 *
 * Copyright (c) 2017 Franklin "Snaipe" Mathieu <me@snai.pe>
 * Use of this source-code is govered by the MIT license, which
 * can be found in the LICENSE file.
 */

package pager_test

import (
	"log"
	"os"

	"snai.pe/go-pager"
)

func ExampleOpen() {
	out, err := pager.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	out.Write([]byte("Hello, World!"))
	// Output: Hello, World!
}

func ExampleOpen_Stderr() {
	out, err := pager.OpenPager("", os.Stderr)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	out.Write([]byte("Hello, World!"))
}

func ExampleOpen_Environment() {
	out, err := pager.OpenPager(os.Getenv("GO_PAGER"), os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	out.Write([]byte("Hello, World!"))
	// Output: Hello, World!
}
