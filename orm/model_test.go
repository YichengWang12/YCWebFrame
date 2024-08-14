package orm

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestModelWithTableName(t *testing.T) {
	testCases := []struct {
		name          string
		val           any
		opt           ModelOpt
		wantTableName string
		wantErr       error
	}{
		{
			// We do not check empty strings
			name:          "empty string",
			val:           &TestModel{},
			opt:           ModelWithTableName(""),
			wantTableName: "",
		},
		{
			name:          "table name",
			val:           &TestModel{},
			opt:           ModelWithTableName("test_model_t"),
			wantTableName: "test_model_t",
		},
	}

	r := NewRegistry()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := r.Register(tc.val, tc.opt)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantTableName, m.tableName)
		})
	}

}

func TestModelWithColumnName(t *testing.T) {
	testCases := []struct {
		name        string
		val         any
		opt         ModelOpt
		field       string
		wantColName string
		wantErr     error
	}{
		{
			name:        "new name",
			val:         &TestModel{},
			opt:         ModelWithColumnName("FirstName", "first_name_new"),
			field:       "FirstName",
			wantColName: "first_name_new",
		},
		{
			name:        "empty new name",
			val:         &TestModel{},
			opt:         ModelWithColumnName("FirstName", ""),
			field:       "FirstName",
			wantColName: "",
		},
	}

	r := NewRegistry()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := r.Register(tc.val, tc.opt)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantColName, m.fieldMap[tc.field].colName)
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
