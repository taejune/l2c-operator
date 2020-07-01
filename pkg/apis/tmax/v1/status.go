package v1

type Status string

const (
	StatusSucceeded = Status("Succeeded")
	StatusReady     = Status("Ready")
	StatusFailed    = Status("Failed")
	StatusRunning   = Status("Running")
	StatusPending   = Status("Pending")
	StatusSkipped   = Status("Skipped")
)

type Phase string

const (
	PhaseNull      = Phase("")
	PhaseAnalyze   = Phase("analyze")
	PhaseDbDeploy  = Phase("db-deploy")
	PhaseDbMigrate = Phase("db-migrate")
	PhaseBuild     = Phase("build")
	PhaseTest      = Phase("test")
	PhaseDeploy    = Phase("deploy")
)

var Phases = []Phase{PhaseAnalyze, PhaseDbDeploy, PhaseDbMigrate, PhaseBuild, PhaseTest, PhaseDeploy}

type PhaseTaskRunMap struct {
	Phase       Phase
	TaskRunName string
}
