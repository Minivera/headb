//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package enum

import "github.com/go-jet/jet/v2/postgres"

var UserStatus = &struct {
	Pending  postgres.StringExpression
	Accepted postgres.StringExpression
	Denied   postgres.StringExpression
}{
	Pending:  postgres.NewEnumValue("pending"),
	Accepted: postgres.NewEnumValue("accepted"),
	Denied:   postgres.NewEnumValue("denied"),
}
