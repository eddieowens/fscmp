package fscmp

import (
	"fmt"
	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/suite"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

type PublicTestSuite struct {
	suite.Suite

	TestFolderBase     string
	TestFolderBaseName string
	TestFolder         string
}

func (p *PublicTestSuite) AfterTest(_, _ string) {
	_ = os.RemoveAll(p.TestFolderBase)
}

func (p *PublicTestSuite) BeforeTest(_, _ string) {
	p.TestFolderBaseName = faker.Username()
	p.TestFolderBase = filepath.Join(os.TempDir(), p.TestFolderBaseName)
	p.TestFolder = filepath.Join(p.TestFolderBase, "test")
	_ = os.MkdirAll(p.TestFolder, os.ModePerm)
}

func (p *PublicTestSuite) TestEqualsFolders() {
	// -- Given
	//
	testFilename := faker.Username() + ".txt"
	testFile := filepath.Join(p.TestFolder, testFilename)
	_ = os.WriteFile(testFile, []byte(`this is a text file`), os.ModePerm)

	expected := fstest.MapFS{
		filepath.Join("test", testFilename): {
			Data: []byte(`this is a text file`),
		},
	}

	// -- When
	//
	diff, err := EqualFilesystems(expected, os.DirFS(p.TestFolderBase))

	// -- Then
	//
	if p.NoError(err) {
		p.Empty(diff.String())
	}
}

func (p *PublicTestSuite) TestEqualsFoldersDiff() {
	// -- Given
	//
	testFilename := faker.Username() + ".txt"
	testFile := filepath.Join(p.TestFolder, testFilename)
	relative := filepath.Join("test", testFilename)
	_ = os.WriteFile(testFile, []byte(`this is a text fil`), os.ModePerm)

	expectedFs := fstest.MapFS{
		relative: {
			Data: []byte(`this is a text file`),
		},
	}

	expectedDiff := FsDiff{
		Diffs: []FileDiff{
			{
				Path: relative,
				Diffs: []LineDiff{
					{
						ActualLineNum:   1,
						ExpectedLineNum: 1,
						ActualContent:   `this is a text fil`,
						ExpectedContent: `this is a text file`,
					},
				},
			},
		},
	}

	// -- When
	//
	diff, err := EqualFilesystems(expectedFs, os.DirFS(p.TestFolderBase))

	// -- Then
	//
	if p.NoError(err) {
		p.Equal(expectedDiff, diff)
		p.Equal(fmt.Sprintf("Differences found in filesystem\n\nFile test/%s:\n\tExpected@1: this is a text file ; Actual@1: this is a text fil\n", testFilename), diff.String())
	}
}

func (p *PublicTestSuite) TestEqualsFoldersIgnoreFileSpace() {
	// -- Given
	//

	type test struct {
		given        string
		expected     string
		expectedDiff *FsDiff
	}
	testFilename := faker.Username() + ".txt"
	testFile := filepath.Join(p.TestFolder, testFilename)
	defer os.Remove(testFile)
	relative := filepath.Join("test", testFilename)

	tests := []test{
		{
			given: `
this is a text file
`,
			expected: `this is a text file`,
		},
		{
			given: `
this is a text file
more text
`,
			expected: `
this is a text file

more text


`,
		},
		{
			given: `
this is a text file
more text
`,
			expected: `
this is a text file

more text

should fail
`,
			expectedDiff: &FsDiff{
				Diffs: []FileDiff{
					{
						Path: relative,
						Diffs: []LineDiff{
							{
								ExpectedContent: "should fail",
								ExpectedLineNum: 6,
							},
						},
					},
				},
			},
		},
		{
			given: `
this is a text file
more text




should fail


`,
			expected: `
this is a text file

more text
`,
			expectedDiff: &FsDiff{
				Diffs: []FileDiff{
					{
						Path: relative,
						Diffs: []LineDiff{
							{
								ActualContent: "should fail",
								ActualLineNum: 8,
							},
						},
					},
				},
			},
		},
	}

	// -- When
	//
	for i, v := range tests {
		expectedFs := fstest.MapFS{
			relative: {
				Data: []byte(v.expected),
			},
		}
		_ = os.WriteFile(testFile, []byte(v.given), os.ModePerm)

		diff, err := EqualFilesystems(expectedFs, os.DirFS(p.TestFolderBase), IgnoreFileSpaces())

		// -- Then
		//
		if p.NoError(err) {
			if v.expectedDiff == nil {
				v.expectedDiff = &FsDiff{}
			}
			p.Equalf(*v.expectedDiff, diff, "test %d", i)
		}
	}
}

func (p *PublicTestSuite) TestEqualsIgnoreLineSpaces() {
	// -- Given
	//
	testFilename := faker.Username() + ".txt"
	testFile := filepath.Join(p.TestFolder, testFilename)
	_ = os.WriteFile(testFile, []byte(`    this is a text file    `), os.ModePerm)

	expected := fstest.MapFS{
		filepath.Join("test", testFilename): {
			Data: []byte(`this is a text file`),
		},
	}

	// -- When
	//
	diff, err := EqualFilesystems(expected, os.DirFS(p.TestFolderBase), IgnoreLineSpaces())

	// -- Then
	//
	if p.NoError(err) {
		p.Equal(FsDiff{}, diff)
	}
}

func TestPublicTestSuite(t *testing.T) {
	suite.Run(t, new(PublicTestSuite))
}
