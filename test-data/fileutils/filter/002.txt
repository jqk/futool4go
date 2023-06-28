package filediff

import (
	"testing"
)

var content string = `bufferSize: 1024
baseGroup: "group 0"
groups:
  - name: "group 0"
    action: 1
    datafile: data1.txt
    paths:
      - test-data/origin/compare_base_0
      - test-data/origin/compare_base_1
  - name: "group 1"
    action: 2
    datafile: data2.txt
    paths:
      - test-data/origin/compare_target
filter:
  ignoreExtensionCase: true
  include:
    - "*.txt"
    - "*.md"
  exclude:
    - "*.log"
  minFileSize: 1024
  maxFileSize: 1048576
resultPath: test-data/output`

var expected Config = Config{
	BufferSize: 1024,
	BaseGroup:  "group 0",
	Groups: []Group{
		{
			Name:     "group 0",
			Action:   1,
			DataFile: "data1.txt",
			Paths: []string{
				"test-data/origin/compare_base_0",
				"test-data/origin/compare_base_1",
			},
		},
		{
			Name:     "group 1",
			Action:   2,
			DataFile: "data2.txt",
			Paths: []string{
				"test-data/origin/compare_target",
			},
		},
	},
	Filter: Filter{
		IgnoreExtensionCase: true,
		Include: []string{
			"*.txt",
			"*.md",
		},
		Exclude: []string{
			"*.log",
		},
		MinFileSize: 1024,
		MaxFileSize: 1048576,
	},
	ResultPath: "test-data/output",
}

func TestLoadConfigFromString(t *testing.T) {
	config, err := LoadConfigFromString(content, "yaml")
	if err != nil {
		t.Errorf("LoadConfigFromString() returned error: %v", err)
	}

	if s := config.Diff(&expected); s != "" {
		t.Errorf("Diff found: %s\nLoadConfigFromString() returned:\n%+v\nExpected:\n%+v", s, *config, expected)
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	config, err := LoadConfigFromFile("test-data/config.yaml")
	if err != nil {
		t.Errorf("LoadConfigFromFile() returned error: %v", err)
	}

	if s := config.Diff(&expected); s != "" {
		t.Errorf("Diff found: %s\nLoadConfigFromFile() returned:\n%+v\nExpected:\n%+v", s, *config, expected)
	}
}
