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
			address, round_since, msg_hash, info, create_time, notify_time
		)
		values (
			$1, $2, $3, $4::jsonb, now(), null
		)
	on conflict (address, round_since, msg_hash) do
		update set
			info = $4::jsonb
`

	sqlJWalletFind = `
	select
		address, round_since, msg_hash, info, create_time, notify_time
	from jwallets
	where address = $1 and round_since = $2 and msg_hash = $3
`

	sqlJWalletFindAllNotNotified = `
	select
		address, round_since, msg_hash, info, create_time, notify_time
	from jwallets
	where notify_time is null
`

	sqlJWalletNotified = `
	update jwallets
		set notify_time = $4
	where address = $1 and round_since = $2 and msg_hash = $3
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
		&r.Address, &r.RoundSince, &r.MsgHash, &infoJson, &r.CreateTime, &r.NotifyTime,
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
		&r.Address, &r.RoundSince, &r.MsgHash, &infoJson, &r.CreateTime, &r.NotifyTime,
	)
	if err == nil {
		err = json.Unmarshal(infoJson, &r.Info)
	}

	list := memo.([]domain.JettonWallet)
	list = append(list, r)
	return list, err
}

func (repo *JettonWalletRepository) InsertIfNotExists(address string, roundSince uint32, msgHash string, info domain.RelatedTransactionInfo) (*domain.JettonWallet, error) {

	infoJson, _ := json.Marshal(info)
	results, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
		{
			Query: sqlJWalletInsertIfNotExists,
			Args: []interface{}{
				address, roundSince, msgHash, infoJson,
			},
			Affect: 1,
		},
		{
			Query:   sqlJWalletFind,
			Args:    []interface{}{address, roundSince, msgHash},
			ReadOne: readJettonWallet,
		},
	})

	result, _ := results[1].(*domain.JettonWallet)
	return result, err
}

func (repo *JettonWalletRepository) Find(address string, roundSince uint32, msgHash string) (*domain.JettonWallet, error) {
	results, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
		{
			Query:   sqlJWalletFind,
			Args:    []interface{}{address, roundSince, msgHash},
			ReadOne: readJettonWallet,
		},
	})
	result, _ := results[0].(*domain.JettonWallet)
	return result, err
}

func (repo *JettonWalletRepository) FindAllNotNotified() ([]domain.JettonWallet, error) {
	results, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
		{
			Query:   sqlJWalletFindAllNotNotified,
			Args:    []interface{}{},
			Init:    make([]domain.JettonWallet, 0),
			ReadAll: readAllJettonWallets,
		},
	})
	result, _ := results[0].([]domain.JettonWallet)
	return result, err
}

func (repo *JettonWalletRepository) UpdateNotified(address string, roundSince uint32, msgHash string, timestamp time.Time) error {
	_, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
		{
			Query:  sqlJWalletNotified,
			Args:   []interface{}{address, roundSince, msgHash, timestamp},
			Affect: 1,
		},
	})
	return err
}
