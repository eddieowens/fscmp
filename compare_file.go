package fscmp

import (
	"io/fs"
	"strings"
)

type compareFileOpts struct {
	IgnoreLineSpaces bool
	IgnoreFileSpaces bool
}

func compare(expectedFs, actualFs fs.FS, fp string, o compareFileOpts) *FileDiff {
	fileDiff := &FileDiff{
		Path: fp,
	}
	expected, err := expectedFs.Open(fp)
	if err != nil {
		fileDiff.Error = err
		return fileDiff
	}

	actual, err := actualFs.Open(fp)
	if err != nil {
		fileDiff.Error = err
		fileDiff.Path = fp
		return fileDiff
	}

	expectedScanner := NewScanner(expected)
	actualScanner := NewScanner(actual)
	for eScan, aScan := expectedScanner.Scan(), actualScanner.Scan(); eScan == aScan && eScan; eScan, aScan = expectedScanner.Scan(), actualScanner.Scan() {
		var expectedCon, actualCon string
		if o.IgnoreFileSpaces {
			expectedCon = expectedScanner.NextNonEmptyLine()
			actualCon = actualScanner.NextNonEmptyLine()
		} else {
			expectedCon = expectedScanner.Text()
			actualCon = actualScanner.Text()
		}

		if o.IgnoreLineSpaces {
			expectedCon = strings.TrimSpace(expectedCon)
			actualCon = strings.TrimSpace(actualCon)
		}

		if expectedCon != actualCon {
			fileDiff.Diffs = append(fileDiff.Diffs, LineDiff{
				ExpectedContent: expectedCon,
				ActualContent:   actualCon,
				ExpectedLineNum: expectedScanner.LineNum,
				ActualLineNum:   actualScanner.LineNum,
			})
		}
	}

	// drain the scanners
	for expectedScanner.Scan() {
		text := expectedScanner.Text()
		if !o.IgnoreFileSpaces || text != "" {
			fileDiff.Diffs = append(fileDiff.Diffs, LineDiff{
				ExpectedContent: text,
				ExpectedLineNum: expectedScanner.LineNum,
			})
		}
	}

	for actualScanner.Scan() {
		text := actualScanner.Text()
		if !o.IgnoreFileSpaces || text != "" {
			fileDiff.Diffs = append(fileDiff.Diffs, LineDiff{
				ActualContent: text,
				ActualLineNum: actualScanner.LineNum,
			})
		}
	}

	if len(fileDiff.Diffs) == 0 {
		return nil
	}

	return fileDiff
}
