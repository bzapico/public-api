package cli

import (
	"encoding/json"
	"fmt"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-application-go"
	"github.com/nalej/grpc-application-manager-go"
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/grpc-public-api-go"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"io/ioutil"
)

type Applications struct {
	Connection
	Credentials
}

func NewApplications(address string, port int, insecure bool, caCertPath string) *Applications {
	return &Applications{
		Connection:  *NewConnection(address, port, insecure, caCertPath),
		Credentials: *NewEmptyCredentials(DefaultPath),
	}
}

func (a *Applications) load() {
	err := a.LoadCredentials()
	if err != nil {
		log.Fatal().Str("trace", err.DebugReport()).Msg("cannot load credentials, try login first")
	}
}

func (a *Applications) getClient() (grpc_public_api_go.ApplicationsClient, *grpc.ClientConn) {
	conn, err := a.GetConnection()
	if err != nil {
		log.Fatal().Str("trace", err.DebugReport()).Msg("cannot create the connection with the Nalej platform")
	}
	appsClient := grpc_public_api_go.NewApplicationsClient(conn)
	return appsClient, conn
}

func (a *Applications) createAddDescriptorRequest(organizationID string, descriptorPath string) (*grpc_application_go.AddAppDescriptorRequest, derrors.Error) {

	descPath := GetPath(descriptorPath)
	content, err := ioutil.ReadFile(descPath)
	if err != nil {
		return nil, derrors.AsError(err, "cannot read descriptor")
	}

	addDescriptorRequest := &grpc_application_go.AddAppDescriptorRequest{}
	err = json.Unmarshal(content, &addDescriptorRequest)
	if err != nil {
		return nil, derrors.AsError(err, "cannot unmarshal structure")
	}

	addDescriptorRequest.OrganizationId = organizationID

	return addDescriptorRequest, nil
}

func (a *Applications) ShowDescriptorHelp(exampleName string, storageType string) {
	// convert string sType to StorageType
	sType := a.GetStorageType(storageType)
	if exampleName == "simple" {
		a.ShowDescriptorExample(sType)
	} else if exampleName == "complex" {
		a.ShowComplexDescriptorExample(sType)
	} else if exampleName == "multireplica" {
		a.ShowMultiReplicaDescriptorExample(sType)
	}else {
		fmt.Println("Supported examples: simple, complex, multireplica")
	}
}

func (a *Applications) ShowDescriptorExample(sType grpc_application_go.StorageType) {
	toAdd := a.getBasicDescriptor(sType)
	err := a.PrintResult(toAdd)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load sample application descriptor")
	}
}

func (a *Applications) ShowComplexDescriptorExample(sType grpc_application_go.StorageType) {
	toAdd := a.getComplexDescriptor(sType)
	err := a.PrintResult(toAdd)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load sample application descriptor")
	}
}

func (a *Applications) ShowMultiReplicaDescriptorExample(sType grpc_application_go.StorageType) {
	toAdd := a.getMultiReplicaDescriptor(sType)
	err := a.PrintResult(toAdd)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load sample application descriptor")
	}
}

func (a *Applications) GetStorageType(sType string) grpc_application_go.StorageType {
	switch sType {
	case "ephemeral":
		return grpc_application_go.StorageType_EPHEMERAL
	case "local":
		return grpc_application_go.StorageType_CLUSTER_LOCAL
	case "replica":
		return grpc_application_go.StorageType_CLUSTER_REPLICA
	case "cloud":
		return grpc_application_go.StorageType_CLOUD_PERSISTENT
	}
	return grpc_application_go.StorageType_EPHEMERAL
}

func (a *Applications) AddDescriptor(organizationID string, descriptorPath string) {

	if organizationID == "" {
		log.Fatal().Msg("organizationID cannot be empty")
	}

	a.load()
	ctx, cancel := a.GetContext()
	client, conn := a.getClient()
	defer conn.Close()
	defer cancel()

	addDescriptorRequest, aErr := a.createAddDescriptorRequest(organizationID, descriptorPath)
	if aErr != nil {
		log.Fatal().Str("trace", aErr.DebugReport()).Msg("cannot load application descriptor")
	}
	added, err := client.AddAppDescriptor(ctx, addDescriptorRequest)
	a.PrintResultOrError(added, err, "cannot add a new application descriptor")
}

func (a *Applications) GetDescriptor(organizationID string, descriptorID string) {
	if organizationID == "" {
		log.Fatal().Msg("organizationID cannot be empty")
	}
	if descriptorID == "" {
		log.Fatal().Msg("descriptorID cannot be empty")
	}
	a.load()
	ctx, cancel := a.GetContext()
	client, conn := a.getClient()
	defer conn.Close()
	defer cancel()

	appDescriptorID := &grpc_application_go.AppDescriptorId{
		OrganizationId:  organizationID,
		AppDescriptorId: descriptorID,
	}
	descriptor, err := client.GetAppDescriptor(ctx, appDescriptorID)
	a.PrintResultOrError(descriptor, err, "cannot obtain descriptor")
}

func (a *Applications) DeleteDescriptor(organizationID string, descriptorID string) {
	if organizationID == "" {
		log.Fatal().Msg("organizationID cannot be empty")
	}
	if descriptorID == "" {
		log.Fatal().Msg("descriptorID cannot be empty")
	}

	a.load()
	ctx, cancel := a.GetContext()
	client, conn := a.getClient()
	defer conn.Close()
	defer cancel()

	appDescriptorID := &grpc_application_go.AppDescriptorId{
		OrganizationId:  organizationID,
		AppDescriptorId: descriptorID,
	}
	_, err := client.DeleteAppDescriptor(ctx, appDescriptorID)
	a.PrintSuccessOrError(err, "cannot delete given descriptor", "application descriptor removed")
}

func (a *Applications) ListDescriptors(organizationID string) {
	if organizationID == "" {
		log.Fatal().Msg("organizationID cannot be empty")
	}
	a.load()
	ctx, cancel := a.GetContext()
	client, conn := a.getClient()
	defer conn.Close()
	defer cancel()

	orgID := &grpc_organization_go.OrganizationId{
		OrganizationId: organizationID,
	}
	list, err := client.ListAppDescriptors(ctx, orgID)
	a.PrintResultOrError(list, err, "cannot obtain descriptor list")
}

func (a *Applications) ModifyAppDescriptorLabels(organizationID string, descriptorID string, add bool, rawLabels string){
	if organizationID == "" {
		log.Fatal().Msg("organizationID cannot be empty")
	}
	if descriptorID == "" {
		log.Fatal().Msg("descriptorID cannot be empty")
	}
	if rawLabels == "" {
		log.Fatal().Msg("labels cannot be empty")
	}
	a.load()
	ctx, cancel := a.GetContext()
	client, conn := a.getClient()
	defer conn.Close()
	defer cancel()
	updateRequest := &grpc_application_go.UpdateAppDescriptorRequest{
		OrganizationId: organizationID,
		AppDescriptorId:      descriptorID,
		AddLabels: add,
		RemoveLabels: !add,
		Labels: GetLabels(rawLabels),
	}
	updated, err := client.UpdateAppDescriptor(ctx, updateRequest)
	a.PrintResultOrError(updated, err, "cannot update application descriptor labels")
}


func (a *Applications) Deploy(organizationID string, appDescriptorID string, name string, description string) {
	if organizationID == "" {
		log.Fatal().Msg("organizationID cannot be empty")
	}
	if appDescriptorID == "" {
		log.Fatal().Msg("descriptorID cannot be empty")
	}
	a.load()
	ctx, cancel := a.GetContext()
	client, conn := a.getClient()
	defer conn.Close()
	defer cancel()

	deployRequest := &grpc_application_manager_go.DeployRequest{
		OrganizationId:  organizationID,
		AppDescriptorId: appDescriptorID,
		Name:            name,
		Description:     description,
	}
	deployed, err := client.Deploy(ctx, deployRequest)
	a.PrintResultOrError(deployed, err, "cannot deploy application")
}

func (a *Applications) Undeploy(organizationID string, appInstanceID string) {
	if organizationID == "" {
		log.Fatal().Msg("organizationID cannot be empty")
	}
	if appInstanceID == "" {
		log.Fatal().Msg("instanceID cannot be empty")
	}
	a.load()
	ctx, cancel := a.GetContext()
	client, conn := a.getClient()
	defer conn.Close()
	defer cancel()

	undeployRequest := &grpc_application_go.AppInstanceId{
		OrganizationId: organizationID,
		AppInstanceId:  appInstanceID,
	}
	_, err := client.Undeploy(ctx, undeployRequest)
	a.PrintSuccessOrError(err, "cannot undeploy application", "application instance undeployed")
}

func (a *Applications) ListInstances(organizationID string) {
	if organizationID == "" {
		log.Fatal().Msg("organizationID cannot be empty")
	}

	a.load()
	ctx, cancel := a.GetContext()
	client, conn := a.getClient()
	defer conn.Close()
	defer cancel()

	orgID := &grpc_organization_go.OrganizationId{
		OrganizationId: organizationID,
	}
	list, err := client.ListAppInstances(ctx, orgID)
	a.PrintResultOrError(list, err, "cannot list application instances")
}

func (a *Applications) GetInstance(organizationID string, appInstanceID string) {

	if organizationID == "" {
		log.Fatal().Msg("organizationID cannot be empty")
	}
	if appInstanceID == "" {
		log.Fatal().Msg("instanceID cannot be empty")
	}
	a.load()
	ctx, cancel := a.GetContext()
	client, conn := a.getClient()
	defer conn.Close()
	defer cancel()

	instID := &grpc_application_go.AppInstanceId{
		OrganizationId: organizationID,
		AppInstanceId:  appInstanceID,
	}
	inst, err := client.GetAppInstance(ctx, instID)
	a.PrintResultOrError(inst, err, "cannot obtain application instance information")
}
