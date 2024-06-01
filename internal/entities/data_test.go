package entities

import (
	"encoding/json"
	"io"
	"os"
	"testing"
)

func TestBlockUnmarshal(t *testing.T) {
	for _, tc := range tests {
		t.Run(tc.Title, func(t *testing.T) {
			f, err := os.Open(tc.Path)
			if err != nil {
				t.Fatal(err)
			}

			data, err := io.ReadAll(f)
			if err != nil {
				t.Fatal(err)
			}

			block := new(Block)

			err = json.Unmarshal(data, block)

			if tc.MustFail && err == nil {
				t.Fatal("unmarshall must fail")
			}

			if !tc.MustFail && err != nil {
				t.Fatalf("unmarshall must pass but it is failed. %s\n", err.Error())
			}
		})
	}
}

type TestCase struct {
	Title    string
	Path     string
	MustFail bool
}

var tests = []TestCase{
	{
		Title: "Valid block json object",
		Path:  "./test_data/test.block.valid.json",
	},
	{
		Title: "Valid block json object. No TXs",
		Path:  "./test_data/test.block.valid.notx.json",
	},
	{
		Title: "Valid block json object empty",
		Path:  "./test_data/test.block.valid.empty.json",
	},
	{
		Title:    "Valid block json object",
		Path:     "./test_data/test.block.invalid.corrupted.json",
		MustFail: true,
	},
}
