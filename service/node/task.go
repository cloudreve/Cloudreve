package node

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/cloudreve/Cloudreve/v4/application/dependency"
	"github.com/cloudreve/Cloudreve/v4/ent/task"
	"github.com/cloudreve/Cloudreve/v4/inventory/types"
	"github.com/cloudreve/Cloudreve/v4/pkg/cluster"
	"github.com/cloudreve/Cloudreve/v4/pkg/cluster/routes"
	"github.com/cloudreve/Cloudreve/v4/pkg/filemanager/workflows"
	"github.com/cloudreve/Cloudreve/v4/pkg/logging"
	"github.com/cloudreve/Cloudreve/v4/pkg/queue"
	"github.com/cloudreve/Cloudreve/v4/pkg/serializer"
	"github.com/gin-gonic/gin"
)

type (
	CreateSlaveTaskParamCtx struct{}
)

func CreateTaskInSlave(s *cluster.CreateSlaveTask, c *gin.Context) (int, error) {
	dep := dependency.FromContext(c)
	registry := dep.TaskRegistry()

	props, err := slaveTaskPropsFromContext(c)
	if err != nil {
		return 0, serializer.NewError(serializer.CodeParamErr, "failed to get master props from header", err)
	}

	var t queue.Task
	switch s.Type {
	case queue.SlaveUploadTaskType:
		t = workflows.NewSlaveUploadTask(c, props, registry.NextID(), s.State)
	case queue.SlaveCreateArchiveTaskType:
		t = workflows.NewSlaveCreateArchiveTask(c, props, registry.NextID(), s.State)
	case queue.SlaveExtractArchiveType:
		t = workflows.NewSlaveExtractArchiveTask(c, props, registry.NextID(), s.State)
	default:
		return 0, serializer.NewError(serializer.CodeParamErr, "type not supported", nil)
	}

	if err := dep.SlaveQueue(c).QueueTask(c, t); err != nil {
		return 0, serializer.NewError(serializer.CodeInternalSetting, "failed to queue task", err)
	}

	registry.Set(t.ID(), t)
	return t.ID(), nil
}

type (
	GetSlaveTaskParamCtx struct{}
	GetSlaveTaskService  struct {
		ID int `uri:"id" binding:"required"`
	}
)

func (s *GetSlaveTaskService) Get(c *gin.Context) (*cluster.SlaveTaskSummary, error) {
	dep := dependency.FromContext(c)
	registry := dep.TaskRegistry()

	t, ok := registry.Get(s.ID)
	if !ok {
		return nil, serializer.NewError(serializer.CodeNotFound, "task not found", nil)
	}
	status := t.Status()
	_, clearOnComplete := c.GetQuery(routes.SlaveClearTaskRegistryQuery)
	if clearOnComplete && status == task.StatusCompleted ||
		status == task.StatusError ||
		status == task.StatusCanceled {
		registry.Delete(s.ID)
	}

	res := &cluster.SlaveTaskSummary{
		Status:       status,
		PrivateState: t.State(),
		Progress:     t.Progress(c),
	}
	err := t.Error()
	if err != nil {
		res.Error = err.Error()
	}

	return res, nil
}

func slaveTaskPropsFromContext(ctx context.Context) (*types.SlaveTaskProps, error) {
	nodeIdStr, ok := ctx.Value(cluster.SlaveNodeIDCtx{}).(string)
	if !ok {
		return nil, fmt.Errorf("failed to get node ID from context")
	}

	nodeId, err := strconv.Atoi(nodeIdStr)
	if err != nil {
		return nil, fmt.Errorf("failed to convert node ID to int: %w", err)
	}

	masterSiteUrl := cluster.MasterSiteUrlFromContext(ctx)
	if masterSiteUrl == "" {
		return nil, fmt.Errorf("failed to get master site URL from context")
	}

	masterSiteVersion, ok := ctx.Value(cluster.MasterSiteVersionCtx{}).(string)
	if !ok {
		return nil, fmt.Errorf("failed to get master site version from context")
	}

	masterSiteId, ok := ctx.Value(cluster.MasterSiteIDCtx{}).(string)
	if !ok {
		return nil, fmt.Errorf("failed to convert master site ID to int: %w", err)
	}

	props := &types.SlaveTaskProps{
		NodeID:            nodeId,
		MasterSiteID:      masterSiteId,
		MasterSiteURl:     masterSiteUrl,
		MasterSiteVersion: masterSiteVersion,
	}

	return props, nil
}

type (
	FolderCleanupParamCtx struct{}
)

func Cleanup(args *cluster.FolderCleanup, c *gin.Context) error {
	l := logging.FromContext(c)
	ae := serializer.NewAggregateError()
	for _, p := range args.Path {
		l.Info("Cleaning up folder %q", p)
		if err := os.RemoveAll(p); err != nil {
			l.Warning("Failed to clean up folder %q: %s", p, err)
			ae.Add(p, err)
		}
	}

	return ae.Aggregate()
}

type (
	CreateSlaveDownloadTaskParamCtx  struct{}
	GetSlaveDownloadTaskParamCtx     struct{}
	CancelSlaveDownloadTaskParamCtx  struct{}
	SelectSlaveDownloadFilesParamCtx struct{}
	TestSlaveDownloadParamCtx        struct{}
)
