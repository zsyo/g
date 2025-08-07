// Copyright (c) 2019 voidint <voidint@126.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package checksum

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/voidint/g/pkg/errs"
)

func TestVerifyFile(t *testing.T) {
	type args struct {
		algo             Algorithm
		expectedChecksum string
		filename         string
	}
	tests := []struct {
		name string
		args args
		err  error
	}{
		{
			name: "SHA1",
			args: args{
				algo:             SHA1,
				expectedChecksum: "8233f28c479ff758b3b4ba9ad66069db68811e59",
				filename:         "./testdata/hello.txt",
			},
			err: nil,
		},
		{
			name: "SHA256",
			args: args{
				algo:             SHA256,
				expectedChecksum: "a5f4396b45548597f81681147f53c66065d5137f2fbd85e6758a8983107228e4",
				filename:         "./testdata/hello.txt",
			},
			err: nil,
		},
		{
			name: "unsupported checksum algorithm",
			args: args{
				algo:             Algorithm("hello"),
				expectedChecksum: "",
				filename:         "./testdata/hello.txt",
			},
			err: errs.ErrUnsupportedChecksumAlgorithm,
		},
		{
			name: "checksum not matched",
			args: args{
				algo:             SHA256,
				expectedChecksum: "hello",
				filename:         "./testdata/hello.txt",
			},
			err: errs.ErrChecksumNotMatched,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.err, VerifyFile(tt.args.algo, tt.args.expectedChecksum, tt.args.filename))
		})
	}
}
