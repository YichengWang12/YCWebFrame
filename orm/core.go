package orm

import (
	"WebFrame/orm/internal/valuer"
	"WebFrame/orm/model"
)

type core struct {
	r          model.Registry
	dialect    Dialect
	valCreator valuer.Creator
	model      *model.Model
}
