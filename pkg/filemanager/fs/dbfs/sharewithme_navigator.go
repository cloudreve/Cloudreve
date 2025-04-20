package dbfs

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudreve/Cloudreve/v4/ent"
	"github.com/cloudreve/Cloudreve/v4/inventory"
	"github.com/cloudreve/Cloudreve/v4/pkg/boolset"
	"github.com/cloudreve/Cloudreve/v4/pkg/cache"
	"github.com/cloudreve/Cloudreve/v4/pkg/filemanager/fs"
	"github.com/cloudreve/Cloudreve/v4/pkg/hashid"
	"github.com/cloudreve/Cloudreve/v4/pkg/logging"
	"github.com/cloudreve/Cloudreve/v4/pkg/setting"
)

var sharedWithMeNavigatorCapability = &boolset.BooleanSet{}

// NewSharedWithMeNavigator creates a navigator for user's "shared with me" file system.
func NewSharedWithMeNavigator(u *ent.User, fileClient inventory.FileClient, l logging.Logger,
	config *setting.DBFS, hasher hashid.Encoder) Navigator {
	n := &sharedWithMeNavigator{
		user:       u,
		l:          l,
		fileClient: fileClient,
		config:     config,
		hasher:     hasher,
	}
	n.baseNavigator = newBaseNavigator(fileClient, defaultFilter, u, hasher, config)
	return n
}

type sharedWithMeNavigator struct {
	l          logging.Logger
	user       *ent.User
	fileClient inventory.FileClient
	config     *setting.DBFS
	hasher     hashid.Encoder

	root *File
	*baseNavigator
}

func (t *sharedWithMeNavigator) Recycle() {

}

func (n *sharedWithMeNavigator) PersistState(kv cache.Driver, key string) {
}

func (n *sharedWithMeNavigator) RestoreState(s State) error {
	return nil
}

func (t *sharedWithMeNavigator) To(ctx context.Context, path *fs.URI) (*File, error) {
	// Anonymous user does not have a trash folder.
	if inventory.IsAnonymousUser(t.user) {
		return nil, ErrLoginRequired
	}

	elements := path.Elements()
	if len(elements) > 0 {
		// Shared with me folder is a flatten tree, only root can be accessed.
		return nil, fs.ErrPathNotExist.WithError(fmt.Errorf("invalid Path %q", path))
	}

	if t.root == nil {
		rootFile, err := t.fileClient.Root(ctx, t.user)
		if err != nil {
			t.l.Info("User's root folder not found: %s, will initialize it.", err)
			return nil, ErrFsNotInitialized
		}

		t.root = newFile(nil, rootFile)
		rootPath := newSharedWithMeUri("")
		t.root.Path[pathIndexRoot], t.root.Path[pathIndexUser] = rootPath, rootPath
		t.root.OwnerModel = t.user
		t.root.IsUserRoot = true
		t.root.CapabilitiesBs = t.Capabilities(false).Capability
	}

	return t.root, nil
}

func (t *sharedWithMeNavigator) Children(ctx context.Context, parent *File, args *ListArgs) (*ListResult, error) {
	args.SharedWithMe = true
	res, err := t.baseNavigator.children(ctx, nil, args)
	if err != nil {
		return nil, err
	}

	// Adding user uri for each file.
	for i := 0; i < len(res.Files); i++ {
		res.Files[i].Path[pathIndexUser] = newSharedWithMeUri(hashid.EncodeFileID(t.hasher, res.Files[i].Model.ID))
	}

	return res, nil
}

func (t *sharedWithMeNavigator) Capabilities(isSearching bool) *fs.NavigatorProps {
	res := &fs.NavigatorProps{
		Capability:            sharedWithMeNavigatorCapability,
		OrderDirectionOptions: fullOrderDirectionOption,
		OrderByOptions:        fullOrderByOption,
		MaxPageSize:           t.config.MaxPageSize,
	}

	if isSearching {
		res.OrderByOptions = searchLimitedOrderByOption
	}

	return res
}

func (t *sharedWithMeNavigator) Walk(ctx context.Context, levelFiles []*File, limit, depth int, f WalkFunc) error {
	return errors.New("not implemented")
}

func (n *sharedWithMeNavigator) FollowTx(ctx context.Context) (func(), error) {
	if _, ok := ctx.Value(inventory.TxCtx{}).(*inventory.Tx); !ok {
		return nil, fmt.Errorf("navigator: no inherited transaction found in context")
	}
	newFileClient, _, _, err := inventory.WithTx(ctx, n.fileClient)
	if err != nil {
		return nil, err
	}

	oldFileClient := n.fileClient
	revert := func() {
		n.fileClient = oldFileClient
		n.baseNavigator.fileClient = oldFileClient
	}

	n.fileClient = newFileClient
	n.baseNavigator.fileClient = newFileClient
	return revert, nil
}

func (n *sharedWithMeNavigator) ExecuteHook(ctx context.Context, hookType fs.HookType, file *File) error {
	return nil
}
