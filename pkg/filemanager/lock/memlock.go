package lock

import (
	"container/heap"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/cloudreve/Cloudreve/v4/inventory/types"
	"github.com/cloudreve/Cloudreve/v4/pkg/hashid"
	"github.com/cloudreve/Cloudreve/v4/pkg/logging"
	"github.com/cloudreve/Cloudreve/v4/pkg/util"
	"github.com/gofrs/uuid"
	"github.com/samber/lo"
)

var (
	// ErrConfirmationFailed is returned by a LockSystem's Confirm method.
	ErrConfirmationFailed = errors.New("memlock: confirmation failed")
	ErrNoSuchLock         = errors.New("memlock: no such lock")
	ErrLocked             = errors.New("memlock: locked")
)

// LockSystem manages access to a collection of named resources. The elements
// in a lock name are separated by slash ('/', U+002F) characters, regardless
// of host operating system convention.
type LockSystem interface {
	Create(now time.Time, details ...LockDetails) ([]string, error)
	Unlock(now time.Time, tokens ...string) error
	Confirm(now time.Time, requests LockInfo) (func(), string, error)
	Refresh(now time.Time, duration time.Duration, token string) (LockDetails, error)
}

// LockDetails are a lock's metadata.
type LockDetails struct {
	// Root is the root resource name being locked. For a zero-depth lock, the
	// root is the only resource being locked.
	Root string
	// Namespace of this lock.
	Ns string
	// Duration is the lock timeout. A negative duration means infinite.
	Duration time.Duration
	// Owner of this lock
	Owner Owner
	// ZeroDepth is whether the lock has zero depth. If it does not have zero
	// depth, it has infinite depth.
	ZeroDepth bool
	// FileType is the type of the file being locked. This is used to display user-friendly error message.
	Type types.FileType
	// Optional, customize the token of the lock.
	Token string
}

func (d *LockDetails) Key() string {
	return d.Ns + "/" + d.Root
}

type Owner struct {
	// Name of the application who are currently lock this.
	Application Application `json:"application"`
}

type Application struct {
	Type     string `json:"type"`
	InnerXML string `json:"inner_xml,omitempty"`
	ViewerID string `json:"viewer_id,omitempty"`
}

// LockInfo is a lock confirmation request.
type LockInfo struct {
	Ns    string
	Root  string
	Token []string
}

type memLS struct {
	l       logging.Logger
	hasher  hashid.Encoder
	mu      sync.Mutex
	byName  map[string]map[string]*memLSNode
	byToken map[string]*memLSNode
	gen     uint64
	// byExpiry only contains those nodes whose LockDetails have a finite
	// Duration and are yet to expire.
	byExpiry byExpiry
}

// NewMemLS returns a new in-memory LockSystem.
func NewMemLS(hasher hashid.Encoder, l logging.Logger) LockSystem {
	return &memLS{
		byName:  make(map[string]map[string]*memLSNode),
		byToken: make(map[string]*memLSNode),
		hasher:  hasher,
		l:       l,
	}
}

func (m *memLS) Confirm(now time.Time, request LockInfo) (func(), string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.collectExpiredNodes(now)

	m.l.Debug("Memlock confirm: NS:%s, Root: %s, Token: %v", request.Ns, request.Root, request.Token)
	n := m.lookup(request.Ns, request.Root, request.Token...)
	if n == nil {
		return nil, "", ErrConfirmationFailed
	}

	m.hold(n)
	return func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		m.unhold(n)
	}, n.token, nil
}

func (m *memLS) Refresh(now time.Time, duration time.Duration, token string) (LockDetails, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.collectExpiredNodes(now)

	m.l.Debug("Memlock refresh: Token: %s, Duration: %v", token, duration)
	n := m.byToken[token]
	if n == nil {
		return LockDetails{}, ErrNoSuchLock
	}
	if n.held {
		return LockDetails{}, ErrLocked
	}
	if n.byExpiryIndex >= 0 {
		heap.Remove(&m.byExpiry, n.byExpiryIndex)
	}
	n.details.Duration = duration
	if n.details.Duration >= 0 {
		n.expiry = now.Add(n.details.Duration)
		heap.Push(&m.byExpiry, n)
	}
	return n.details, nil
}

func (m *memLS) Create(now time.Time, details ...LockDetails) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.collectExpiredNodes(now)

	conflicts := make([]*ConflictDetail, 0)
	locks := make([]*memLSNode, 0, len(details))
	for i, detail := range details {
		// TODO: remove in production
		// if !strings.Contains(detail.Ns, "my") && !strings.Contains(detail.Ns, "trash") {
		// 	panic("invalid namespace")
		// }
		// Check lock conflicts
		detail.Root = util.SlashClean(detail.Root)
		m.l.Debug("Memlock create: NS:%s, Root: %s, Duration: %v, ZeroDepth: %v", detail.Ns, detail.Root, detail.Duration, detail.ZeroDepth)
		conflict := m.canCreate(i, detail.Ns, detail.Root, detail.ZeroDepth)
		if len(conflict) > 0 {
			conflicts = append(conflicts, conflict...)
			// Stop processing more locks since there's already conflicts
			break
		} else {
			// Create locks
			n := m.create(detail.Ns, detail.Root, detail.Token)
			m.byToken[n.token] = n
			n.details = detail
			if n.details.Duration >= 0 {
				n.expiry = now.Add(n.details.Duration)
				heap.Push(&m.byExpiry, n)
			}
			locks = append(locks, n)
		}
	}

	if len(conflicts) > 0 {
		for _, l := range locks {
			m.remove(l)
		}

		return nil, ConflictError(conflicts)
	}

	return lo.Map(locks, func(item *memLSNode, index int) string {
		return item.token
	}), nil
}

func (m *memLS) canCreate(index int, ns, name string, zeroDepth bool) []*ConflictDetail {
	n := m.byName[ns]
	if n == nil {
		return nil
	}

	conflicts := make([]*ConflictDetail, 0)
	canCreate := walkToRoot(name, func(name0 string, first bool) bool {
		n := m.byName[ns][name0]
		if n == nil {
			return true
		}

		if first {
			if n.token != "" {
				// The target node is already locked.
				conflicts = append(conflicts, n.toConflictDetail(index, m.hasher))
				return false
			}
			if !zeroDepth {
				// The requested lock depth is infinite, and the fact that n exists
				// (n != nil) means that a descendent of the target node is locked.
				conflicts = append(conflicts,
					lo.MapToSlice(n.childLocks, func(key string, value *memLSNode) *ConflictDetail {
						return value.toConflictDetail(index, m.hasher)
					},
					)...)
				return false
			}
		} else if n.token != "" && !n.details.ZeroDepth {
			// An ancestor of the target node is locked with infinite depth.
			conflicts = append(conflicts, n.toConflictDetail(index, m.hasher))
			return false
		}
		return true
	})

	if !canCreate {
		return conflicts
	}

	return nil
}

func (m *memLS) Unlock(now time.Time, tokens ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.collectExpiredNodes(now)
	conflicts := make([]*ConflictDetail, 0)
	toBeRemoved := make([]*memLSNode, 0, len(tokens))

	for i, token := range tokens {
		n := m.byToken[token]
		if n == nil {
			return ErrNoSuchLock
		}
		if n.held {
			conflicts = append(conflicts, n.toConflictDetail(i, m.hasher))
		} else {
			toBeRemoved = append(toBeRemoved, n)
		}
	}

	if len(conflicts) > 0 {
		return ConflictError(conflicts)
	}

	for _, n := range toBeRemoved {
		m.remove(n)
	}

	return nil
}

func (m *memLS) collectExpiredNodes(now time.Time) {
	for len(m.byExpiry) > 0 {
		if now.Before(m.byExpiry[0].expiry) {
			break
		}
		m.remove(m.byExpiry[0])
	}
}

func (m *memLS) create(ns, name, token string) (ret *memLSNode) {
	if _, ok := m.byName[ns]; !ok {
		m.byName[ns] = make(map[string]*memLSNode)
	}

	if token == "" {
		token = uuid.Must(uuid.NewV4()).String()
	}

	walkToRoot(name, func(name0 string, first bool) bool {
		n := m.byName[ns][name0]
		if n == nil {
			n = &memLSNode{
				details: LockDetails{
					Root: name0,
				},
				childLocks:    make(map[string]*memLSNode),
				byExpiryIndex: -1,
			}
			m.byName[ns][name0] = n
		}
		n.refCount++
		if first {
			n.token = token
			ret = n
		} else {
			n.childLocks[token] = ret
		}
		return true
	})
	return ret
}

func (m *memLS) lookup(ns, name string, tokens ...string) (n *memLSNode) {
	for _, token := range tokens {
		n = m.byToken[token]
		if n == nil || n.held {
			continue
		}
		if n.details.Ns != ns {
			continue
		}
		if name == n.details.Root {
			return n
		}
		if n.details.ZeroDepth {
			continue
		}
		if n.details.Root == "/" || strings.HasPrefix(name, n.details.Root+"/") {
			return n
		}
	}
	return nil
}

func (m *memLS) remove(n *memLSNode) {
	delete(m.byToken, n.token)
	token := n.token
	n.token = ""
	walkToRoot(n.details.Root, func(name0 string, first bool) bool {
		x := m.byName[n.details.Ns][name0]
		x.refCount--
		delete(x.childLocks, token)
		if x.refCount == 0 {
			delete(m.byName[n.details.Ns], name0)
			if len(m.byName[n.details.Ns]) == 0 {
				delete(m.byName, n.details.Root)
			}
		}
		return true
	})
	if n.byExpiryIndex >= 0 {
		heap.Remove(&m.byExpiry, n.byExpiryIndex)
	}
}

func (m *memLS) hold(n *memLSNode) {
	if n.held {
		panic("dbfs: memLS inconsistent held state")
	}
	n.held = true
	if n.details.Duration >= 0 && n.byExpiryIndex >= 0 {
		heap.Remove(&m.byExpiry, n.byExpiryIndex)
	}
}

func (m *memLS) unhold(n *memLSNode) {
	if !n.held {
		panic("dbfs: memLS inconsistent held state")
	}
	n.held = false
	if n.details.Duration >= 0 {
		heap.Push(&m.byExpiry, n)
	}
}

func walkToRoot(name string, f func(name0 string, first bool) bool) bool {
	for first := true; ; first = false {
		if !f(name, first) {
			return false
		}
		if name == "/" {
			break
		}
		name = name[:strings.LastIndex(name, "/")]
		if name == "" {
			name = "/"
		}
	}
	return true
}

type memLSNode struct {
	// details are the lock metadata. Even if this node's name is not explicitly locked,
	// details.Root will still equal the node's name.
	details LockDetails
	// token is the unique identifier for this node's lock. An empty token means that
	// this node is not explicitly locked.
	token string
	// refCount is the number of self-or-descendent nodes that are explicitly locked.
	refCount int
	// expiry is when this node's lock expires.
	expiry time.Time
	// byExpiryIndex is the index of this node in memLS.byExpiry. It is -1
	// if this node does not expire, or has expired.
	byExpiryIndex int
	// held is whether this node's lock is actively held by a Confirm call.
	held bool
	// childLocks hold the relation between lock token and child locks.
	// This is used to find out who is locking this file.
	childLocks map[string]*memLSNode
}

func (n *memLSNode) toConflictDetail(index int, hasher hashid.Encoder) *ConflictDetail {
	return &ConflictDetail{
		Path: n.details.Root,
		Owner: Owner{
			Application: n.details.Owner.Application,
		},
		Token: n.token,
		Index: index,
		Type:  n.details.Type,
	}
}

type byExpiry []*memLSNode

func (b *byExpiry) Len() int {
	return len(*b)
}

func (b *byExpiry) Less(i, j int) bool {
	return (*b)[i].expiry.Before((*b)[j].expiry)
}

func (b *byExpiry) Swap(i, j int) {
	(*b)[i], (*b)[j] = (*b)[j], (*b)[i]
	(*b)[i].byExpiryIndex = i
	(*b)[j].byExpiryIndex = j
}

func (b *byExpiry) Push(x interface{}) {
	n := x.(*memLSNode)
	n.byExpiryIndex = len(*b)
	*b = append(*b, n)
}

func (b *byExpiry) Pop() interface{} {
	i := len(*b) - 1
	n := (*b)[i]
	(*b)[i] = nil
	n.byExpiryIndex = -1
	*b = (*b)[:i]
	return n
}

// ConflictDetail represent lock conflicts that can be present to end users.
type ConflictDetail struct {
	Path  string         `json:"path,omitempty"`
	Token string         `json:"token,omitempty"`
	Owner Owner          `json:"owner,omitempty"`
	Index int            `json:"-"`
	Type  types.FileType `json:"type"`
}

type ConflictError []*ConflictDetail

func (r ConflictError) Error() string {
	return "conflict with locked resource: " + strings.Join(
		lo.Map(r, func(item *ConflictDetail, index int) string {
			return "\"" + item.Path + "\""
		}), ",")
}

func (r ConflictError) Unwrap() error {
	return ErrLocked
}
