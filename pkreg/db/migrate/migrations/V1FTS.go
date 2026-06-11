package migrations

var V1FTS = Migration{
	Name: "FTS",
	SQL:  `-- FTS indexes are handled at the application layer using LIKE queries`,
}
