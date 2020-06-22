package v1

const (
	TaskAnalyze   = "task-analyze"
	TaskDbDeploy  = "task-db-deploy"
	TaskDbMigrate = "task-db-migrate"
	TaskBuild     = "task-build"
	TaskTest      = "task-test"
	TaskDeploy    = "task-deploy"

	TaskRunAnalyze   = TaskAnalyze
	TaskRunDbDeploy  = TaskDbDeploy
	TaskRunDbMigrate = TaskDbMigrate
	TaskRunBuild     = TaskBuild
	TaskRunTest      = TaskTest
	TaskRunDeploy    = TaskDeploy
)

const (
	LabelL2cName     = "l2c.tmax.io/name"
	LabelL2cRunName  = "l2crun.tmax.io/name"
	LabelL2cRunPhase = "l2crun.tmax.io/phase"
)
