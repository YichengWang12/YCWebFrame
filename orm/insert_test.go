package orm

import (
	"WebFrame/orm/internal/errs"
	"database/sql"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInserter_Build(t *testing.T) {
	db := memoryDB(t)
	testCases := []struct {
		name      string
		q         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			// insert nothing
			name:    "no value",
			q:       NewInserter[TestModel](db).Values(),
			wantErr: errs.ErrInsertZeroRow,
		},
		{
			name: "single values",
			q: NewInserter[TestModel](db).Values(
				&TestModel{
					Id:        1,
					FirstName: "Yicheng",
					Age:       20,
					LastName: &sql.NullString{
						String: "Wang",
						Valid:  true,
					},
				},
			),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`, `first_name`, `age`, `last_name`) VALUES(?,?,?,?);",
				Args: []any{int64(1), "Yicheng", int8(20), &sql.NullString{
					String: "Wang",
					Valid:  true,
				}},
			},
		},
		{
			name: "multiple values",
			q: NewInserter[TestModel](db).Values(
				&TestModel{
					Id:        2,
					FirstName: "Yicheng",
					Age:       20,
					LastName: &sql.NullString{
						String: "Wang",
						Valid:  true,
					},
				},
				&TestModel{
					Id:        3,
					FirstName: "XXX",
					Age:       11,
					LastName: &sql.NullString{
						String: "TT",
						Valid:  true,
					},
				},
			),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`, `first_name`, `age`, `last_name`) VALUES(?,?,?,?),(?,?,?,?);",
				Args: []any{int64(2), "Yicheng", int8(20), &sql.NullString{
					String: "Wang",
					Valid:  true,
				}, int64(3), "XXX", int8(11), &sql.NullString{
					String: "TT",
					Valid:  true,
				}},
			},
		},
		{
			// specify columns
			name: "specify columns",
			q: NewInserter[TestModel](db).Values(
				&TestModel{
					Id:        4,
					FirstName: "Deng",
					Age:       18,
					LastName:  &sql.NullString{String: "Ming", Valid: true},
				}).Columns("FirstName", "LastName"),
			wantQuery: &Query{
				SQL:  "INSERT INTO `test_model`(`first_name`, `last_name`) VALUES(?,?);",
				Args: []any{"Deng", &sql.NullString{String: "Ming", Valid: true}},
			},
		},
		{
			// specify columns
			name: "invalid columns",
			q: NewInserter[TestModel](db).Values(
				&TestModel{
					Id:        1,
					FirstName: "Deng",
					Age:       18,
					LastName:  &sql.NullString{String: "Ming", Valid: true},
				}).Columns("FirstName", "Invalid"),
			wantErr: errs.NewErrUnknownField("Invalid"),
		},
		{
			// upsert
			name: "upsert",
			q: NewInserter[TestModel](db).Values(
				&TestModel{
					Id:        1,
					FirstName: "Deng",
					Age:       18,
					LastName:  &sql.NullString{String: "Ming", Valid: true},
				}).OnDuplicateKey().Update(Assign("FirstName", "Da")),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`, `first_name`, `age`, `last_name`) VALUES(?,?,?,?) " +
					"ON DUPLICATE KEY UPDATE `first_name`= ?;",
				Args: []any{int64(1), "Deng", int8(18), &sql.NullString{String: "Ming", Valid: true}, "Da"},
			},
		},
		{
			// upsert invalid column
			name: "upsert invalid column",
			q: NewInserter[TestModel](db).Values(
				&TestModel{
					Id:        1,
					FirstName: "Deng",
					Age:       18,
					LastName:  &sql.NullString{String: "Ming", Valid: true},
				}).OnDuplicateKey().Update(Assign("Invalid", "Da")),
			wantErr: errs.NewErrUnknownField("Invalid"),
		},
		{
			// 使用原本插入的值
			name: "upsert use insert value",
			q: NewInserter[TestModel](db).Values(
				&TestModel{
					Id:        1,
					FirstName: "Deng",
					Age:       18,
					LastName:  &sql.NullString{String: "Ming", Valid: true},
				},
				&TestModel{
					Id:        2,
					FirstName: "Da",
					Age:       19,
					LastName:  &sql.NullString{String: "Ming", Valid: true},
				}).OnDuplicateKey().Update(C("FirstName"), C("LastName")),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model`(`id`, `first_name`, `age`, `last_name`) VALUES(?,?,?,?),(?,?,?,?) " +
					"ON DUPLICATE KEY UPDATE `first_name`=VALUES(`first_name`),`last_name`=VALUES(`last_name`);",
				Args: []any{int64(1), "Deng", int8(18), &sql.NullString{String: "Ming", Valid: true},
					int64(2), "Da", int8(19), &sql.NullString{String: "Ming", Valid: true}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query, err := tc.q.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, query)
		})
	}
}
