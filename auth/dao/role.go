package dao

const (
	//system admin
	SYSADMIN = 1
	//system member
	SYSMEMBER = 2

	//organization admin
	ORGADMIN = 3
	//organization member
	ORGMEMBER = 4

	//team admin
	TEAMADMIN = 5
	//team member
	TEAMMEMBER = 6
)

const (
	//Members will be able to  pull, push, and add new collaborators to all repositories.
	//ADMIN = 1
	//Members will be able to  pull, and push all repositories
	WRITE = 2
	//Members will be able pull all repositories.
	READ = 3
	//Members will only be able to  pull public repositories.
	//To give a member additional access,
	//you’ll need to add them to teams or make them collaborators on individual repositories.
	//only for organizaiton
	NONE = 4
)

const (
	StatusBadRequest   = "Bad Request"
	StatusUnauthorized = "Unauthorized"
	StatusNotFound     = "Not Found"
	//StatusInternalServerError = "Internal Server Error"
)

const (
	ACTIVE   = 0
	INACTIVE = 1
)
