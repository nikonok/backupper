package helpers_test

import (
	"testing"

	"github.com/nikonok/backupper/helpers"
)

func TestParseScheduledDelete(t *testing.T) {
	tests := []struct {
		filename   string
		date       string
		scheduled  string
		shouldPass bool
	}{
		{
			"delete_2023-09-26T12:30:00+05:00_extra.txt",
			"2023-09-26T12:30:00+05:00",
			"extra.txt",
			true,
		},
		{
			"delete_2023-09-26T12:30:00+05:00",
			"2023-09-26T12:30:00+05:00",
			"",
			false,
		},
		{
			"delete_2023-09-23T18:54:30+02:00_file_1",
			"2023-09-23T18:54:30+02:00",
			"file_1",
			true,
		},
		{
			"2023-09-26T12:30:00+05:00_delete.txt",
			"",
			"",
			false,
		},
		{
			"delete_2023-09-26T12:30:00.txt",
			"",
			"",
			false,
		},
		{
			"delete_2023-09-23T18:54:30+02:00_file 1",
			"2023-09-23T18:54:30+02:00",
			"file 1",
			true,
		},
		{
			"delete_file_1",
			"",
			"",
			false,
		},
		{
			"delete_file 1",
			"",
			"",
			false,
		},
		{
			"_2023-09-23T18:54:30+02:00_file_1",
			"",
			"",
			false,
		},
	}

	for _, test := range tests {
		date, scheduled, ok := helpers.ParseScheduledDelete(test.filename)

		if ok != test.shouldPass {
			t.Errorf("Expected %v for filename %v, but got %v", test.shouldPass, test.filename, ok)
			continue
		}

		if ok && (date != test.date || scheduled != test.scheduled) {
			t.Errorf("For filename %v: expected date %v and scheduled %v, but got date %v and scheduled %v", test.filename, test.date, test.scheduled, date, scheduled)
		}
	}
}
