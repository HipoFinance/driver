package repository

import (
	"driver/domain"
	"encoding/json"
	"time"

	"github.com/behrang/sqlbatch"
)

const (
	sqlStakeInsertIfNotExists = `
	insert into stakes as c (
			address, round_since, hash, state, retried, info, create_time, retry_time, success_time
		)
		values (
			$1, $2, $3, 'new', 0, $4::jsonb, now(), null, null
		)
	on conflict (address, round_since, hash) do
		update set
			info = $4::jsonb
`

	sqlStakeFind = `
	select
		address, round_since, hash, state, retried, info, create_time, retry_time, success_time
	from stakes
	where address = $1 and round_since = $2 and hash = $3
`

	sqlStakeFindAllTriable = `
	select
		address, round_since, hash, state, retried, info, create_time, retry_time, success_time
	from stakes
	where state in ('new', 'error') and retried < $1
`

	sqlStakeSetState = `
	update stakes
		set state = $4
	where address = $1 and round_since = $2 and hash = $3
`

	sqlStakeSetRetrying = `
	update stakes
		set retried = retried + 1, retry_time = $4, state = 'inprogress'
	where address = $1 and round_since = $2 and hash = $3
`

	sqlStakeSetSucess = `
	update stakes
		set success_time = $4, state = 'done'
	where address = $1 and round_since = $2 and hash = $3
`
)

type StakeRepository struct {
	batchHandler BatchHandler
}

func NewStakeRepository(db BatchHandler) *StakeRepository {
	return &StakeRepository{batchHandler: db}
}

func readStake(scan func(...interface{}) error) (interface{}, error) {
	r := domain.StakeRequest{}
	var infoJson []byte
	err := scan(
		&r.Address, &r.RoundSince, &r.Hash, &r.State, &r.Retried, &infoJson, &r.CreateTime, &r.RetryTime, &r.SuccessTime,
	)
	if err != nil {
		return &r, err
	}
	err = json.Unmarshal(infoJson, &r.Info)
	return &r, err
}

func readAllStakes(memo interface{}, scan func(...interface{}) error) (interface{}, error) {
	r := domain.StakeRequest{}
	var infoJson []byte
	err := scan(
		&r.Address, &r.RoundSince, &r.Hash, &r.State, &r.Retried, &infoJson, &r.CreateTime, &r.RetryTime, &r.SuccessTime,
	)
	if err == nil {
		err = json.Unmarshal(infoJson, &r.Info)
	}

	list := memo.([]domain.StakeRequest)
	list = append(list, r)
	return list, err
}

func (repo *StakeRepository) InsertIfNotExists(address string, roundSince uint32, hash string, info domain.StakeRelatedInfo) (*domain.StakeRequest, error) {

	infoJson, _ := json.Marshal(info)
	results, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
		{
			Query: sqlStakeInsertIfNotExists,
			Args: []interface{}{
				address, roundSince, hash, infoJson,
			},
			Affect: 1,
		},
		{
			Query:   sqlStakeFind,
			Args:    []interface{}{address, roundSince, hash},
			ReadOne: readStake,
		},
	})

	result, _ := results[1].(*domain.StakeRequest)
	return result, err
}

func (repo *StakeRepository) Find(address string, roundSince uint32, hash string) (*domain.StakeRequest, error) {
	results, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
		{
			Query:   sqlStakeFind,
			Args:    []interface{}{address, roundSince, hash},
			ReadOne: readStake,
		},
	})
	result, _ := results[0].(*domain.StakeRequest)
	return result, err
}

func (repo *StakeRepository) FindAllTriable(maxRetry int) ([]domain.StakeRequest, error) {
	results, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
		{
			Query:   sqlStakeFindAllTriable,
			Args:    []interface{}{maxRetry},
			Init:    make([]domain.StakeRequest, 0),
			ReadAll: readAllStakes,
		},
	})
	result, _ := results[0].([]domain.StakeRequest)
	return result, err
}

func (repo *StakeRepository) SetState(address string, roundSince uint32, hash string, state string) error {
	_, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
		{
			Query:  sqlStakeSetState,
			Args:   []interface{}{address, roundSince, hash, state},
			Affect: 1,
		},
	})
	return err
}

func (repo *StakeRepository) SetRetrying(address string, roundSince uint32, hash string, timestamp time.Time) error {
	_, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
		{
			Query:  sqlStakeSetRetrying,
			Args:   []interface{}{address, roundSince, hash, timestamp},
			Affect: 1,
		},
	})
	return err
}

func (repo *StakeRepository) SetSuccess(address string, roundSince uint32, hash string, timestamp time.Time) error {
	_, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
		{
			Query:  sqlStakeSetSucess,
			Args:   []interface{}{address, roundSince, hash, timestamp},
			Affect: 1,
		},
	})
	return err
}