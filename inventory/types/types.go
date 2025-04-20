package types

import (
	"time"
)

// UserSetting 用户其他配置
type (
	UserSetting struct {
		ProfileOff          bool        `json:"profile_off,omitempty"`
		PreferredTheme      string      `json:"preferred_theme,omitempty"`
		VersionRetention    bool        `json:"version_retention,omitempty"`
		VersionRetentionExt []string    `json:"version_retention_ext,omitempty"`
		VersionRetentionMax int         `json:"version_retention_max,omitempty"`
		Pined               []PinedFile `json:"pined,omitempty"`
		Language            string      `json:"email_language,omitempty"`
	}

	PinedFile struct {
		Uri  string `json:"uri"`
		Name string `json:"name,omitempty"`
	}

	// GroupSetting 用户组其他配置
	GroupSetting struct {
		CompressSize          int64                  `json:"compress_size,omitempty"` // 可压缩大小
		DecompressSize        int64                  `json:"decompress_size,omitempty"`
		RemoteDownloadOptions map[string]interface{} `json:"remote_download_options,omitempty"` // 离线下载用户组配置
		SourceBatchSize       int                    `json:"source_batch,omitempty"`
		Aria2BatchSize        int                    `json:"aria2_batch,omitempty"`
		MaxWalkedFiles        int                    `json:"max_walked_files,omitempty"`
		TrashRetention        int                    `json:"trash_retention,omitempty"`
		RedirectedSource      bool                   `json:"redirected_source,omitempty"`
	}

	// PolicySetting 非公有的存储策略属性
	PolicySetting struct {
		// Upyun访问Token
		Token string `json:"token"`
		// 允许的文件扩展名
		FileType []string `json:"file_type"`
		// OauthRedirect Oauth 重定向地址
		OauthRedirect string `json:"od_redirect,omitempty"`
		// CustomProxy whether to use custom-proxy to get file content
		CustomProxy bool `json:"custom_proxy,omitempty"`
		// ProxyServer 反代地址
		ProxyServer string `json:"proxy_server,omitempty"`
		// InternalProxy whether to use Cloudreve internal proxy to get file content
		InternalProxy bool `json:"internal_proxy,omitempty"`
		// OdDriver OneDrive 驱动器定位符
		OdDriver string `json:"od_driver,omitempty"`
		// Region 区域代码
		Region string `json:"region,omitempty"`
		// ServerSideEndpoint 服务端请求使用的 Endpoint，为空时使用 Policy.Server 字段
		ServerSideEndpoint string `json:"server_side_endpoint,omitempty"`
		// 分片上传的分片大小
		ChunkSize int64 `json:"chunk_size,omitempty"`
		// 每秒对存储端的 API 请求上限
		TPSLimit float64 `json:"tps_limit,omitempty"`
		// 每秒 API 请求爆发上限
		TPSLimitBurst int `json:"tps_limit_burst,omitempty"`
		// Set this to `true` to force the request to use path-style addressing,
		// i.e., `http://s3.amazonaws.com/BUCKET/KEY `
		S3ForcePathStyle bool `json:"s3_path_style"`
		// File extensions that support thumbnail generation using native policy API.
		ThumbExts []string `json:"thumb_exts,omitempty"`
		// Whether to support all file extensions for thumbnail generation.
		ThumbSupportAllExts bool `json:"thumb_support_all_exts,omitempty"`
		// ThumbMaxSize indicates the maximum allowed size of a thumbnail. 0 indicates that no limit is set.
		ThumbMaxSize int64 `json:"thumb_max_size,omitempty"`
		// Whether to upload file through server's relay.
		Relay bool `json:"relay,omitempty"`
		// Whether to pre allocate space for file before upload in physical disk.
		PreAllocate bool `json:"pre_allocate,omitempty"`
		// MediaMetaExts file extensions that support media meta generation using native policy API.
		MediaMetaExts []string `json:"media_meta_exts,omitempty"`
		// MediaMetaGeneratorProxy whether to use local proxy to generate media meta.
		MediaMetaGeneratorProxy bool `json:"media_meta_generator_proxy,omitempty"`
		// ThumbGeneratorProxy whether to use local proxy to generate thumbnail.
		ThumbGeneratorProxy bool `json:"thumb_generator_proxy,omitempty"`
		// NativeMediaProcessing whether to use native media processing API from storage provider.
		NativeMediaProcessing bool `json:"native_media_processing"`
		// S3DeleteBatchSize the number of objects to delete in each batch.
		S3DeleteBatchSize int `json:"s3_delete_batch_size,omitempty"`
		// StreamSaver whether to use stream saver to download file in Web.
		StreamSaver bool `json:"stream_saver,omitempty"`
		// UseCname whether to use CNAME for endpoint (OSS).
		UseCname bool `json:"use_cname,omitempty"`
		// CDN domain does not need to be signed.
		SourceAuth bool `json:"source_auth,omitempty"`
	}

	FileType         int
	EntityType       int
	GroupPermission  int
	FilePermission   int
	DavAccountOption int
	NodeCapability   int

	NodeSetting struct {
		Provider            DownloaderProvider `json:"provider,omitempty"`
		*QBittorrentSetting `json:"qbittorrent,omitempty"`
		*Aria2Setting       `json:"aria2,omitempty"`
		// 下载监控间隔
		Interval       int  `json:"interval,omitempty"`
		WaitForSeeding bool `json:"wait_for_seeding,omitempty"`
	}

	DownloaderProvider string

	QBittorrentSetting struct {
		Server   string         `json:"server,omitempty"`
		User     string         `json:"user,omitempty"`
		Password string         `json:"password,omitempty"`
		Options  map[string]any `json:"options,omitempty"`
		TempPath string         `json:"temp_path,omitempty"`
	}

	Aria2Setting struct {
		Server   string         `json:"server,omitempty"`
		Token    string         `json:"token,omitempty"`
		Options  map[string]any `json:"options,omitempty"`
		TempPath string         `json:"temp_path,omitempty"`
	}

	TaskPublicState struct {
		Error            string          `json:"error,omitempty"`
		ErrorHistory     []string        `json:"error_history,omitempty"`
		ExecutedDuration time.Duration   `json:"executed_duration,omitempty"`
		RetryCount       int             `json:"retry_count,omitempty"`
		ResumeTime       int64           `json:"resume_time,omitempty"`
		SlaveTaskProps   *SlaveTaskProps `json:"slave_task_props,omitempty"`
	}

	SlaveTaskProps struct {
		NodeID            int    `json:"node_id,omitempty"`
		MasterSiteURl     string `json:"master_site_u_rl,omitempty"`
		MasterSiteID      string `json:"master_site_id,omitempty"`
		MasterSiteVersion string `json:"master_site_version,omitempty"`
	}

	EntityRecycleOption struct {
		UnlinkOnly bool `json:"unlink_only,omitempty"`
	}

	DavAccountProps struct {
	}

	PolicyType string

	FileProps struct {
	}
)

const (
	GroupPermissionIsAdmin = GroupPermission(iota)
	GroupPermissionIsAnonymous
	GroupPermissionShare
	GroupPermissionWebDAV
	GroupPermissionArchiveDownload
	GroupPermissionArchiveTask
	GroupPermissionWebDAVProxy
	GroupPermissionShareDownload
	GroupPermission_CommunityPlaceholder1
	GroupPermissionRemoteDownload
	GroupPermission_CommunityPlaceholder2
	GroupPermissionRedirectedSource // not used
	GroupPermissionAdvanceDelete
	GroupPermission_CommunityPlaceholder3
	GroupPermission_CommunityPlaceholder4
	GroupPermissionSetExplicitUser_placeholder
	GroupPermissionIgnoreFileOwnership // not used
)

const (
	NodeCapabilityNone NodeCapability = iota
	NodeCapabilityCreateArchive
	NodeCapabilityExtractArchive
	NodeCapabilityRemoteDownload
	NodeCapability_CommunityPlaceholder
)

const (
	FileTypeFile FileType = iota
	FileTypeFolder
)

const (
	EntityTypeVersion EntityType = iota
	EntityTypeThumbnail
	EntityTypeLivePhoto
)

func FileTypeFromString(s string) FileType {
	switch s {
	case "file":
		return FileTypeFile
	case "folder":
		return FileTypeFolder
	}
	return -1
}

const (
	DavAccountReadOnly DavAccountOption = iota
	DavAccountProxy
)

const (
	PolicyTypeLocal  = "local"
	PolicyTypeQiniu  = "qiniu"
	PolicyTypeUpyun  = "upyun"
	PolicyTypeOss    = "oss"
	PolicyTypeCos    = "cos"
	PolicyTypeS3     = "s3"
	PolicyTypeOd     = "onedrive"
	PolicyTypeRemote = "remote"
	PolicyTypeObs    = "obs"
)

const (
	DownloaderProviderAria2       = DownloaderProvider("aria2")
	DownloaderProviderQBittorrent = DownloaderProvider("qbittorrent")
)
