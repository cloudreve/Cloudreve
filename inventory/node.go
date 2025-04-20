package inventory

import (
	"context"

	"github.com/cloudreve/Cloudreve/v4/ent"
	"github.com/cloudreve/Cloudreve/v4/ent/node"
)

type (
	LoadNodeStoragePolicy struct{}
	NodeClient            interface {
		TxOperator
		// ListActiveNodes returns the active nodes.
		ListActiveNodes(ctx context.Context, subset []int) ([]*ent.Node, error)
		// ListNodes returns the nodes with pagination.
		ListNodes(ctx context.Context, args *ListNodeParameters) (*ListNodeResult, error)
		// GetNodeById returns the node by id.
		GetNodeById(ctx context.Context, id int) (*ent.Node, error)
		// GetNodeByIds returns the nodes by ids.
		GetNodeByIds(ctx context.Context, ids []int) ([]*ent.Node, error)
		// Upsert upserts a node.
		Upsert(ctx context.Context, n *ent.Node) (*ent.Node, error)
		// Delete deletes a node.
		Delete(ctx context.Context, id int) error
	}
	ListNodeParameters struct {
		*PaginationArgs
		Status node.Status
	}
	ListNodeResult struct {
		*PaginationResults
		Nodes []*ent.Node
	}
)

func NewNodeClient(client *ent.Client) NodeClient {
	return &nodeClient{
		client: client,
	}
}

type nodeClient struct {
	client *ent.Client
}

func (c *nodeClient) SetClient(newClient *ent.Client) TxOperator {
	return &nodeClient{client: newClient}
}

func (c *nodeClient) GetClient() *ent.Client {
	return c.client
}

func (c *nodeClient) ListActiveNodes(ctx context.Context, subset []int) ([]*ent.Node, error) {
	stm := c.client.Node.Query().Where(node.StatusEQ(node.StatusActive))
	if len(subset) > 0 {
		stm = stm.Where(node.IDIn(subset...))
	}
	return stm.All(ctx)
}

func (c *nodeClient) GetNodeByIds(ctx context.Context, ids []int) ([]*ent.Node, error) {
	return withNodeEagerLoading(ctx, c.client.Node.Query().Where(node.IDIn(ids...))).All(ctx)
}

func (c *nodeClient) GetNodeById(ctx context.Context, id int) (*ent.Node, error) {
	return withNodeEagerLoading(ctx, c.client.Node.Query().Where(node.IDEQ(id))).First(ctx)
}

func (c *nodeClient) Delete(ctx context.Context, id int) error {
	return c.client.Node.DeleteOneID(id).Exec(ctx)
}

func (c *nodeClient) ListNodes(ctx context.Context, args *ListNodeParameters) (*ListNodeResult, error) {
	query := c.client.Node.Query()
	if string(args.Status) != "" {
		query = query.Where(node.StatusEQ(args.Status))
	}
	query.Order(getNodeOrderOption(args)...)

	// Count total items
	total, err := query.Clone().
		Count(ctx)
	if err != nil {
		return nil, err
	}

	nodes, err := withNodeEagerLoading(ctx, query).Limit(args.PageSize).Offset(args.Page * args.PageSize).All(ctx)
	if err != nil {
		return nil, err
	}

	return &ListNodeResult{
		PaginationResults: &PaginationResults{
			TotalItems: total,
			Page:       args.Page,
			PageSize:   args.PageSize,
		},
		Nodes: nodes,
	}, nil
}

func (c *nodeClient) Upsert(ctx context.Context, n *ent.Node) (*ent.Node, error) {
	if n.ID == 0 {
		return c.client.Node.Create().
			SetName(n.Name).
			SetServer(n.Server).
			SetSlaveKey(n.SlaveKey).
			SetStatus(n.Status).
			SetType(node.TypeSlave).
			SetSettings(n.Settings).
			SetCapabilities(n.Capabilities).
			SetWeight(n.Weight).
			Save(ctx)
	}

	res, err := c.client.Node.UpdateOne(n).
		SetName(n.Name).
		SetServer(n.Server).
		SetSlaveKey(n.SlaveKey).
		SetStatus(n.Status).
		SetSettings(n.Settings).
		SetCapabilities(n.Capabilities).
		SetWeight(n.Weight).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func getNodeOrderOption(args *ListNodeParameters) []node.OrderOption {
	orderTerm := getOrderTerm(args.Order)
	switch args.OrderBy {
	case node.FieldName:
		return []node.OrderOption{node.ByName(orderTerm), node.ByID(orderTerm)}
	case node.FieldWeight:
		return []node.OrderOption{node.ByWeight(orderTerm), node.ByID(orderTerm)}
	case node.FieldUpdatedAt:
		return []node.OrderOption{node.ByUpdatedAt(orderTerm), node.ByID(orderTerm)}
	default:
		return []node.OrderOption{node.ByID(orderTerm)}
	}
}

func withNodeEagerLoading(ctx context.Context, query *ent.NodeQuery) *ent.NodeQuery {
	if _, ok := ctx.Value(LoadNodeStoragePolicy{}).(bool); ok {
		query = query.WithStoragePolicy(func(gq *ent.StoragePolicyQuery) {
			withStoragePolicyEagerLoading(ctx, gq)
		})
	}

	return query
}
