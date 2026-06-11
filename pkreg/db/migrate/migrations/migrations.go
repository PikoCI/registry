package migrations

type Migration struct {
	Name string
	SQL  string
}

var Migrations = [2]Migration{
	V0Initial,
	V1FTS,
}
