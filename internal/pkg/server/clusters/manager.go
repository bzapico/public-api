/*
 * Copyright (C)  2018 Nalej - All Rights Reserved
 */

package clusters

import (
	"context"
	"github.com/nalej/grpc-common-go"
	"github.com/nalej/grpc-infrastructure-go"
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/grpc-public-api-go"
	"github.com/nalej/public-api/internal/pkg/entities"
)

// Manager structure with the required clients for cluster operations.
type Manager struct {
	clustClient grpc_infrastructure_go.ClustersClient
	nodeClient grpc_infrastructure_go.NodesClient
}

// NewManager creates a Manager using a set of clients.
func NewManager(clustClient grpc_infrastructure_go.ClustersClient,
nodeClient grpc_infrastructure_go.NodesClient) Manager {
	return Manager{
		clustClient: clustClient, nodeClient:nodeClient,
	}
}

// clusterNodeStats determines the number of total and running nodes in a cluster.
func (m * Manager) clusterNodesStats(organizationID string, clusterID string) (int64, int64, error){
	runningNodes := 0

	cID := &grpc_infrastructure_go.ClusterId{
		OrganizationId:       organizationID,
		ClusterId:            clusterID,
	}
	clusterNodes, err := m.nodeClient.ListNodes(context.Background(), cID)
	if err != nil{
		return 0, 0, err
	}
	for _, n := range clusterNodes.Nodes{
		if n.Status == grpc_infrastructure_go.InfraStatus_RUNNING {
			runningNodes++
		}
	}
	return int64(len(clusterNodes.Nodes)), int64(runningNodes), nil
}

// List all the clusters in an organization.
func (m *Manager) List(organizationID *grpc_organization_go.OrganizationId) (*grpc_public_api_go.ClusterList, error) {
	list, err := m.clustClient.ListClusters(context.Background(), organizationID)
	if err != nil{
		return nil, err
	}
	clusters := make([]* grpc_public_api_go.Cluster, 0)
	for _, c := range list.Clusters{

		totalNodes, runningNodes, err := m.clusterNodesStats(organizationID.OrganizationId, c.ClusterId)
		if err != nil{
			return nil, err
		}
		toAdd := &grpc_public_api_go.Cluster{
			OrganizationId:       organizationID.OrganizationId,
			ClusterId:            c.ClusterId,
			Name:                 c.Name,
			Description:          c.Description,
			ClusterType:          c.ClusterType,
			Multitenant:          c.Multitenant,
			Status:               c.Status,
			Labels:               c.Labels,
			TotalNodes:           totalNodes,
			RunningNodes:         runningNodes,
		}
		clusters = append(clusters, toAdd)
	}
	return &grpc_public_api_go.ClusterList{
		Clusters:             clusters,
	}, nil
}

// Update the cluster information.
func (m * Manager) Update(updateClusterRequest *grpc_public_api_go.UpdateClusterRequest) (*grpc_common_go.Success, error) {
	toSend := entities.ToInfraClusterUpdate(*updateClusterRequest)
	_, err := m.clustClient.UpdateCluster(context.Background(), toSend)
	if err != nil{
		return nil, err
	}
	return &grpc_common_go.Success{}, nil
}