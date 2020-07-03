package v1

const (
	TaskAnalyze          = "task-analyze"
	TaskAnalyzeJavaMaven = "sonar-scan-java-maven"
	TaskDbDeploy         = "l2c-deploy-db"
	TaskDbMigrate        = "task-db-migrate"
	TaskDbMigrateTibero  = "tup-tibero"
	TaskBuild            = "s2i"
	TaskTest             = "analyze-image-vulnerabilities"
	TaskDeploy           = "generate-and-deploy-using-kubectl"
)

const (
	LabelL2cName     = "l2c.tmax.io/name"
	LabelL2cRunName  = "l2crun.tmax.io/name"
	LabelL2cRunPhase = "l2crun.tmax.io/phase"
)

const (
	KeyDbPvc    = "db-pvc.yaml"
	KeyDbSvc    = "db-svc.yaml"
	KeyDbSecret = "db-secret.yaml"
	KeyDbDeploy = "db-deploy.yaml"
)
