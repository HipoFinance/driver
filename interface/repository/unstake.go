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
			address, tokens, hash, state, retried, info, create_time, retry_time, success_time
		)
		values (
			$1, $2, $3, 'new', 0, $4::jsonb, now(), null, null
		)
	on conflict (address, tokens, hash) do
		update set
			info = $4::jsonb
`

	sqlUnstakeFind = `
	select
		address, tokens, hash, state, retried, info, create_time, retry_time, success_time
	from unstakes
	where address = $1 and tokens = $2 and hash = $3
`

	sqlUnstakeFindAllTriable = `
	select
		address, tokens, hash, state, retried, info, create_time, retry_time, success_time
	from unstakes
	where state in ('new', 'error') and retried < $1
`

	sqlUnstakeSetState = `
	update unstakes
		set state = $4
	where address = $1 and tokens = $2 and hash = $3
`

	sqlUntakeSetRetrying = `
	update unstakes
		set retried = retried + 1, retry_time = $4, state = 'inprogress'
	where address = $1 and tokens = $2 and hash = $3
`

	sqlUntakeSetSucess = `
	update unstakes
		set success_time = $4, state = 'done'
	where address = $1 and tokens = $2 and hash = $3
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
	var infoJson []byte
	err := scan(
		&r.Address, &r.Tokens, &r.Hash, &r.State, &r.Retried, &infoJson, &r.CreateTime, &r.RetryTime, &r.SuccessTime,
	)
	if err != nil {
		return &r, err
	}
	err = json.Unmarshal(infoJson, &r.Info)
	return &r, err
}

func readAllUnstakes(memo interface{}, scan func(...interface{}) error) (interface{}, error) {
	r := domain.UnstakeRequest{}
	var infoJson []byte
	err := scan(
		&r.Address, &r.Tokens, &r.Hash, &r.State, &r.Retried, &infoJson, &r.CreateTime, &r.RetryTime, &r.SuccessTime,
	)
	if err == nil {
		err = json.Unmarshal(infoJson, &r.Info)
	}

	list := memo.([]domain.UnstakeRequest)
	list = append(list, r)
	return list, err
}

func (repo *UnstakeRepository) InsertIfNotExists(address string, tokens big.Int, hash string, info domain.UnstakeRelatedInfo) (*domain.UnstakeRequest, error) {

	infoJson, _ := json.Marshal(info)
	results, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
		{
			Query: sqlUntakeInsertIfNotExists,
			Args: []interface{}{
				address, tokens, hash, infoJson,
			},
			Affect: 1,
		},
		{
			Query:   sqlUnstakeFind,
			Args:    []interface{}{address, tokens, hash},
			ReadOne: readUnstake,
		},
	})

	result, _ := results[1].(*domain.UnstakeRequest)
	return result, err
}

func (repo *UnstakeRepository) Find(address string, tokens big.Int, hash string) (*domain.UnstakeRequest, error) {
	results, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
		{
			Query:   sqlUnstakeFind,
			Args:    []interface{}{address, tokens, hash},
			ReadOne: readUnstake,
		},
	})
	result, _ := results[0].(*domain.UnstakeRequest)
	return result, err
}

func (repo *UnstakeRepository) FindAllTriable(maxRetry int) ([]domain.UnstakeRequest, error) {
	results, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
		{
			Query:   sqlUnstakeFindAllTriable,
			Args:    []interface{}{maxRetry},
			Init:    make([]domain.UnstakeRequest, 0),
			ReadAll: readAllUnstakes,
		},
	})
	result, _ := results[0].([]domain.UnstakeRequest)
	return result, err
}

func (repo *UnstakeRepository) SetState(address string, tokens big.Int, hash string, state string) error {
	_, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
		{
			Query:  sqlUnstakeSetState,
			Args:   []interface{}{address, tokens, hash, state},
			Affect: 1,
		},
	})
	return err
}

func (repo *UnstakeRepository) SetRetrying(address string, tokens big.Int, hash string, timestamp time.Time) error {
	_, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
		{
			Query:  sqlUntakeSetRetrying,
			Args:   []interface{}{address, tokens, hash, timestamp},
			Affect: 1,
		},
	})
	return err
}

func (repo *UnstakeRepository) SetSuccess(address string, tokens big.Int, hash string, timestamp time.Time) error {
	_, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
		{
			Query:  sqlUntakeSetSucess,
			Args:   []interface{}{address, tokens, hash, timestamp},
			Affect: 1,
		},
	})
	return err
}
