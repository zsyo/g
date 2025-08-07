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
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	ct "github.com/daviddengcn/go-colortext"
	"github.com/dixonwille/wlog/v3"
	"github.com/dixonwille/wmenu/v5"
	"github.com/mholt/archiver/v3"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"github.com/voidint/g/collector"
	"github.com/voidint/g/version"
)

func install(ctx *cli.Context) (err error) {
	vname := ctx.Args().First()
	if vname == "" {
		return cli.ShowSubcommandHelp(ctx)
	}

	// Find matching Go version.
	c, err := collector.NewCollector(strings.Split(os.Getenv(mirrorEnv), mirrorSep)...)
	if err != nil {
		return cli.Exit(errstring(err), 1)
	}
	items, err := c.AllVersions()
	if err != nil {
		return cli.Exit(errstring(err), 1)
	}

	v, err := version.NewFinder(items,
		version.WithFinderPackageKind(version.ArchiveKind),
		version.WithFinderGoos(runtime.GOOS),
		version.WithFinderGoarch(runtime.GOARCH),
	).Find(vname)
	if err != nil {
		return cli.Exit(errstring(err), 1)
	}

	vname = v.Name()
	targetV := filepath.Join(versionsDir, vname)

	// Check if the version is already installed.
	var finfo os.FileInfo
	if finfo, err = os.Stat(targetV); err == nil && finfo.IsDir() {
		return cli.Exit(fmt.Sprintf("[g] %q version has been installed.", vname), 1)
	}

	// Find installation packages for current platform
	pkgs, err := v.FindPackages(version.ArchiveKind, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return cli.Exit(errstring(err), 1)
	}
	var pkg version.Package
	if len(pkgs) > 1 {
		menu := wmenu.NewMenu("Please select the package you want to install.")
		menu.AddColor(
			wlog.Color{Code: ct.Green},
			wlog.Color{Code: ct.Yellow},
			wlog.Color{Code: ct.Magenta},
			wlog.Color{Code: ct.Yellow},
		)
		menu.Action(func(opts []wmenu.Opt) error {
			pkg = opts[0].Value.(version.Package)
			return nil
		})
		for i := range pkgs {
			if i == 0 {
				menu.Option(pkgs[i].FileName, pkgs[i], true, nil)
			} else {
				menu.Option(" "+pkgs[i].FileName, pkgs[i], false, nil)
			}
		}
		if err = menu.Run(); err != nil {
			return cli.Exit(errstring(err), 1)
		}
	} else {
		pkg = pkgs[0]
	}

	skipChecksum := ctx.Bool("skip-checksum")

	if !skipChecksum {
		var checksumNotFound bool
		if pkg.Checksum == "" && pkg.ChecksumURL == "" {
			checksumNotFound = true
			menu := wmenu.NewMenu("Checksum file not found, do you want to continue?")
			menu.IsYesNo(wmenu.DefN)
			menu.Action(func(opts []wmenu.Opt) error {
				skipChecksum = opts[0].Value.(string) == "yes"
				return nil
			})
			if err = menu.Run(); err != nil {
				return cli.Exit(errstring(err), 1)
			}
		}
		if checksumNotFound && !skipChecksum {
			return
		}
	}

	var ext string
	if runtime.GOOS == "windows" {
		ext = "zip"
	} else {
		ext = "tar.gz"
	}
	filename := filepath.Join(downloadsDir, fmt.Sprintf("go%s.%s-%s.%s", vname, runtime.GOOS, runtime.GOARCH, ext))

	if _, err = os.Stat(filename); os.IsNotExist(err) {
		// Download package remotely and verify checksum.
		if _, err = pkg.DownloadWithProgress(filename); err != nil {
			return cli.Exit(errstring(err), 1)
		}

		if !skipChecksum {
			fmt.Println("Computing checksum with", pkg.Algorithm)
			if err = pkg.VerifyChecksum(filename); err != nil {
				return cli.Exit(errstring(err), 1)
			}
			fmt.Println("Checksums matched")
		}

	} else {
		if !skipChecksum {
			// Verify checksum for local package.
			fmt.Println("Computing checksum with", pkg.Algorithm)
			if err = pkg.VerifyChecksum(filename); err != nil {
				_ = os.Remove(filename)
				return cli.Exit(errstring(err), 1)
			}
			fmt.Println("Checksums matched")
		}
	}

	// Clean up legacy files.
	_ = os.RemoveAll(filepath.Join(versionsDir, "go"))

	// Extract installation archive.
	if err = archiver.Unarchive(filename, versionsDir); err != nil {
		return cli.Exit(errstring(err), 1)
	}
	// Rename version directory.
	if err = os.Rename(filepath.Join(versionsDir, "go"), targetV); err != nil {
		return cli.Exit(errstring(err), 1)
	}

	if ctx.Bool("nouse") {
		return nil
	}

	if err = switchVersion(vname); err != nil {
		return cli.Exit(errstring(err), 1)
	}

	// 如果开启了拷贝模式，重新拷贝一份新版本
	if gcopy() {
		_ = os.RemoveAll(copyroot)

		if err = copyDir(targetV, copyroot); err != nil {
			return cli.Exit(errstring(err), 1)
		}
	}

	return nil
}

func switchVersion(vname string) error {
	targetV := filepath.Join(versionsDir, vname)

	// Recreate symbolic link
	_ = os.Remove(goroot)

	if err := mkSymlink(targetV, goroot); err != nil {
		return errors.WithStack(err)
	}
	if output, err := exec.Command(filepath.Join(goroot, "bin", "go"), "version").Output(); err == nil {
		fmt.Printf("Now using %s", strings.TrimPrefix(string(output), "go version "))
	}
	return nil
}

func mkSymlink(oldname, newname string) (err error) {
	if runtime.GOOS == "windows" {
		// Windows 10下无特权用户无法创建符号链接，优先调用mklink /j创建'目录联接'
		if err = exec.Command("cmd", "/c", "mklink", "/j", newname, oldname).Run(); err == nil {
			return nil
		}
	}
	if err = os.Symlink(oldname, newname); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// copyDir 拷贝一个目录及其子目录和文件到另一个目录
func copyDir(src string, dest string) error {
	// 获取源目录的属性
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// 创建目标目录
	err = os.MkdirAll(dest, srcInfo.Mode())
	if err != nil {
		return err
	}

	// 读取源目录下的所有文件和子目录
	dir, err := os.Open(src)
	if err != nil {
		return err
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		return err
	}

	for _, file := range files {
		srcPath := filepath.Join(src, file.Name())
		destPath := filepath.Join(dest, file.Name())

		if file.IsDir() {
			// 如果是子目录，递归拷贝
			err = copyDir(srcPath, destPath)
			if err != nil {
				return err
			}
		} else {
			// 如果是文件，直接拷贝
			err = copyFile(srcPath, destPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// copyFile 拷贝一个文件到另一个文件
func copyFile(src string, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile) // 拷贝数据
	if err != nil {
		return err
	}

	srcInfo, err := os.Stat(src) // 获取源文件的属性
	if err != nil {
		return err
	}
	err = os.Chmod(dest, srcInfo.Mode()) // 设置目标文件的属性和源文件一致
	if err != nil {
		return err
	}

	return nil
}
