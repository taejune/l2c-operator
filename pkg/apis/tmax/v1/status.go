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
	PhaseAnalyze   = Phase("Analyze")
	PhaseDbDeploy  = Phase("DB-Deploy")
	PhaseDbMigrate = Phase("DB-Migrate")
	PhaseBuild     = Phase("Build")
	PhaseTest      = Phase("Test")
	PhaseDeploy    = Phase("Deploy")
)

var Phases = []Phase{PhaseAnalyze, PhaseDbDeploy, PhaseDbMigrate, PhaseBuild, PhaseTest, PhaseDeploy}
var PhaseTaskRuns = map[Phase]string{
	PhaseAnalyze:   TaskRunAnalyze,
	PhaseDbDeploy:  TaskRunDbDeploy,
	PhaseDbMigrate: TaskRunDbMigrate,
	PhaseBuild:     TaskRunBuild,
	PhaseTest:      TaskRunTest,
	PhaseDeploy:    TaskRunDeploy,
}

type PhaseTaskRunMap struct {
	Phase       Phase
	TaskRunName string
}
