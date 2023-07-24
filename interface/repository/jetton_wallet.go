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
			address, round_since, info, create_time, notify_time
		)
		values (
			$1, $2, $3::jsonb, now(), null
		)
	on conflict (address, round_since) do
		update set
			info = $3::jsonb
`

	sqlJWalletFind = `
	select
		address, round_since, info, create_time, notify_time
	from jwallets
	where address = $1
`

	sqlJWalletFindAllNotNotified = `
	select
		address, round_since, info, create_time, notify_time
	from jwallets
	where notify_time is null
`

	sqlJWalletNotified = `
	update jwallets
		set notify_time = $2
	where address = $1
`

//	sqlJWalletRemove = `
//	delete from jwallet where address = $1
//
// `
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
		&r.Address, &r.RoundSince, &infoJson, &r.CreateTime, &r.NotifyTime,
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
		&r.Address, &r.RoundSince, &infoJson, &r.CreateTime, &r.NotifyTime,
	)
	if err == nil {
		err = json.Unmarshal(infoJson, &r.Info)
	}

	list := memo.([]domain.JettonWallet)
	list = append(list, r)
	return list, err
}

func (repo *JettonWalletRepository) InsertIfNotExists(address string, roundSince uint32, info domain.RelatedTransactionInfo) (*domain.JettonWallet, error) {

	infoJson, _ := json.Marshal(info)
	results, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
		{
			Query: sqlJWalletInsertIfNotExists,
			Args: []interface{}{
				address, roundSince, infoJson,
			},
			Affect: 1,
		},
		{
			Query:   sqlJWalletFind,
			Args:    []interface{}{address},
			ReadOne: readJettonWallet,
		},
	})

	result, _ := results[1].(*domain.JettonWallet)
	return result, err
}

// func (repo *BlockRepository) Find(billId uuid.UUID, reference string) (*domain.Block, error) {
// 	results, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
// 		{
// 			Query:   sqlBlockFind,
// 			Args:    []interface{}{billId, reference},
// 			ReadOne: readBlock,
// 		},
// 	})
// 	result, _ := results[0].(*domain.Block)
// 	return result, err
// }

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

func (repo *JettonWalletRepository) UpdateNotified(address string, timestamp time.Time) error {
	_, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
		{
			Query:  sqlJWalletNotified,
			Args:   []interface{}{address, timestamp},
			Affect: 1,
		},
	})
	return err
}
