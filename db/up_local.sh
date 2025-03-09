migrate -source file://./migrations -database "mysql://root:pass@tcp(localhost:3306)/cascii" up $1
