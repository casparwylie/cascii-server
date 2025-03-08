migrate -source file://./migrations -database "mysql://root:pass@tcp(localhost:3306)/ascii" up $1
