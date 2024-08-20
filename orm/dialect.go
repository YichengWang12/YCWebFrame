package orm

import "WebFrame/orm/internal/errs"

var (
	MySQL   Dialect = &mysqlDialect{}
	SQLite3 Dialect = &sqlite3Dialect{}
)

type Dialect interface {
	quoter() byte
	buildUpdate(b *builder, odk *Update) error
}

type standardSQL struct {
}

func (s standardSQL) quoter() byte {
	//TODO implement me
	panic("implement me")
}

func (s standardSQL) buildUpdate(b *builder, odk *Update) error {
	//TODO implement me
	panic("implement me")
}

type mysqlDialect struct {
	standardSQL
}

func (m *mysqlDialect) quoter() byte {
	return '`'
}

func (m *mysqlDialect) buildUpdate(b *builder, odk *Update) error {
	b.sb.WriteString(" ON DUPLICATE KEY UPDATE ")
	for idx, a := range odk.assigns {
		if idx > 0 {
			b.sb.WriteByte(',')
		}
		switch assign := a.(type) {
		case Column:
			fd, ok := b.model.FieldMap[assign.name]
			if !ok {
				return errs.NewErrUnknownField(assign.name)
			}
			b.quote(fd.ColName)
			b.sb.WriteString("=VALUES(")
			b.quote(fd.ColName)
			b.sb.WriteByte(')')
		case Assignment:
			err := b.buildColumn(assign.col)
			if err != nil {
				return err
			}
			b.sb.WriteString("=?")
			b.addArgs(assign.val)
		default:
			return errs.NewErrUnsupportedAssignableType(a)
		}
	}
	return nil
}

type sqlite3Dialect struct {
	standardSQL
}

func (m *sqlite3Dialect) quoter() byte {
	return '`'
}
