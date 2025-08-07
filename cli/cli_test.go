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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/voidint/g/version"
)

func Test_ghome(t *testing.T) {
	t.Run("查询ghome路径", func(t *testing.T) {
		home, err := os.UserHomeDir()
		assert.Nil(t, err)
		assert.Equal(t, filepath.Join(home, ".g"), ghome())
	})
}

func Test_inuse(t *testing.T) {
	t.Run("查询当前使用中的go版本", func(t *testing.T) {
		rootDir := filepath.Join(os.TempDir(), fmt.Sprintf(".g_%d", time.Now().Unix()))
		goroot = filepath.Join(rootDir, "go")
		versionsDir = filepath.Join(rootDir, "versions")
		vDir := filepath.Join(versionsDir, "1.12.6")

		_ = os.MkdirAll(versionsDir, 0755)
		_ = os.MkdirAll(vDir, 0755)
		defer os.RemoveAll(rootDir)

		assert.Nil(t, mkSymlink(vDir, goroot))
		assert.Equal(t, "1.12.6", inuse(goroot))
	})
}

func Test_render(t *testing.T) {
	t.Run("渲染go版本列表(text)", func(t *testing.T) {
		var got strings.Builder
		items := []*version.Version{
			version.MustNew("1.19beta1"),
			version.MustNew("1.10beta2"),
			version.MustNew("1.7"),
			version.MustNew("1.8.1"),
			version.MustNew("1.21.0"),
			version.MustNew("1.21rc4"),
		}
		sort.Sort(version.Collection(items))

		render(textMode, map[string]bool{"1.8.1": true}, items, &got)
		assert.Equal(t, "  1.7\n* 1.8.1\n  1.10beta2\n  1.19beta1\n  1.21rc4\n  1.21.0\n", got.String())
	})

	t.Run("渲染go版本列表(json)", func(t *testing.T) {
		var actual strings.Builder
		items := []*version.Version{
			version.MustNew("1.19beta1"),
			version.MustNew("1.10beta2"),
			version.MustNew("1.7"),
			version.MustNew("1.8.1"),
			version.MustNew("1.21.0"),
			version.MustNew("1.21rc4"),
		}
		sort.Sort(version.Collection(items))

		installed := map[string]bool{"1.8.1": true}
		render(jsonMode, installed, items, &actual)

		vs := make([]versionOut, 0, len(items))
		for _, item := range items {
			vo := versionOut{
				Version:  item.Name(),
				Packages: item.Packages(),
			}
			if inuse, found := installed[item.Name()]; found {
				vo.InUse = inuse
				vo.Installed = found
			}
			vs = append(vs, vo)
		}

		var expected strings.Builder
		enc := json.NewEncoder(&expected)
		enc.SetIndent("", "    ")
		_ = enc.Encode(&vs)
		assert.Equal(t, expected.String(), actual.String())
	})
}

func Test_wrapstring(t *testing.T) {
	t.Run("包装字符串", func(t *testing.T) {
		assert.Equal(t, "[g] Hello world", wrapstring("hello world"))
	})
}

func Test_errstring(t *testing.T) {
	t.Run("返回错误字符串", func(t *testing.T) {
		assert.Equal(t, "", errstring(nil))
		assert.Equal(t, "[g] Hello world", errstring(errors.New("hello world")))
	})
}
