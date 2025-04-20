package setting

import (
	"github.com/cloudreve/Cloudreve/v4/ent"
	"github.com/cloudreve/Cloudreve/v4/inventory"
	"github.com/cloudreve/Cloudreve/v4/pkg/boolset"
	"github.com/cloudreve/Cloudreve/v4/pkg/hashid"
	"github.com/samber/lo"
	"time"
)

type ListDavAccountResponse struct {
	Accounts   []DavAccount                 `json:"accounts"`
	Pagination *inventory.PaginationResults `json:"pagination"`
}

func BuildListDavAccountResponse(res *inventory.ListDavAccountResult, hasher hashid.Encoder) *ListDavAccountResponse {
	return &ListDavAccountResponse{
		Accounts: lo.Map(res.Accounts, func(item *ent.DavAccount, index int) DavAccount {
			return BuildDavAccount(item, hasher)
		}),
		Pagination: res.PaginationResults,
	}
}

type DavAccount struct {
	ID        string              `json:"id"`
	CreatedAt time.Time           `json:"created_at"`
	Name      string              `json:"name"`
	Uri       string              `json:"uri"`
	Password  string              `json:"password"`
	Options   *boolset.BooleanSet `json:"options"`
}

func BuildDavAccount(account *ent.DavAccount, hasher hashid.Encoder) DavAccount {
	return DavAccount{
		ID:        hashid.EncodeDavAccountID(hasher, account.ID),
		CreatedAt: account.CreatedAt,
		Name:      account.Name,
		Uri:       account.URI,
		Password:  account.Password,
		Options:   account.Options,
	}
}
