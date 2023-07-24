package usecase

import (
	"driver/domain"
	"driver/interface/repository"
)

const (
	ExtractionMemoKey = "extraction"
)

type MemoInteractor struct {
	memoRepository *repository.MemoRepository
}

func NewMemoInteractor(memoRepository *repository.MemoRepository) *MemoInteractor {
	interactor := &MemoInteractor{
		memoRepository: memoRepository,
	}
	return interactor
}

func (interactor *MemoInteractor) GetLatestProcessedHash() (string, error) {
	memo, err := interactor.memoRepository.Find(ExtractionMemoKey)
	if err != nil || memo == nil {
		return "", err
	}

	var extractionMemo domain.ExtractionMemo
	extractionMemo.FromJson(memo.Memo)
	return extractionMemo.LatestProcessedHash, nil
}

func (interactor *MemoInteractor) SetLatestProcessedHash(hash string) error {
	memo, err := interactor.memoRepository.Find(ExtractionMemoKey)
	if err != nil {
		return err
	}

	var extractionMemo domain.ExtractionMemo
	if memo != nil {
		extractionMemo.FromJson(memo.Memo)
	}
	extractionMemo.LatestProcessedHash = hash
	interactor.memoRepository.Upsert(ExtractionMemoKey, &extractionMemo)
	return nil
}
