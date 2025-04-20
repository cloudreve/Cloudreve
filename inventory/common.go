package inventory

import (
	"encoding/base64"
	"encoding/json"
	"entgo.io/ent/dialect/sql"
	"fmt"
	"github.com/cloudreve/Cloudreve/v4/pkg/conf"
	"github.com/cloudreve/Cloudreve/v4/pkg/hashid"
	"time"
)

type (
	OrderDirection string
	PaginationArgs struct {
		UseCursorPagination bool
		Page                int
		PageToken           string
		PageSize            int
		OrderBy             string
		Order               OrderDirection
	}

	PaginationResults struct {
		Page          int    `json:"page"`
		PageSize      int    `json:"page_size"`
		TotalItems    int    `json:"total_items,omitempty"`
		NextPageToken string `json:"next_token,omitempty"`
		IsCursor      bool   `json:"is_cursor,omitempty"`
	}

	PageToken struct {
		Time          *time.Time `json:"time,omitempty"`
		ID            int        `json:"-"`
		IDHash        string     `json:"id,omitempty"`
		String        string     `json:"string,omitempty"`
		Int           int        `json:"int,omitempty"`
		StartWithFile bool       `json:"start_with_file,omitempty"`
	}
)

const (
	OrderDirectionAsc  = OrderDirection("asc")
	OrderDirectionDesc = OrderDirection("desc")
)

var (
	ErrTooManyArguments = fmt.Errorf("too many arguments")
)

func pageTokenFromString(s string, hasher hashid.Encoder, idType int) (*PageToken, error) {
	sB64Decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 for page token: %w", err)
	}

	token := &PageToken{}
	if err := json.Unmarshal(sB64Decoded, token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal page token: %w", err)
	}

	id, err := hasher.Decode(token.IDHash, idType)
	if err != nil {
		return nil, fmt.Errorf("failed to decode id: %w", err)
	}

	if token.Time == nil {
		token.Time = &time.Time{}
	}

	token.ID = id
	return token, nil
}

func (p *PageToken) Encode(hasher hashid.Encoder, encodeFunc hashid.EncodeFunc) (string, error) {
	p.IDHash = encodeFunc(hasher, p.ID)
	res, err := json.Marshal(p)
	if err != nil {
		return "", fmt.Errorf("failed to marshal page token: %w", err)
	}

	return base64.StdEncoding.EncodeToString(res), nil
}

// sqlParamLimit returns the max number of sql parameters.
func sqlParamLimit(dbType conf.DBType) int {
	switch dbType {
	case conf.PostgresDB:
		return 34464
	case conf.SQLiteDB, conf.SQLite3DB:
		// https://www.sqlite.org/limits.html
		return 32766
	default:
		return 32766
	}
}

// getOrderTerm returns the order term for ent.
func getOrderTerm(d OrderDirection) sql.OrderTermOption {
	switch d {
	case OrderDirectionDesc:
		return sql.OrderDesc()
	default:
		return sql.OrderAsc()
	}
}

func capPageSize(maxSQlParam, preferredSize, margin int) int {
	// Page size should not be bigger than max SQL parameter
	pageSize := preferredSize
	if maxSQlParam > 0 && pageSize > maxSQlParam-margin || pageSize == 0 {
		pageSize = maxSQlParam - margin
	}

	return pageSize
}

type StorageDiff map[int]int64

func (s *StorageDiff) Merge(diff StorageDiff) {
	for k, v := range diff {
		(*s)[k] += v
	}
}
