package stenography_service

import (
	"errors"
	"mime/multipart"

	"github.com/magnacartaam/chain-solutions/services/go-api/internal/stenography"
)

type AlgoService struct {
	algo stenography.StegoAlgorithm
}

func NewAlgoService() *AlgoService {
	return &AlgoService{
		algo: stenography.NewAlgorithm(),
	}
}

func (s *AlgoService) HideData(fileHeader *multipart.FileHeader, message string) ([]byte, error) {
	if message == "" {
		return nil, errors.New("message cannot be empty")
	}

	srcFile, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer srcFile.Close()

	resultBytes, err := s.algo.Hide(srcFile, []byte(message))
	if err != nil {
		return nil, err
	}

	return resultBytes, nil
}

func (s *AlgoService) ExtractData(fileHeader *multipart.FileHeader) (string, error) {
	srcFile, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer srcFile.Close()

	msgBytes, err := s.algo.Extract(srcFile)
	if err != nil {
		return "", err
	}

	return string(msgBytes), nil
}
