package downloader

import (
	"context"
	"encoding/gob"
	"fmt"
)

var (
	ErrTaskNotFount = fmt.Errorf("task not found")
)

type (
	Downloader interface {
		// Create a task with the given URL and options overwriting the default settings, returns a task handle for future operations.
		CreateTask(ctx context.Context, url string, options map[string]interface{}) (*TaskHandle, error)
		// Info returns the status of the task with the given handle.
		Info(ctx context.Context, handle *TaskHandle) (*TaskStatus, error)
		// Cancel the task with the given handle.
		Cancel(ctx context.Context, handle *TaskHandle) error
		// SetFilesToDownload sets the files to download for the task with the given handle.
		SetFilesToDownload(ctx context.Context, handle *TaskHandle, args ...*SetFileToDownloadArgs) error
		// Test tests the connection to the downloader.
		Test(ctx context.Context) (string, error)
	}

	// TaskHandle represents a task handle for future operations
	TaskHandle struct {
		ID   string `json:"id"`
		Hash string `json:"hash"`
	}
	Status     string
	TaskStatus struct {
		FollowedBy    *TaskHandle `json:"-"` // Indicate if the task handle is changed
		SavePath      string      `json:"save_path,omitempty"`
		Name          string      `json:"name"`
		State         Status      `json:"state"`
		Total         int64       `json:"total"`
		Downloaded    int64       `json:"downloaded"`
		DownloadSpeed int64       `json:"download_speed"`
		Uploaded      int64       `json:"uploaded"`
		UploadSpeed   int64       `json:"upload_speed"`
		Hash          string      `json:"hash,omitempty"`
		Files         []TaskFile  `json:"files,omitempty"`
		Pieces        []byte      `json:"pieces,omitempty"` // Hexadecimal representation of the download progress of the peer. The highest bit corresponds to the piece at index 0.
		NumPieces     int         `json:"num_pieces,omitempty"`
		ErrorMessage  string      `json:"error_message,omitempty"`
	}

	TaskFile struct {
		Index    int     `json:"index"`
		Name     string  `json:"name"`
		Size     int64   `json:"size"`
		Progress float64 `json:"progress"`
		Selected bool    `json:"selected"`
	}

	SetFileToDownloadArgs struct {
		Index    int  `json:"index"`
		Download bool `json:"download"`
	}
)

const (
	StatusDownloading Status = "downloading"
	StatusSeeding     Status = "seeding"
	StatusCompleted   Status = "completed"
	StatusError       Status = "error"
	StatusUnknown     Status = "unknown"

	DownloaderCtxKey = "downloader"
)

func init() {
	gob.Register(TaskHandle{})
	gob.Register(TaskStatus{})
}
