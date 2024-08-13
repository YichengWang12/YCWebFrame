package orm

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_parseModel(t *testing.T) {
	testCases := []struct {
		name      string
		val       any
		wantModel *model
		wantErr   error
	}{
		{
			name:    "test model",
			val:     TestModel{},
			wantErr: errors.New("orm: Only supports one-level pointer as input, such as *User"),
		},
		{
			// Pointer
			name: "pointer",
			val:  &TestModel{},
			wantModel: &model{
				tableName: "test_model",
				fieldMap: map[string]*field{
					"Id": {
						colName: "id",
					},
					"FirstName": {
						colName: "first_name",
					},
					"Age": {
						colName: "age",
					},
					"LastName": {
						colName: "last_name",
					},
				},
			},
		},
		{
			// Multiple pointers
			name: "multiple pointer",
			val: func() any {
				val := &TestModel{}
				return &val
			}(),
			wantErr: errors.New("orm: Only supports one-level pointer as input, such as *User"),
		},
		{
			name:    "map",
			val:     map[string]string{},
			wantErr: errors.New("orm: Only supports one-level pointer as input, such as *User"),
		},
		{
			name:    "slice",
			val:     []int{},
			wantErr: errors.New("orm: Only supports one-level pointer as input, such as *User"),
		},
		{
			name:    "basic type",
			val:     0,
			wantErr: errors.New("orm: Only supports one-level pointer as input, such as *User"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := parseModel(tc.val)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantModel, m)
		})
	}
}

func Test_underscoreName(t *testing.T) {
	testCases := []struct {
		name    string
		srcStr  string
		wantStr string
	}{
		{
			name:    "all uppercase",
			srcStr:  "ID",
			wantStr: "i_d",
		},
		{
			name:    "all lowercase",
			srcStr:  "id",
			wantStr: "id",
		},
		{
			name:    "use number",
			srcStr:  "Table1Name",
			wantStr: "table1_name",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantStr, underscoreName(tc.srcStr))
		})
	}
}
