// Copyright 2019 Andrei Tudor Călin
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

package zerocopy_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"sync"
	"testing"

	"acln.ro/zerocopy"
)

func TestTeeRead(t *testing.T) {
	primary, err := zerocopy.NewPipe()
	if err != nil {
		t.Fatal(err)
	}

	const n = 3

	secondaries := make([]*zerocopy.Pipe, n)
	for i := 0; i < n; i++ {
		secondary, err := zerocopy.NewPipe()
		if err != nil {
			t.Fatal(err)
		}
		secondaries[i] = secondary
	}

	ws := make([]io.Writer, n)
	for i := 0; i < n; i++ {
		ws[i] = secondaries[i]
	}
	primary.Tee(ws...)

	msg := "hello world"

	errs := make([]error, 0, n)

	var wg sync.WaitGroup
	wg.Add(n + 1)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			buf := make([]byte, len(msg))
			_, err := io.ReadFull(secondaries[i], buf)
			if err != nil {
				errs[i] = err
				return
			}
			if string(buf) != msg {
				errs[i] = fmt.Errorf("got %q, want %q", buf, msg)
			}
		}(i)
	}

	var primaryerr error

	go func() {
		_, primaryerr = io.Copy(ioutil.Discard, onlyReader{primary})
		wg.Done()
	}()

	if _, err := io.WriteString(primary, msg); err != nil {
		t.Fatal(err)
	}
	primary.CloseWrite()
	println("waiting")
	wg.Wait()
	if primaryerr != nil {
		t.Error(primaryerr)
	}
	for _, err := range errs {
		if err != nil {
			t.Error(err)
		}
	}
}

type onlyReader struct {
	io.Reader
}
