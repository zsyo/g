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

package cli

import (
	"fmt"
	"os"
	"sort"

	"github.com/k0kubun/go-ansi"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"github.com/voidint/g/version"
)

func list(ctx *cli.Context) (err error) {
	items, err := listLocalVersions(versionsDir)
	if err != nil || len(items) <= 0 {
		fmt.Printf("No version installed yet\n\n")
		return nil
	}

	var renderMode uint8
	switch ctx.String("output") {
	case "json":
		renderMode = jsonMode
	default:
		renderMode = textMode
	}

	render(renderMode, installed(), items, ansi.NewAnsiStdout())
	return nil
}

// listLocalVersions List the versions in the specified directory in ascending order
func listLocalVersions(dirPath string) ([]*version.Version, error) {
	dirs, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	items := make([]*version.Version, 0, len(dirs))
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}

		v, err := version.New(d.Name())
		if err != nil || v == nil {
			continue
		}
		items = append(items, v)
	}
	sort.Sort(version.Collection(items)) // asc order
	return items, nil
}
