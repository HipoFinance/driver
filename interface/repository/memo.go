package repository

import (
	"driver/domain"

	"github.com/behrang/sqlbatch"
)

const (
	sqlMemoUpsert = `
	insert into memos as c (
			key, memo
		)
		values (
			$1, $2::jsonb
		)
	on conflict (key) do
		update set
			memo = $2::jsonb
`

	sqlMemoFind = `
	select
		key, memo
	from memos
	where key = $1
`
)

type MemoRepository struct {
	batchHandler BatchHandler
}

func NewMemoRepository(db BatchHandler) *MemoRepository {
	return &MemoRepository{batchHandler: db}
}

func readMemo(scan func(...interface{}) error) (interface{}, error) {
	r := domain.Memo{}
	var jstr []byte
	err := scan(
		&r.Key, &jstr,
	)
	if err != nil {
		return &r, err
	}
	r.Memo = string(jstr)
	return &r, nil
}

func readAllMemos(all interface{}, scan func(...interface{}) error) (interface{}, error) {
	r := domain.Memo{}
	var jstr []byte
	err := scan(
		&r.Key, &jstr,
	)
	if err == nil {
		r.Memo = string(jstr)
	}

	list := all.([]domain.Memo)
	list = append(list, r)
	return list, err
}

func (repo *MemoRepository) Upsert(key string, memo domain.Memorable) (*domain.Memo, error) {

	jstr := memo.ToJson()
	results, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
		{
			Query: sqlMemoUpsert,
			Args: []interface{}{
				key, jstr,
			},
			Affect: 1,
		},
		{
			Query:   sqlJWalletFind,
			Args:    []interface{}{key},
			ReadOne: readMemo,
		},
	})

	result, _ := results[1].(*domain.Memo)
	return result, err
}

func (repo *MemoRepository) Find(key string) (*domain.Memo, error) {
	results, err := repo.batchHandler.Batch(&BatchOptionNormal, []sqlbatch.Command{
		{
			Query:   sqlMemoFind,
			Args:    []interface{}{key},
			ReadOne: readMemo,
		},
	})
	result, _ := results[0].(*domain.Memo)
	return result, err
}
