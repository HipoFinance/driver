package repository

import (
	"driver/domain"
	"encoding/json"
	"math/big"
	"time"

	"github.com/behrang/sqlbatch"
)

const (
	sqlUntakeInsertIfNotExists = `
	insert into unstakes as c (
			address, tokens, hash, state, retry_count, info, create_time, retry_time, sent_time, verified_time
		)
		values (
			$1, $2, $3, 'new', 0, $4::jsonb, now(), null, null, null
		)
	on conflict (hash) do
		update set
			info = $4::jsonb
`

	sqlUnstakeFind = `
	select
		address, tokens, hash, state, retry_count, info, create_time, retry_time, sent_time, verified_time
	from unstakes
	where hash = $1
`

	sqlUnstakeFindAllTriable = `
	select
		address, tokens, hash, state, retry_count, info, create_time, retry_time, sent_time, verified_time
	from unstakes
	where state in ('new', 'error', 'retriable') and retry_count < $1
`

	sqlUnstakeFindAllVerifiable = `
	select
		address, tokens, hash, state, retry_count, info, create_time, retry_time, sent_time, verified_time
	from unstakes
	where state in ('sent')
`

	sqlUnstakeSetState = `
	update unstakes
		set state = $2
	where hash = $1
`

	sqlUntakeSetRetrying = `
	update unstakes
		set retry_count = retry_count + 1, retry_time = $2, state = 'inprogress'
	where hash = $1
`

	sqlUntakeSetSent = `
	update unstakes
		set sent_time = $2, state = 'sent'
	where hash = $1
`

	sqlUntakeSetVerified = `
	update unstakes
		set sent_time = $2, state = 'verified'
	where hash = $1
`
)

type UnstakeRepository struct {
	batchHandler BatchHandler
}

func NewUnstakeRepository(db BatchHandler) *UnstakeRepository {
	return &UnstakeRepository{batchHandler: db}
}

func readUnstake(scan func(...interface{}) error) (interface{}, error) {
	r := domain.UnstakeRequest{}
	var tokenStr string
	var infoJson []byte
	err := scan(
		&r.Address, &tokenStr, &r.Hash, &r.State, &r.RetryCount, &infoJson, &r.CreateTime, &r.RetryTime, &r.SentTime, &r.VerifiedTime,
	)
	if err != nil {
		return &r, err
	}
	err = r.Tokens.UnmarshalText([]byte(tokenStr))
	if err != nil {
		return &r, err
	}
	err = json.Unmarshal(infoJson, &r.Info)
	return &r, err
}

func readAllUnstakes(memo interface{}, scan func(...interface{}) error) (interface{}, error) {
	r := domain.UnstakeRequest{}
	var tokenStr string
	var infoJson []byte
	err := scan(
		&r.Address, &tokenStr, &r.Hash, &r.State, &r.RetryCount, &infoJson, &r.CreateTime, &r.RetryTime, &r.SentTime, &r.VerifiedTime,
	)

	if err == nil {
		err = r.Tokens.UnmarshalText([]byte(tokenStr))
	}

	if err == nil {
		err = json.Unmarshal(infoJson, &r.Info)
	}

	list := memo.([]*domain.UnstakeRequest)
	list = append(list, &r)
	return list, err
}

func (repo *UnstakeRepository) InsertIfNotExists(address string, tokens big.Int, hash string, info domain.UnstakeRelatedInfo) (*domain.UnstakeRequest, error) {

	infoJson, _ := json.Marshal(info)
	results, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
		{
			Query: sqlUntakeInsertIfNotExists,
			Args: []interface{}{
				address, tokens.String(), hash, infoJson,
			},
			Affect: 1,
		},
		{
			Query:   sqlUnstakeFind,
			Args:    []interface{}{hash},
			ReadOne: readUnstake,
		},
	})

	result, _ := results[1].(*domain.UnstakeRequest)
	return result, err
}

func (repo *UnstakeRepository) Find(hash string) (*domain.UnstakeRequest, error) {
	results, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
		{
			Query:   sqlUnstakeFind,
			Args:    []interface{}{hash},
			ReadOne: readUnstake,
		},
	})
	result, _ := results[0].(*domain.UnstakeRequest)
	return result, err
}

func (repo *UnstakeRepository) FindAllTriable(maxRetry int) ([]*domain.UnstakeRequest, error) {
	results, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
		{
			Query:   sqlUnstakeFindAllTriable,
			Args:    []interface{}{maxRetry},
			Init:    make([]*domain.UnstakeRequest, 0),
			ReadAll: readAllUnstakes,
		},
	})
	result, _ := results[0].([]*domain.UnstakeRequest)
	return result, err
}

func (repo *UnstakeRepository) FindAllVerifiable() ([]*domain.UnstakeRequest, error) {
	results, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
		{
			Query:   sqlUnstakeFindAllVerifiable,
			Args:    []interface{}{},
			Init:    make([]*domain.UnstakeRequest, 0),
			ReadAll: readAllUnstakes,
		},
	})
	result, _ := results[0].([]*domain.UnstakeRequest)
	return result, err
}

func (repo *UnstakeRepository) SetState(hash string, state string) error {
	_, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
		{
			Query:  sqlUnstakeSetState,
			Args:   []interface{}{hash, state},
			Affect: 1,
		},
	})
	return err
}

func (repo *UnstakeRepository) SetRetrying(hash string, timestamp time.Time) error {
	_, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
		{
			Query:  sqlUntakeSetRetrying,
			Args:   []interface{}{hash, timestamp},
			Affect: 1,
		},
	})
	return err
}

func (repo *UnstakeRepository) SetSent(hash string, timestamp time.Time) error {
	_, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
		{
			Query:  sqlUntakeSetSent,
			Args:   []interface{}{hash, timestamp},
			Affect: 1,
		},
	})
	return err
}

func (repo *UnstakeRepository) SetVerified(hash string, timestamp time.Time) error {
	_, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
		{
			Query:  sqlUntakeSetVerified,
			Args:   []interface{}{hash, timestamp},
			Affect: 1,
		},
	})
	return err
}
