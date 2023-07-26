package repository

import (
	"driver/domain"
	"encoding/json"
	"time"

	"github.com/behrang/sqlbatch"
)

const (
	sqlJWalletInsertIfNotExists = `
	insert into jwallets as c (
			address, round_since, hash, state, info, create_time, notify_time
		)
		values (
			$1, $2, $3, 'new', $4::jsonb, now(), null
		)
	on conflict (address, round_since, hash) do
		update set
			info = $4::jsonb
`

	sqlJWalletFind = `
	select
		address, round_since, hash, state, info, create_time, notify_time
	from jwallets
	where address = $1 and round_since = $2 and hash = $3
`

	sqlJWalletFindAllToNotify = `
	select
		address, round_since, hash, state, info, create_time, notify_time
	from jwallets
	where notify_time is null and state in ('new', 'error')
`

	sqlJWalletSetState = `
	update jwallets
		set state = $4
	where address = $1 and round_since = $2 and hash = $3
`

	sqlJWalletSetNotified = `
	update jwallets
		set notify_time = $4, state = 'done'
	where address = $1 and round_since = $2 and hash = $3
`
)

type JettonWalletRepository struct {
	batchHandler BatchHandler
}

func NewJettonWalletRepository(db BatchHandler) *JettonWalletRepository {
	return &JettonWalletRepository{batchHandler: db}
}

func readJettonWallet(scan func(...interface{}) error) (interface{}, error) {
	r := domain.JettonWallet{}
	var infoJson []byte
	err := scan(
		&r.Address, &r.RoundSince, &r.Hash, &r.State, &infoJson, &r.CreateTime, &r.NotifyTime,
	)
	if err != nil {
		return &r, err
	}
	err = json.Unmarshal(infoJson, &r.Info)
	return &r, err
}

func readAllJettonWallets(memo interface{}, scan func(...interface{}) error) (interface{}, error) {
	r := domain.JettonWallet{}
	var infoJson []byte
	err := scan(
		&r.Address, &r.RoundSince, &r.Hash, &r.State, &infoJson, &r.CreateTime, &r.NotifyTime,
	)
	if err == nil {
		err = json.Unmarshal(infoJson, &r.Info)
	}

	list := memo.([]domain.JettonWallet)
	list = append(list, r)
	return list, err
}

func (repo *JettonWalletRepository) InsertIfNotExists(address string, roundSince uint32, hash string, info domain.RelatedTransactionInfo) (*domain.JettonWallet, error) {

	infoJson, _ := json.Marshal(info)
	results, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
		{
			Query: sqlJWalletInsertIfNotExists,
			Args: []interface{}{
				address, roundSince, hash, infoJson,
			},
			Affect: 1,
		},
		{
			Query:   sqlJWalletFind,
			Args:    []interface{}{address, roundSince, hash},
			ReadOne: readJettonWallet,
		},
	})

	result, _ := results[1].(*domain.JettonWallet)
	return result, err
}

func (repo *JettonWalletRepository) Find(address string, roundSince uint32, hash string) (*domain.JettonWallet, error) {
	results, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
		{
			Query:   sqlJWalletFind,
			Args:    []interface{}{address, roundSince, hash},
			ReadOne: readJettonWallet,
		},
	})
	result, _ := results[0].(*domain.JettonWallet)
	return result, err
}

func (repo *JettonWalletRepository) FindAllToNotify() ([]domain.JettonWallet, error) {
	results, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
		{
			Query:   sqlJWalletFindAllToNotify,
			Args:    []interface{}{},
			Init:    make([]domain.JettonWallet, 0),
			ReadAll: readAllJettonWallets,
		},
	})
	result, _ := results[0].([]domain.JettonWallet)
	return result, err
}

func (repo *JettonWalletRepository) SetState(address string, roundSince uint32, hash string, state string) error {
	_, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
		{
			Query:  sqlJWalletSetState,
			Args:   []interface{}{address, roundSince, hash, state},
			Affect: 1,
		},
	})
	return err
}

func (repo *JettonWalletRepository) SetNotified(address string, roundSince uint32, hash string, timestamp time.Time) error {
	_, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
		{
			Query:  sqlJWalletSetNotified,
			Args:   []interface{}{address, roundSince, hash, timestamp},
			Affect: 1,
		},
	})
	return err
}
