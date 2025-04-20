package dbfs

import (
	"context"
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

var trashNavigatorCapability = &boolset.BooleanSet{}

// NewTrashNavigator creates a navigator for user's "trash" file system.
func NewTrashNavigator(u *ent.User, fileClient inventory.FileClient, l logging.Logger, config *setting.DBFS,
	hasher hashid.Encoder) Navigator {
	return &trashNavigator{
		user:          u,
		l:             l,
		fileClient:    fileClient,
		config:        config,
		baseNavigator: newBaseNavigator(fileClient, defaultFilter, u, hasher, config),
	}
}

type trashNavigator struct {
	l          logging.Logger
	user       *ent.User
	fileClient inventory.FileClient
	config     *setting.DBFS

	*baseNavigator
}

func (t *trashNavigator) Recycle() {

}

func (n *trashNavigator) PersistState(kv cache.Driver, key string) {
}

func (n *trashNavigator) RestoreState(s State) error {
	return nil
}

func (t *trashNavigator) To(ctx context.Context, path *fs.URI) (*File, error) {
	// Anonymous user does not have a trash folder.
	if inventory.IsAnonymousUser(t.user) {
		return nil, ErrLoginRequired
	}

	elements := path.Elements()
	if len(elements) > 1 {
		// Trash folder is a flatten tree, only 1 layer is supported.
		return nil, fs.ErrPathNotExist.WithError(fmt.Errorf("invalid Path %q", path))
	}

	if len(elements) == 0 {
		// Trash folder has no root.
		return nil, nil
	}

	current, err := t.walkNext(ctx, nil, elements[0], true)
	if err != nil {
		return nil, fmt.Errorf("failed to walk into %q: %w", elements[0], err)
	}

	current.Path[pathIndexUser] = newTrashUri(current.Model.Name)
	current.Path[pathIndexRoot] = current.Path[pathIndexUser]
	current.OwnerModel = t.user
	return current, nil
}

func (t *trashNavigator) Children(ctx context.Context, parent *File, args *ListArgs) (*ListResult, error) {
	if parent != nil {
		return nil, fs.ErrPathNotExist
	}

	res, err := t.baseNavigator.children(ctx, nil, args)
	if err != nil {
		return nil, err
	}

	// Adding user uri for each file.
	for i := 0; i < len(res.Files); i++ {
		res.Files[i].Path[pathIndexUser] = newTrashUri(res.Files[i].Model.Name)
	}

	return res, nil
}

func (t *trashNavigator) Capabilities(isSearching bool) *fs.NavigatorProps {
	res := &fs.NavigatorProps{
		Capability:            trashNavigatorCapability,
		OrderDirectionOptions: fullOrderDirectionOption,
		OrderByOptions:        fullOrderByOption,
		MaxPageSize:           t.config.MaxPageSize,
	}

	if isSearching {
		res.OrderByOptions = searchLimitedOrderByOption
	}

	return res
}

func (t *trashNavigator) Walk(ctx context.Context, levelFiles []*File, limit, depth int, f WalkFunc) error {
	return t.baseNavigator.walk(ctx, levelFiles, limit, depth, f)
}

func (n *trashNavigator) FollowTx(ctx context.Context) (func(), error) {
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

func (n *trashNavigator) ExecuteHook(ctx context.Context, hookType fs.HookType, file *File) error {
	return nil
}
