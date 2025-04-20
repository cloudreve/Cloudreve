package migrator

import (
	"context"
	"fmt"
	"github.com/cloudreve/Cloudreve/v4/application/migrator/conf"
	"github.com/cloudreve/Cloudreve/v4/application/migrator/model"
)

// TODO:
// 1. Policy thumb proxy migration

type (
	settignMigrator func(allSettings map[string]string, name, value string) ([]settingMigrated, error)
	settingMigrated struct {
		name  string
		value string
	}
	// PackProduct 容量包商品
	PackProduct struct {
		ID    int64  `json:"id"`
		Name  string `json:"name"`
		Size  uint64 `json:"size"`
		Time  int64  `json:"time"`
		Price int    `json:"price"`
		Score int    `json:"score"`
	}
	GroupProducts struct {
		ID        int64    `json:"id"`
		Name      string   `json:"name"`
		GroupID   uint     `json:"group_id"`
		Time      int64    `json:"time"`
		Price     int      `json:"price"`
		Score     int      `json:"score"`
		Des       []string `json:"des"`
		Highlight bool     `json:"highlight"`
	}
)

var noopMigrator = func(allSettings map[string]string, name, value string) ([]settingMigrated, error) {
	return nil, nil
}

var migrators = map[string]settignMigrator{
	"siteKeywords":                   noopMigrator,
	"over_used_template":             noopMigrator,
	"download_timeout":               noopMigrator,
	"preview_timeout":                noopMigrator,
	"doc_preview_timeout":            noopMigrator,
	"slave_node_retry":               noopMigrator,
	"slave_ping_interval":            noopMigrator,
	"slave_recover_interval":         noopMigrator,
	"slave_transfer_timeout":         noopMigrator,
	"onedrive_monitor_timeout":       noopMigrator,
	"onedrive_source_timeout":        noopMigrator,
	"share_download_session_timeout": noopMigrator,
	"onedrive_callback_check":        noopMigrator,
	"mail_activation_template":       noopMigrator,
	"mail_reset_pwd_template":        noopMigrator,
	"appid":                          noopMigrator,
	"appkey":                         noopMigrator,
	"wechat_enabled":                 noopMigrator,
	"wechat_appid":                   noopMigrator,
	"wechat_mchid":                   noopMigrator,
	"wechat_serial_no":               noopMigrator,
	"wechat_api_key":                 noopMigrator,
	"wechat_pk_content":              noopMigrator,
	"hot_share_num":                  noopMigrator,
	"defaultTheme":                   noopMigrator,
	"theme_options":                  noopMigrator,
	"max_worker_num":                 noopMigrator,
	"max_parallel_transfer":          noopMigrator,
	"secret_key":                     noopMigrator,
	"avatar_size_m":                  noopMigrator,
	"avatar_size_s":                  noopMigrator,
	"home_view_method":               noopMigrator,
	"share_view_method":              noopMigrator,
	"cron_recycle_upload_session":    noopMigrator,
	"captcha_type": func(allSettings map[string]string, name, value string) ([]settingMigrated, error) {
		if value == "tcaptcha" {
			value = "normal"
		}
		return []settingMigrated{
			{
				name:  "captcha_type",
				value: value,
			},
		}, nil
	},
	"captcha_TCaptcha_CaptchaAppId": noopMigrator,
	"captcha_TCaptcha_AppSecretKey": noopMigrator,
	"captcha_TCaptcha_SecretId":     noopMigrator,
	"captcha_TCaptcha_SecretKey":    noopMigrator,
	"thumb_file_suffix": func(allSettings map[string]string, name, value string) ([]settingMigrated, error) {
		return []settingMigrated{
			{
				name:  "thumb_entity_suffix",
				value: value,
			},
		}, nil
	},
	"thumb_max_src_size": func(allSettings map[string]string, name, value string) ([]settingMigrated, error) {
		return []settingMigrated{
			{
				name:  "thumb_music_cover_max_size",
				value: value,
			},
			{
				name:  "thumb_libreoffice_max_size",
				value: value,
			},
			{
				name:  "thumb_ffmpeg_max_size",
				value: value,
			},
			{
				name:  "thumb_vips_max_size",
				value: value,
			},
			{
				name:  "thumb_builtin_max_size",
				value: value,
			},
		}, nil
	},
	"initial_files":          noopMigrator,
	"office_preview_service": noopMigrator,
	"phone_required":         noopMigrator,
	"phone_enabled":          noopMigrator,
	"wopi_session_timeout": func(allSettings map[string]string, name, value string) ([]settingMigrated, error) {
		return []settingMigrated{
			{
				name:  "viewer_session_timeout",
				value: value,
			},
		}, nil
	},
	"custom_payment_enabled":  noopMigrator,
	"custom_payment_endpoint": noopMigrator,
	"custom_payment_secret":   noopMigrator,
	"custom_payment_name":     noopMigrator,
}

func (m *Migrator) migrateSettings() error {
	m.l.Info("Migrating settings...")
	// 1. List all settings
	var settings []model.Setting
	if err := model.DB.Find(&settings).Error; err != nil {
		return fmt.Errorf("failed to list v3 settings: %w", err)
	}

	m.l.Info("Found %d v3 setting pairs to be migrated.", len(settings))

	allSettings := make(map[string]string)
	for _, s := range settings {
		allSettings[s.Name] = s.Value
	}

	migratedSettings := make([]settingMigrated, 0)
	for _, s := range settings {
		if s.Name == "thumb_file_suffix" {
			m.state.ThumbSuffix = s.Value
		}
		if s.Name == "avatar_path" {
			m.state.V3AvatarPath = s.Value
		}
		migrator, ok := migrators[s.Name]
		if ok {
			newSettings, err := migrator(allSettings, s.Name, s.Value)
			if err != nil {
				return fmt.Errorf("failed to migrate setting %q: %w", s.Name, err)
			}
			migratedSettings = append(migratedSettings, newSettings...)
		} else {
			migratedSettings = append(migratedSettings, settingMigrated{
				name:  s.Name,
				value: s.Value,
			})
		}
	}

	tx, err := m.v4client.Tx(context.Background())
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	// Insert hash_id_salt
	if conf.SystemConfig.HashIDSalt != "" {
		if err := tx.Setting.Create().SetName("hash_id_salt").SetValue(conf.SystemConfig.HashIDSalt).Exec(context.Background()); err != nil {
			if err := tx.Rollback(); err != nil {
				return fmt.Errorf("failed to rollback transaction: %w", err)
			}
			return fmt.Errorf("failed to create setting hash_id_salt: %w", err)
		}
	} else {
		return fmt.Errorf("hash ID salt is not set, please set it from v3 conf file")
	}

	for _, s := range migratedSettings {
		if err := tx.Setting.Create().SetName(s.name).SetValue(s.value).Exec(context.Background()); err != nil {
			if err := tx.Rollback(); err != nil {
				return fmt.Errorf("failed to rollback transaction: %w", err)
			}
			return fmt.Errorf("failed to create setting %q: %w", s.name, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
