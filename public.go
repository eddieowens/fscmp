package fscmp

import (
	"fmt"
	"io/fs"
	"strings"
)

type FsDiff struct {
	Diffs []FileDiff
}

func (f FsDiff) String() string {
	if len(f.Diffs) > 0 {
		builder := strings.Builder{}
		builder.WriteString("Differences found in filesystem\n\n")
		for _, v := range f.Diffs {
			builder.WriteString(v.String())
		}
		return builder.String()
	}
	return ""
}

type FileDiff struct {
	Path  string
	Error error
	Diffs []LineDiff
}

func (f FileDiff) String() string {
	if len(f.Diffs) > 0 || f.Error != nil {
		builder := strings.Builder{}
		builder.WriteString(fmt.Sprintf("File %s:\n", f.Path))
		if f.Error != nil {
			builder.WriteString(fmt.Sprintf("Failed to open file: %s\n", f.Error.Error()))
		}

		for _, v := range f.Diffs {
			builder.WriteString(v.String())
		}

		return builder.String()
	}

	return ""
}

type LineDiff struct {
	ExpectedContent string
	ActualContent   string
	ExpectedLineNum int
	ActualLineNum   int
}

func (l LineDiff) String() string {
	if l.ExpectedContent != "" || l.ActualContent != "" {
		return fmt.Sprintf("\tExpected@%d: %s ; Actual@%d: %s\n", l.ExpectedLineNum, l.ExpectedContent, l.ActualLineNum, l.ActualContent)
	}
	return ""
}

type Opt func(o *equalOpts)

func IgnoreLineSpaces() Opt {
	return func(o *equalOpts) {
		o.IgnoreLineSpaces = true
	}
}

func IgnoreFileSpaces() Opt {
	return func(o *equalOpts) {
		o.IgnoreFileSpaces = true
	}
}

type equalOpts struct {
	compareFileOpts
}

func EqualFilesystems(expected, actual fs.FS, opts ...Opt) (FsDiff, error) {
	o := &equalOpts{}
	for _, v := range opts {
		v(o)
	}

	diff := FsDiff{}
	err := fs.WalkDir(expected, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		fileDiff := compare(expected, actual, path, o.compareFileOpts)

		if fileDiff != nil {
			diff.Diffs = append(diff.Diffs, *fileDiff)
		}

		return nil
	})
	return diff, err
}
