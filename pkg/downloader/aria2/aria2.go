package aria2

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cloudreve/Cloudreve/v4/inventory/types"
	"github.com/cloudreve/Cloudreve/v4/pkg/downloader"
	"github.com/cloudreve/Cloudreve/v4/pkg/downloader/aria2/rpc"
	"github.com/cloudreve/Cloudreve/v4/pkg/logging"
	"github.com/cloudreve/Cloudreve/v4/pkg/setting"
	"github.com/cloudreve/Cloudreve/v4/pkg/util"
	"github.com/gofrs/uuid"
	"github.com/samber/lo"
)

const (
	Aria2TempFolder        = "aria2"
	deleteTempFileDuration = 120 * time.Second
)

type aria2Client struct {
	l        logging.Logger
	settings setting.Provider

	options *types.Aria2Setting
	timeout time.Duration
	caller  rpc.Client
}

func New(l logging.Logger, settings setting.Provider, options *types.Aria2Setting) downloader.Downloader {
	rpcServer := options.Server
	rpcUrl, err := url.Parse(options.Server)
	if err == nil {
		// add /jsonrpc to the url if not present
		rpcUrl.Path = "/jsonrpc"
		rpcServer = rpcUrl.String()
	}

	options.Server = rpcServer
	return &aria2Client{
		l:        l,
		settings: settings,
		options:  options,
		timeout:  time.Duration(10) * time.Second,
	}
}

func (a *aria2Client) CreateTask(ctx context.Context, url string, options map[string]interface{}) (*downloader.TaskHandle, error) {
	caller := a.caller
	if caller == nil {
		var err error
		caller, err = rpc.New(ctx, a.options.Server, a.options.Token, a.timeout, nil)
		if err != nil {
			return nil, fmt.Errorf("cannot create rpc client: %w", err)
		}
	}

	path := a.tempPath(ctx)
	a.l.Info("Creating aria2 task with url %q saving to %q...", url, path)

	// Create the download task options
	downloadOptions := map[string]interface{}{}
	for k, v := range a.options.Options {
		downloadOptions[k] = v
	}
	for k, v := range options {
		downloadOptions[k] = v
	}
	downloadOptions["dir"] = path
	downloadOptions["follow-torrent"] = "mem"

	gid, err := caller.AddURI(url, downloadOptions)
	if err != nil || gid == "" {
		return nil, err
	}

	return &downloader.TaskHandle{
		ID: gid,
	}, nil
}

func (a *aria2Client) Info(ctx context.Context, handle *downloader.TaskHandle) (*downloader.TaskStatus, error) {
	caller := a.caller
	if caller == nil {
		var err error
		caller, err = rpc.New(ctx, a.options.Server, a.options.Token, a.timeout, nil)
		if err != nil {
			return nil, fmt.Errorf("cannot create rpc client: %w", err)
		}
	}

	status, err := caller.TellStatus(handle.ID)
	if err != nil {
		return nil, fmt.Errorf("aria2 rpc error: %w", err)
	}

	state := downloader.StatusDownloading
	switch status.Status {
	case "active":
		if status.BitTorrent.Mode != "" && status.CompletedLength == status.TotalLength {
			state = downloader.StatusSeeding
		} else {
			state = downloader.StatusDownloading
		}
	case "waiting", "paused":
		state = downloader.StatusDownloading
	case "complete":
		state = downloader.StatusCompleted
	case "error":
		state = downloader.StatusError
	case "cancelled", "removed":
		a.l.Debug("Task %q is cancelled", handle.ID)
		return nil, fmt.Errorf("Task canceled: %w", downloader.ErrTaskNotFount)
	}

	totalLength, _ := strconv.ParseInt(status.TotalLength, 10, 64)
	downloaded, _ := strconv.ParseInt(status.CompletedLength, 10, 64)
	downloadSpeed, _ := strconv.ParseInt(status.DownloadSpeed, 10, 64)
	uploaded, _ := strconv.ParseInt(status.UploadLength, 10, 64)
	uploadSpeed, _ := strconv.ParseInt(status.UploadSpeed, 10, 64)
	numPieces, _ := strconv.Atoi(status.NumPieces)
	savePath := filepath.ToSlash(status.Dir)

	res := &downloader.TaskStatus{
		State:         state,
		Name:          status.BitTorrent.Info.Name,
		Total:         totalLength,
		Downloaded:    downloaded,
		DownloadSpeed: downloadSpeed,
		Uploaded:      uploaded,
		UploadSpeed:   uploadSpeed,
		SavePath:      savePath,
		NumPieces:     numPieces,
		ErrorMessage:  status.ErrorMessage,
		Hash:          status.InfoHash,
		Files: lo.Map(status.Files, func(item rpc.FileInfo, index int) downloader.TaskFile {
			index, _ = strconv.Atoi(item.Index)
			size, _ := strconv.ParseInt(item.Length, 10, 64)
			completed, _ := strconv.ParseInt(item.CompletedLength, 10, 64)
			relPath := strings.TrimPrefix(filepath.ToSlash(item.Path), savePath)
			// Remove first letter if any
			if len(relPath) > 0 {
				relPath = relPath[1:]
			}
			progress := 0.0
			if size > 0 {
				progress = float64(completed) / float64(size)
			}
			return downloader.TaskFile{
				Index:    index,
				Name:     relPath,
				Size:     size,
				Progress: progress,
				Selected: item.Selected == "true",
			}
		}),
	}

	if len(status.FollowedBy) > 0 {
		res.FollowedBy = &downloader.TaskHandle{
			ID: status.FollowedBy[0],
		}
	}

	if len(status.Files) == 1 && res.Name == "" {
		res.Name = path.Base(filepath.ToSlash(status.Files[0].Path))
	}

	if status.BitField != "" {
		res.Pieces = make([]byte, len(status.BitField)/2)
		// Convert hex string to bytes
		for i := 0; i < len(status.BitField); i += 2 {
			b, _ := strconv.ParseInt(status.BitField[i:i+2], 16, 8)
			res.Pieces[i/2] = byte(b)
		}
	}

	return res, nil
}

func (a *aria2Client) Cancel(ctx context.Context, handle *downloader.TaskHandle) error {
	caller := a.caller
	if caller == nil {
		var err error
		caller, err = rpc.New(ctx, a.options.Server, a.options.Token, a.timeout, nil)
		if err != nil {
			return fmt.Errorf("cannot create rpc client: %w", err)
		}
	}

	status, err := a.Info(ctx, handle)
	if err != nil {
		return fmt.Errorf("cannot get task: %w", err)
	}

	// Delay to delete temp download folder to avoid being locked by aria2
	defer func() {
		go func(parent string, l logging.Logger) {
			time.Sleep(deleteTempFileDuration)
			err := os.RemoveAll(parent)
			if err != nil {
				l.Warning("Failed to delete temp download folder: %q: %s", parent, err)
			}
		}(status.SavePath, a.l)
	}()

	if _, err := caller.Remove(handle.ID); err != nil {
		return fmt.Errorf("aria2 rpc error: %w", err)
	}

	return nil
}

func (a *aria2Client) SetFilesToDownload(ctx context.Context, handle *downloader.TaskHandle, args ...*downloader.SetFileToDownloadArgs) error {
	caller := a.caller
	if caller == nil {
		var err error
		caller, err = rpc.New(ctx, a.options.Server, a.options.Token, a.timeout, nil)
		if err != nil {
			return fmt.Errorf("cannot create rpc client: %w", err)
		}
	}

	status, err := a.Info(ctx, handle)
	if err != nil {
		return fmt.Errorf("cannot get task: %w", err)
	}

	selected := lo.SliceToMap(
		lo.Filter(status.Files, func(item downloader.TaskFile, _ int) bool {
			return item.Selected
		}),
		func(item downloader.TaskFile) (int, bool) {
			return item.Index, true
		},
	)
	for _, arg := range args {
		if !arg.Download {
			delete(selected, arg.Index)
		}
	}

	_, err = caller.ChangeOption(handle.ID, map[string]interface{}{"select-file": strings.Join(lo.MapToSlice(selected, func(key int, value bool) string {
		return strconv.Itoa(key)
	}), ",")})
	return err
}

func (a *aria2Client) Test(ctx context.Context) (string, error) {
	caller := a.caller
	if caller == nil {
		var err error
		caller, err = rpc.New(ctx, a.options.Server, a.options.Token, a.timeout, nil)
		if err != nil {
			return "", fmt.Errorf("cannot create rpc client: %w", err)
		}
	}

	version, err := caller.GetVersion()
	if err != nil {
		return "", fmt.Errorf("cannot call aria2: %w", err)
	}

	return version.Version, nil
}

func (a *aria2Client) tempPath(ctx context.Context) string {
	guid, _ := uuid.NewV4()

	// Generate a unique path for the task
	base := util.RelativePath(a.options.TempPath)
	if a.options.TempPath == "" {
		base = util.DataPath(a.settings.TempPath(ctx))
	}
	path := filepath.Join(
		base,
		Aria2TempFolder,
		guid.String(),
	)
	return path
}
