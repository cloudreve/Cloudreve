package inventory

import (
	"context"
	rawsql "database/sql"
	"database/sql/driver"
	"fmt"
	"os"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/cloudreve/Cloudreve/v4/application/constants"
	"github.com/cloudreve/Cloudreve/v4/ent"
	"github.com/cloudreve/Cloudreve/v4/ent/group"
	"github.com/cloudreve/Cloudreve/v4/ent/node"
	_ "github.com/cloudreve/Cloudreve/v4/ent/runtime"
	"github.com/cloudreve/Cloudreve/v4/ent/setting"
	"github.com/cloudreve/Cloudreve/v4/ent/storagepolicy"
	"github.com/cloudreve/Cloudreve/v4/inventory/debug"
	"github.com/cloudreve/Cloudreve/v4/inventory/types"
	"github.com/cloudreve/Cloudreve/v4/pkg/boolset"
	"github.com/cloudreve/Cloudreve/v4/pkg/cache"
	"github.com/cloudreve/Cloudreve/v4/pkg/conf"
	"github.com/cloudreve/Cloudreve/v4/pkg/logging"
	"github.com/cloudreve/Cloudreve/v4/pkg/util"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"modernc.org/sqlite"
)

const (
	DBVersionPrefix           = "db_version_"
	EnvDefaultOverwritePrefix = "CR_SETTING_DEFAULT_"
	EnvEnableAria2            = "CR_ENABLE_ARIA2"
)

// InitializeDBClient runs migration and returns a new ent.Client with additional configurations
// for hooks and interceptors.
func InitializeDBClient(l logging.Logger,
	client *ent.Client, kv cache.Driver, requiredDbVersion string) (*ent.Client, error) {
	ctx := context.WithValue(context.Background(), logging.LoggerCtx{}, l)
	if needMigration(client, ctx, requiredDbVersion) {
		// Run the auto migration tool.
		if err := migrate(l, client, ctx, kv, requiredDbVersion); err != nil {
			return nil, fmt.Errorf("failed to migrate database: %w", err)
		}
	} else {
		l.Info("Database schema is up to date.")
	}

	//createMockData(client, ctx)
	return client, nil
}

// NewRawEntClient returns a new ent.Client without additional configurations.
func NewRawEntClient(l logging.Logger, config conf.ConfigProvider) (*ent.Client, error) {
	l.Info("Initializing database connection...")
	dbConfig := config.Database()
	confDBType := dbConfig.Type
	if confDBType == conf.SQLite3DB || confDBType == "" {
		confDBType = conf.SQLiteDB
	}

	var (
		err    error
		client *sql.Driver
	)

	switch confDBType {
	case conf.SQLiteDB:
		dbFile := util.RelativePath(dbConfig.DBFile)
		l.Info("Connect to SQLite database %q.", dbFile)
		client, err = sql.Open("sqlite3", util.RelativePath(dbConfig.DBFile))
	case conf.PostgresDB:
		l.Info("Connect to Postgres database %q.", dbConfig.Host)
		client, err = sql.Open("postgres", fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
			dbConfig.Host,
			dbConfig.User,
			dbConfig.Password,
			dbConfig.Name,
			dbConfig.Port))
	case conf.MySqlDB, conf.MsSqlDB:
		l.Info("Connect to MySQL/SQLServer database %q.", dbConfig.Host)
		var host string
		if dbConfig.UnixSocket {
			host = fmt.Sprintf("unix(%s)",
				dbConfig.Host)
		} else {
			host = fmt.Sprintf("(%s:%d)",
				dbConfig.Host,
				dbConfig.Port)
		}

		client, err = sql.Open(string(confDBType), fmt.Sprintf("%s:%s@%s/%s?charset=%s&parseTime=True&loc=Local",
			dbConfig.User,
			dbConfig.Password,
			host,
			dbConfig.Name,
			dbConfig.Charset))
	default:
		return nil, fmt.Errorf("unsupported database type %q", confDBType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool
	db := client.DB()
	db.SetMaxIdleConns(50)
	if confDBType == "sqlite" || confDBType == "UNSET" {
		db.SetMaxOpenConns(1)
	} else {
		db.SetMaxOpenConns(100)
	}

	// Set timeout
	db.SetConnMaxLifetime(time.Second * 30)

	driverOpt := ent.Driver(client)

	// Enable verbose logging for debug mode.
	if config.System().Debug {
		l.Debug("Debug mode is enabled for DB client.")
		driverOpt = ent.Driver(debug.DebugWithContext(client, func(ctx context.Context, i ...any) {
			logging.FromContext(ctx).Debug(i[0].(string), i[1:]...)
		}))
	}

	return ent.NewClient(driverOpt), nil
}

type sqlite3Driver struct {
	*sqlite.Driver
}

type sqlite3DriverConn interface {
	Exec(string, []driver.Value) (driver.Result, error)
}

func (d sqlite3Driver) Open(name string) (conn driver.Conn, err error) {
	conn, err = d.Driver.Open(name)
	if err != nil {
		return
	}
	_, err = conn.(sqlite3DriverConn).Exec("PRAGMA foreign_keys = ON;", nil)
	if err != nil {
		_ = conn.Close()
	}
	return
}

func init() {
	rawsql.Register("sqlite3", sqlite3Driver{Driver: &sqlite.Driver{}})
}

// needMigration exams if required schema version is satisfied.
func needMigration(client *ent.Client, ctx context.Context, requiredDbVersion string) bool {
	c, _ := client.Setting.Query().Where(setting.NameEQ(DBVersionPrefix + requiredDbVersion)).Count(ctx)
	return c == 0
}

func migrate(l logging.Logger, client *ent.Client, ctx context.Context, kv cache.Driver, requiredDbVersion string) error {
	l.Info("Start initializing database schema...")
	l.Info("Creating basic table schema...")
	if err := client.Schema.Create(ctx); err != nil {
		return fmt.Errorf("Failed creating schema resources: %w", err)
	}

	migrateDefaultSettings(l, client, ctx, kv)

	if err := migrateDefaultStoragePolicy(l, client, ctx); err != nil {
		return fmt.Errorf("failed migrating default storage policy: %w", err)
	}

	if err := migrateSysGroups(l, client, ctx); err != nil {
		return fmt.Errorf("failed migrating default storage policy: %w", err)
	}

	client.Setting.Create().SetName(DBVersionPrefix + requiredDbVersion).SetValue("installed").Save(ctx)
	return nil
}

func migrateDefaultSettings(l logging.Logger, client *ent.Client, ctx context.Context, kv cache.Driver) {
	// clean kv cache
	if err := kv.DeleteAll(); err != nil {
		l.Warning("Failed to remove all KV entries while schema migration: %s", err)
	}

	// List existing settings into a map
	existingSettings := make(map[string]struct{})
	settings, err := client.Setting.Query().All(ctx)
	if err != nil {
		l.Warning("Failed to query existing settings: %s", err)
	}

	for _, s := range settings {
		existingSettings[s.Name] = struct{}{}
	}

	l.Info("Insert default settings...")
	for k, v := range DefaultSettings {
		if _, ok := existingSettings[k]; ok {
			l.Debug("Skip inserting setting %s, already exists.", k)
			continue
		}

		if override, ok := os.LookupEnv(EnvDefaultOverwritePrefix + k); ok {
			l.Info("Override default setting %q with env value %q", k, override)
			v = override
		}

		client.Setting.Create().SetName(k).SetValue(v).SaveX(ctx)
	}
}

func migrateDefaultStoragePolicy(l logging.Logger, client *ent.Client, ctx context.Context) error {
	if _, err := client.StoragePolicy.Query().Where(storagepolicy.ID(1)).First(ctx); err == nil {
		l.Info("Default storage policy (ID=1) already exists, skip migrating.")
		return nil
	}

	l.Info("Insert default storage policy...")
	if _, err := client.StoragePolicy.Create().
		SetName("Default storage policy").
		SetType(types.PolicyTypeLocal).
		SetDirNameRule(util.DataPath("uploads/{uid}/{path}")).
		SetFileNameRule("{uid}_{randomkey8}_{originname}").
		SetSettings(&types.PolicySetting{
			ChunkSize:   25 << 20, // 25MB
			PreAllocate: true,
		}).
		Save(ctx); err != nil {
		return fmt.Errorf("failed to create default storage policy: %w", err)
	}

	return nil
}

func migrateSysGroups(l logging.Logger, client *ent.Client, ctx context.Context) error {
	if err := migrateAdminGroup(l, client, ctx); err != nil {
		return err
	}

	if err := migrateUserGroup(l, client, ctx); err != nil {
		return err
	}

	if err := migrateAnonymousGroup(l, client, ctx); err != nil {
		return err
	}

	if err := migrateMasterNode(l, client, ctx); err != nil {
		return err
	}

	return nil
}

func migrateAdminGroup(l logging.Logger, client *ent.Client, ctx context.Context) error {
	if _, err := client.Group.Query().Where(group.ID(1)).First(ctx); err == nil {
		l.Info("Default admin group (ID=1) already exists, skip migrating.")
		return nil
	}

	l.Info("Insert default admin group...")
	permissions := &boolset.BooleanSet{}
	boolset.Sets(map[types.GroupPermission]bool{
		types.GroupPermissionIsAdmin:             true,
		types.GroupPermissionShare:               true,
		types.GroupPermissionWebDAV:              true,
		types.GroupPermissionWebDAVProxy:         true,
		types.GroupPermissionArchiveDownload:     true,
		types.GroupPermissionArchiveTask:         true,
		types.GroupPermissionShareDownload:       true,
		types.GroupPermissionRemoteDownload:      true,
		types.GroupPermissionRedirectedSource:    true,
		types.GroupPermissionAdvanceDelete:       true,
		types.GroupPermissionIgnoreFileOwnership: true,
		// TODO: review default permission
	}, permissions)
	if _, err := client.Group.Create().
		SetName("Admin").
		SetStoragePoliciesID(1).
		SetMaxStorage(1 * constants.TB). // 1 TB default storage
		SetPermissions(permissions).
		SetSettings(&types.GroupSetting{
			SourceBatchSize:  1000,
			Aria2BatchSize:   50,
			MaxWalkedFiles:   100000,
			TrashRetention:   7 * 24 * 3600,
			RedirectedSource: true,
		}).
		Save(ctx); err != nil {
		return fmt.Errorf("failed to create default admin group: %w", err)
	}

	return nil
}

func migrateUserGroup(l logging.Logger, client *ent.Client, ctx context.Context) error {
	if _, err := client.Group.Query().Where(group.ID(2)).First(ctx); err == nil {
		l.Info("Default user group (ID=2) already exists, skip migrating.")
		return nil
	}

	l.Info("Insert default user group...")
	permissions := &boolset.BooleanSet{}
	boolset.Sets(map[types.GroupPermission]bool{
		types.GroupPermissionShare:            true,
		types.GroupPermissionShareDownload:    true,
		types.GroupPermissionRedirectedSource: true,
	}, permissions)
	if _, err := client.Group.Create().
		SetName("User").
		SetStoragePoliciesID(1).
		SetMaxStorage(1 * constants.GB). // 1 GB default storage
		SetPermissions(permissions).
		SetSettings(&types.GroupSetting{
			SourceBatchSize:  10,
			Aria2BatchSize:   1,
			MaxWalkedFiles:   100000,
			TrashRetention:   7 * 24 * 3600,
			RedirectedSource: true,
		}).
		Save(ctx); err != nil {
		return fmt.Errorf("failed to create default user group: %w", err)
	}

	return nil
}

func migrateAnonymousGroup(l logging.Logger, client *ent.Client, ctx context.Context) error {
	if _, err := client.Group.Query().Where(group.ID(AnonymousGroupID)).First(ctx); err == nil {
		l.Info("Default anonymous group (ID=3) already exists, skip migrating.")
		return nil
	}

	l.Info("Insert default anonymous group...")
	permissions := &boolset.BooleanSet{}
	boolset.Sets(map[types.GroupPermission]bool{
		types.GroupPermissionIsAnonymous:   true,
		types.GroupPermissionShareDownload: true,
	}, permissions)
	if _, err := client.Group.Create().
		SetName("Anonymous").
		SetPermissions(permissions).
		SetSettings(&types.GroupSetting{
			MaxWalkedFiles:   100000,
			RedirectedSource: true,
		}).
		Save(ctx); err != nil {
		return fmt.Errorf("failed to create default anonymous group: %w", err)
	}

	return nil
}

func migrateMasterNode(l logging.Logger, client *ent.Client, ctx context.Context) error {
	if _, err := client.Node.Query().Where(node.TypeEQ(node.TypeMaster)).First(ctx); err == nil {
		l.Info("Default master node already exists, skip migrating.")
		return nil
	}

	capabilities := &boolset.BooleanSet{}
	boolset.Sets(map[types.NodeCapability]bool{
		types.NodeCapabilityCreateArchive:  true,
		types.NodeCapabilityExtractArchive: true,
		types.NodeCapabilityRemoteDownload: true,
	}, capabilities)

	stm := client.Node.Create().
		SetType(node.TypeMaster).
		SetCapabilities(capabilities).
		SetName("Master").
		SetSettings(&types.NodeSetting{
			Provider: types.DownloaderProviderAria2,
		}).
		SetStatus(node.StatusActive)

	_, enableAria2 := os.LookupEnv(EnvEnableAria2)
	if enableAria2 {
		l.Info("Aria2 is override as enabled.")
		stm.SetSettings(&types.NodeSetting{
			Provider: types.DownloaderProviderAria2,
			Aria2Setting: &types.Aria2Setting{
				Server: "http://127.0.0.1:6800/jsonrpc",
			},
		})
	}

	l.Info("Insert default master node...")
	if _, err := stm.Save(ctx); err != nil {
		return fmt.Errorf("failed to create default master node: %w", err)
	}

	return nil
}

func createMockData(client *ent.Client, ctx context.Context) {
	//userCount := 100
	//folderCount := 10000
	//fileCount := 25000
	//
	//// create users
	//pwdDigest, _ := digestPassword("52121225")
	//userCreates := make([]*ent.UserCreate, userCount)
	//for i := 0; i < userCount; i++ {
	//	nick := uuid.Must(uuid.NewV4()).String()
	//	userCreates[i] = client.User.Create().
	//		SetEmail(nick + "@cloudreve.org").
	//		SetNick(nick).
	//		SetPassword(pwdDigest).
	//		SetStatus(user.StatusActive).
	//		SetGroupID(1)
	//}
	//users, err := client.User.CreateBulk(userCreates...).Save(ctx)
	//if err != nil {
	//	panic(err)
	//}
	//
	//// Create root folder
	//rootFolderCreates := make([]*ent.FileCreate, userCount)
	//folderIds := make([][]int, 0, folderCount*userCount+userCount)
	//for i, user := range users {
	//	rootFolderCreates[i] = client.File.Create().
	//		SetName(RootFolderName).
	//		SetOwnerID(user.ID).
	//		SetType(int(FileTypeFolder))
	//}
	//rootFolders, err := client.File.CreateBulk(rootFolderCreates...).Save(ctx)
	//for _, rootFolders := range rootFolders {
	//	folderIds = append(folderIds, []int{rootFolders.ID, rootFolders.OwnerID})
	//}
	//if err != nil {
	//	panic(err)
	//}
	//
	//// create random folder
	//for i := 0; i < folderCount*userCount; i++ {
	//	parent := lo.Sample(folderIds)
	//	res := client.File.Create().
	//		SetName(uuid.Must(uuid.NewV4()).String()).
	//		SetType(int(FileTypeFolder)).
	//		SetOwnerID(parent[1]).
	//		SetFileChildren(parent[0]).
	//		SaveX(ctx)
	//	folderIds = append(folderIds, []int{res.ID, res.OwnerID})
	//}

	for i := 0; i < 255; i++ {
		fmt.Printf("%d/", i)
	}
}
