package repository

import (
	"driver/domain"
	"encoding/json"

	"github.com/behrang/sqlbatch"
)

const (
	sqlJWalletInsertIfNotExists = `
	insert into jwallets as c (
			address, info, create_time, notify_time
		)
		values (
			$1, $2::jsonb, now(), null
		)
	on conflict (address) do
		update set
			info = $2::jsonb
`

	sqlJWalletFind = `
	select
		address, info, create_time, notify_time
	from jwallets
	where address = $1
`

	sqlJWalletFindAllNotNotified = `
	select
		address, info, create_time, notify_time
	from jwallets
	where notify_time is null
`

	sqlJWalletRemove = `
	delete from jwallet where address = $1
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
		&r.Address, &infoJson, &r.CreateTime, &r.NotifyTime,
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
		&r.Address, &infoJson, &r.CreateTime, &r.NotifyTime,
	)
	list := memo.([]domain.JettonWallet)
	list = append(list, r)
	return list, err
}

func (repo *JettonWalletRepository) InsertIfNotExists(address string, info []domain.RelatedTransactionInfo) (*domain.JettonWallet, error) {

	infoJson, _ := json.Marshal(info)
	results, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
		{
			Query: sqlJWalletInsertIfNotExists,
			Args: []interface{}{
				address, infoJson,
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

// func (repo *BlockRepository) SetRejected(billId uuid.UUID, reference string) error {
// 	_, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
// 		{
// 			Query:  sqlBlockSetRejected,
// 			Args:   []interface{}{billId, reference},
// 			Affect: 1,
// 		},
// 	})
// 	return err
// }

// func (repo *BlockRepository) SetVerified(billId uuid.UUID, reference string) error {
// 	_, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
// 		{
// 			Query:  sqlBlockSetVerified,
// 			Args:   []interface{}{billId, reference},
// 			Affect: 1,
// 		},
// 	})

// 	return err
// }

// func (repo *BlockRepository) SetReversing(billId uuid.UUID, reference string) error {
// 	_, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
// 		{
// 			Query:  sqlBlockSetReversing,
// 			Args:   []interface{}{billId, reference},
// 			Affect: 1,
// 		},
// 	})

// 	return err
// }

// func (repo *BlockRepository) SetReversed(billId uuid.UUID, reference string) error {
// 	_, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
// 		{
// 			Query:  sqlBlockSetReversed,
// 			Args:   []interface{}{billId, reference},
// 			Affect: 1,
// 		},
// 	})

// 	return err
// }

// func (repo *BlockRepository) Reset(billId uuid.UUID, reference string) error {
// 	_, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
// 		{
// 			Query:  sqlBlockReset,
// 			Args:   []interface{}{billId, reference},
// 			Affect: 1,
// 		},
// 	})
// 	return err
// }

// func (repo *BlockRepository) Remove(billId uuid.UUID, reference string) error {
// 	_, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
// 		{
// 			Query:  sqlBlockRemove,
// 			Args:   []interface{}{billId, reference},
// 			Affect: 1,
// 		},
// 	})
// 	return err
// }
