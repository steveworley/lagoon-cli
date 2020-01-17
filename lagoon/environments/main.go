package environments

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/amazeeio/lagoon-cli/api"
	"github.com/amazeeio/lagoon-cli/graphql"
	"github.com/amazeeio/lagoon-cli/output"
)

// Environments .
type Environments struct {
	debug bool
	api   api.Client
}

// Client .
type Client interface {
	DeployEnvironmentBranch(string, string) ([]byte, error)
	DeleteEnvironment(string, string) ([]byte, error)
	GetDeploymentLog(string) ([]byte, error)
	GetEnvironmentInfo(string, string) ([]byte, error)
	ListEnvironmentVariables(string, string, bool) ([]byte, error)
	GetEnvironmentDeployments(string, string) ([]byte, error)
	GetEnvironmentTasks(string, string) ([]byte, error)
	RunDrushArchiveDump(string, string) ([]byte, error)
	RunDrushSQLDump(string, string) ([]byte, error)
	RunDrushCacheClear(string, string) ([]byte, error)
	RunCustomTask(string, string, api.Task) ([]byte, error)
	AddEnvironmentVariableToEnvironment(string, string, api.EnvVariable) ([]byte, error)
	DeleteEnvironmentVariableFromEnvironment(string, string, api.EnvVariable) ([]byte, error)
}

// New .
func New(debug bool) (Client, error) {
	lagoonAPI, err := graphql.LagoonAPI(debug)
	if err != nil {
		return &Environments{}, err
	}
	return &Environments{
		debug: debug,
		api:   lagoonAPI,
	}, nil

}

var noDataError = "no data returned from the lagoon api"

// DeployEnvironmentBranch .
func (e *Environments) DeployEnvironmentBranch(projectName string, branchName string) ([]byte, error) {
	customRequest := api.CustomRequest{
		Query: `mutation ($project: String!, $branch: String!){
			deployEnvironmentBranch(
				input: {
					project:{name: $project}
					branchName: $branch
				}
			)
		}`,
		Variables: map[string]interface{}{
			"project": projectName,
			"branch":  branchName,
		},
		MappedResult: "deployEnvironmentBranch",
	}
	returnResult, err := e.api.Request(customRequest)
	return returnResult, err
}

// DeleteEnvironment .
func (e *Environments) DeleteEnvironment(projectName string, environmentName string) ([]byte, error) {
	evironment := api.DeleteEnvironment{
		Name:    environmentName,
		Project: projectName,
		Execute: true,
	}
	returnResult, err := e.api.DeleteEnvironment(evironment)
	return returnResult, err
}

// GetEnvironmentInfo will get basic info about a project
func (e *Environments) GetEnvironmentInfo(projectName string, environmentName string) ([]byte, error) {
	// get project info from lagoon
	project := api.Project{
		Name: projectName,
	}
	projectByName, err := e.api.GetProjectByName(project, graphql.ProjectByNameFragment)
	if err != nil {
		return []byte(""), err
	}
	var projectInfo api.Project
	err = json.Unmarshal([]byte(projectByName), &projectInfo)
	if err != nil {
		return []byte(""), err
	}
	// get the environment info from lagoon, we need the environment ID for later
	// we consume the project ID here
	environment := api.EnvironmentByName{
		Name:    environmentName,
		Project: projectInfo.ID,
	}
	environmentByName, err := e.api.GetEnvironmentByName(environment, graphql.EnvironmentByNameFragment)
	if err != nil {
		return []byte(""), err
	}
	var environmentInfo api.Environment
	err = json.Unmarshal([]byte(environmentByName), &environmentInfo)
	if err != nil {
		return []byte(""), err
	}

	returnResult, err := processEnvInfo(environmentByName)
	if err != nil {
		return []byte(""), err
	}
	return returnResult, nil
}

func processEnvInfo(projectByName []byte) ([]byte, error) {
	var environment api.Environment
	err := json.Unmarshal([]byte(projectByName), &environment)
	if err != nil {
		return []byte(""), err
	}
	environmentData := processEnvExtra(environment)
	var data []output.Data
	data = append(data, environmentData)
	dataMain := output.Table{
		Header: []string{"ID", "EnvironmentName", "EnvironmentType", "DeployType", "Created", "Route", "Routes", "MonitoringURLS", "AutoIdle", "DeployTitle", "DeployBaseRef", "DeployHeadRef"},
		Data:   data,
	}
	return json.Marshal(dataMain)
}

func processEnvExtra(environment api.Environment) []string {
	envID := returnNonEmptyString(strconv.Itoa(environment.ID))
	envName := returnNonEmptyString(string(environment.Name))
	envEnvironmentType := returnNonEmptyString(string(environment.EnvironmentType))
	envDeployType := returnNonEmptyString(string(environment.DeployType))
	envCreated := returnNonEmptyString(string(environment.Created))
	envRoute := returnNonEmptyString(string(environment.Route))
	envRoutes := returnNonEmptyString(string(environment.Routes))
	envMonitoringUrls := returnNonEmptyString(string(environment.MonitoringUrls))
	envDeployTitle := returnNonEmptyString(string(environment.DeployTitle))
	envDeployBaseRef := returnNonEmptyString(string(environment.DeployBaseRef))
	envDeployHeadRef := returnNonEmptyString(string(environment.DeployHeadRef))
	envAutoIdle := *environment.AutoIdle
	data := []string{
		fmt.Sprintf("%v", envID),
		fmt.Sprintf("%v", envName),
		fmt.Sprintf("%v", envEnvironmentType),
		fmt.Sprintf("%v", envDeployType),
		fmt.Sprintf("%v", envCreated),
		fmt.Sprintf("%v", envRoute),
		fmt.Sprintf("%v", envRoutes),
		fmt.Sprintf("%v", envMonitoringUrls),
		fmt.Sprintf("%v", envAutoIdle),
		fmt.Sprintf("%v", envDeployTitle),
		fmt.Sprintf("%v", envDeployBaseRef),
		fmt.Sprintf("%v", envDeployHeadRef),
	}
	return data
}

func returnNonEmptyString(value string) string {
	if len(value) == 0 {
		value = "-"
	}
	return value
}

// AddOrUpdateEnvironment .
func AddOrUpdateEnvironment(projectName string, environmentName string, jsonPatch string) ([]byte, error) {
	lagoonAPI, err := graphql.LagoonAPI()
	if err != nil {
		return []byte(""), err
	} // get project info from lagoon, we need the project ID for later
	project := api.Project{
		Name: projectName,
	}
	projectByName, err := lagoonAPI.GetProjectByName(project, graphql.ProjectNameID)
	if err != nil {
		return []byte(""), err
	}
	var projectInfo api.Project
	err = json.Unmarshal([]byte(projectByName), &projectInfo)
	if err != nil {
		return []byte(""), err
	}
	environment := api.AddUpdateEnvironment{}
	err = json.Unmarshal([]byte(jsonPatch), &environment.Patch)
	if err != nil {
		return []byte(""), err
	}
	environment.Name = environmentName

	projectAddResult, err := lagoonAPI.AddOrUpdateEnvironment(projectInfo.ID, environment)
	if err != nil {
		return []byte(""), err
	}
	return projectAddResult, nil
}
