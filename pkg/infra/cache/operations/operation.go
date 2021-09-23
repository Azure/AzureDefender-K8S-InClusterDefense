package operations

type OPERATION string

const (
	SET OPERATION = "Set"
	GET OPERATION = "Get"
)

type OPERATION_STATUS string

const (
	HIT  OPERATION_STATUS = "Hit"
	MISS OPERATION_STATUS = "Miss"
)
