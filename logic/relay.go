package logic

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gravitl/netmaker/database"
	"github.com/gravitl/netmaker/logger"
	"github.com/gravitl/netmaker/models"
)

// CreateRelay - creates a relay
func CreateRelay(relay models.RelayRequest) ([]models.LegacyNode, models.LegacyNode, error) {
	var returnnodes []models.LegacyNode

	node, err := GetNodeByID(relay.NodeID)
	if err != nil {
		return returnnodes, models.LegacyNode{}, err
	}
	if node.OS != "linux" {
		return returnnodes, models.LegacyNode{}, fmt.Errorf("only linux machines can be relay nodes")
	}
	err = ValidateRelay(relay)
	if err != nil {
		return returnnodes, models.LegacyNode{}, err
	}
	node.IsRelay = "yes"
	node.RelayAddrs = relay.RelayAddrs

	node.SetLastModified()
	nodeData, err := json.Marshal(&node)
	if err != nil {
		return returnnodes, node, err
	}
	if err = database.Insert(node.ID, string(nodeData), database.NODES_TABLE_NAME); err != nil {
		return returnnodes, models.LegacyNode{}, err
	}
	returnnodes, err = SetRelayedNodes(true, node.Network, node.RelayAddrs)
	if err != nil {
		return returnnodes, node, err
	}
	return returnnodes, node, nil
}

// SetRelayedNodes- set relayed nodes
func SetRelayedNodes(setRelayed bool, networkName string, addrs []string) ([]models.LegacyNode, error) {
	var returnnodes []models.LegacyNode
	networkNodes, err := GetNetworkNodes(networkName)
	if err != nil {
		return returnnodes, err
	}
	for _, node := range networkNodes {
		if node.IsServer != "yes" {
			for _, addr := range addrs {
				if addr == node.Address || addr == node.Address6 {
					if setRelayed {
						node.IsRelayed = "yes"
					} else {
						node.IsRelayed = "no"
					}
					data, err := json.Marshal(&node)
					if err != nil {
						return returnnodes, err
					}
					database.Insert(node.ID, string(data), database.NODES_TABLE_NAME)
					returnnodes = append(returnnodes, node)
				}
			}
		}
	}
	return returnnodes, nil
}
func GetRelayedNodes(relayNode *models.LegacyNode) ([]models.LegacyNode, error) {
	var returnnodes []models.LegacyNode
	networkNodes, err := GetNetworkNodes(relayNode.Network)
	if err != nil {
		return returnnodes, err
	}
	for _, node := range networkNodes {
		if node.IsServer != "yes" {
			for _, addr := range relayNode.RelayAddrs {
				if addr == node.Address || addr == node.Address6 {
					returnnodes = append(returnnodes, node)
				}
			}
		}
	}
	return returnnodes, nil
}

// ValidateRelay - checks if relay is valid
func ValidateRelay(relay models.RelayRequest) error {
	var err error
	//isIp := functions.IsIpCIDR(gateway.RangeString)
	empty := len(relay.RelayAddrs) == 0
	if empty {
		err = errors.New("IP Ranges Cannot Be Empty")
	}
	return err
}

// UpdateRelay - updates a relay
func UpdateRelay(network string, oldAddrs []string, newAddrs []string) []models.LegacyNode {
	var returnnodes []models.LegacyNode
	time.Sleep(time.Second / 4)
	_, err := SetRelayedNodes(false, network, oldAddrs)
	if err != nil {
		logger.Log(1, err.Error())
	}
	returnnodes, err = SetRelayedNodes(true, network, newAddrs)
	if err != nil {
		logger.Log(1, err.Error())
	}
	return returnnodes
}

// DeleteRelay - deletes a relay
func DeleteRelay(network, nodeid string) ([]models.LegacyNode, models.LegacyNode, error) {
	var returnnodes []models.LegacyNode
	node, err := GetNodeByID(nodeid)
	if err != nil {
		return returnnodes, models.LegacyNode{}, err
	}
	returnnodes, err = SetRelayedNodes(false, node.Network, node.RelayAddrs)
	if err != nil {
		return returnnodes, node, err
	}

	node.IsRelay = "no"
	node.RelayAddrs = []string{}
	node.SetLastModified()

	data, err := json.Marshal(&node)
	if err != nil {
		return returnnodes, models.LegacyNode{}, err
	}
	if err = database.Insert(nodeid, string(data), database.NODES_TABLE_NAME); err != nil {
		return returnnodes, models.LegacyNode{}, err
	}
	return returnnodes, node, nil
}
